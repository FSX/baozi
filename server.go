package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

import (
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"golang.org/x/crypto/acme/autocert"
)

func RunServer(cfg *Config) error {
	m := autocert.Manager{
		Cache:  autocert.DirCache(cfg.CertPath),
		Prompt: autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			if _, ok := cfg.Hosts[host]; ok {
				return nil
			}
			return errors.New("Unkown host(" + host + ")")
		},
	}
	s := &http.Server{
		Addr:      cfg.HttpsAddr,
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		Handler:   ServeHTTP(cfg),
	}
	errchan := make(chan error)

	go (func(err chan error) {
		handler := m.HTTPHandler(ServeHTTP(cfg))
		// TODO
		// if *AUTOREDIRECT {
		// 	handler = m.HTTPHandler(nil)
		// }
		err <- http.ListenAndServe(cfg.HttpAddr, handler)
	})(errchan)

	go (func(err chan error) {
		err <- s.ListenAndServeTLS("", "")
	})(errchan)

	// FIXME: This only gets the first error that's in the channel.
	// In main.go there's no code that checks if there are errors.
	// Need to make sure both servers exit when one errors out.
	return <-errchan
}

func ServeHTTP(cfg *Config) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if upstreams, ok := cfg.Hosts[req.Host]; ok {
			forwarder, _ := forward.New(forward.PassHostHeader(true)) // TODO: Handle errors.
			loadbalancer, _ := roundrobin.New(forwarder)              // TODO: Handle errors.

			for _, upstream := range upstreams {
				if url, err := url.Parse(upstream); err == nil {
					loadbalancer.UpsertServer(url)
				} else {
					fmt.Println(err.Error())
				}
			}

			if cfg.Hsts != "" {
				res.Header().Set("Strict-Transport-Security", cfg.Hsts)
			}

			loadbalancer.ServeHTTP(res, req)
		} else {
			http.Error(res, "upstream server not found", http.StatusNotImplemented)
		}
	})
}
