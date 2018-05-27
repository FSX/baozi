package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

func RunServer(cfg *Config) error {
	manager := autocert.Manager{
		Cache:  autocert.DirCache(cfg.CertPath),
		Prompt: autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			if _, ok := cfg.Rules[host]; ok {
				return nil
			}
			return errors.New("Unkown host(" + host + ")")
		},
		Email: cfg.AdminEmail,
	}
	server := &http.Server{
		Addr:      cfg.HttpsAddr, // Doesn't work? autocert still listens on 443 anyway.
		TLSConfig: &tls.Config{GetCertificate: manager.GetCertificate},
		Handler:   ServeHTTP(cfg),
	}
	errchan := make(chan error)

	go (func(err chan error) {
		err <- http.ListenAndServe(cfg.HttpAddr, manager.HTTPHandler(nil))
	})(errchan)

	go (func(err chan error) {
		err <- server.ListenAndServeTLS("", "")
	})(errchan)

	// FIXME: This only gets the first error that's in the channel.
	// In main.go there's no code that checks if there are errors.
	// Need to make sure both servers exit when one errors out.
	return <-errchan
}

func ServeHTTP(cfg *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rule, ok := cfg.Rules[r.Host]
		if !ok {
			upstreamNotAvailable(w)
			return
		}

		if !empty(rule.Redirect) {
			if target, ok := cfg.Rules[rule.Redirect]; ok {
				http.Redirect(w, r, makeUrl(rule.Redirect, target.TLS), http.StatusMovedPermanently)
			} else {
				upstreamNotAvailable(w)
				return
			}
		}

		if r.TLS == nil && rule.TLS {
			http.Redirect(w, r, makeUrl(rule.Upstream, rule.TLS), http.StatusMovedPermanently)
			return
		}

		upstream := rule.Upstream

		for k, v := range rule.Paths {
			if strings.HasPrefix(r.URL.Path, k) {
				upstream = v
				break
			}
		}

		origin, _ := url.Parse(upstream)
		director := func(r *http.Request) {
			r.Header.Add("X-Forwarded-Host", r.Host)
			r.Header.Add("X-Origin-Host", origin.Host)
			r.Header.Set("Connection", "close")
			r.URL.Scheme = "http"
			r.URL.Host = origin.Host
		}
		proxy := &httputil.ReverseProxy{Director: director}

		if cfg.Hsts != "" {
			w.Header().Set("Strict-Transport-Security", cfg.Hsts)
		}

		proxy.ServeHTTP(w, r)
	})
}

func makeUrl(host string, tls bool) string {
	if tls {
		return fmt.Sprintf("https://%s/", host)
	} else {
		return fmt.Sprintf("http://%s/", host)
	}
}

func upstreamNotAvailable(w http.ResponseWriter) {
	http.Error(w, "upstream server not available", http.StatusNotImplemented)
}
