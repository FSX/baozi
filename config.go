package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

type Config struct {
	HttpAddr   string          `json:"http_addr,omitempty"`
	HttpsAddr  string          `json:"http_addr,omitempty"`
	CertPath   string          `json:"cert_path,omitempty"`
	Hsts       string          `json:"hsts,omitempty"`
	Verbose    bool            `json:"verbose,omitempty"`
	AdminEmail string          `json:"admin_email,omitempty"`
	Rules      map[string]Rule `json:"rules,omitempty"`
}

type Rule struct {
	Upstream string            `json:"upstream,omitempty"`
	Redirect string            `json:"redirect,omitempty"`
	TLS      bool              `json:"tls,omitempty"`
	Paths    map[string]string `json:"paths,omitempty"`
}

func Loadconfig() *Config {
	var f = flag.String("f", "/etc/baozi.conf", "Specifies the configuration file.")
	var v = flag.Bool("v", false, "Produce more verbose output.")
	flag.Parse()

	if !fileExists(*f) {
		log.Fatalf("%s: does not exist", *f)
	}

	cfg := loadjson(*f)
	cfg.Verbose = *v

	if cfg.CertPath == "" {
		cfg.CertPath = "/etc/ssl/baozi"
	}
	if !fileExists(cfg.CertPath) {
		log.Fatalf("%s: does not exist", cfg.CertPath)
	}

	for k, v := range cfg.Rules {
		checkrule(k, v)
	}

	return cfg
}

func checkrule(host string, rule Rule) {
	if empty(rule.Upstream) && empty(rule.Redirect) {
		log.Fatalf("%s: either upstream or redirect must be set\n", host)
	}
	if rule.Redirect != "" && rule.TLS {
		log.Fatalf("%s: tls is ignored when redirect is set\n", host)
	}

	// TODO: Check host.
	// TODO: Check upstream.
	// TODO: Check redirect.
	// TODO: Check paths.
}

func loadjson(filepath string) *Config {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &Config{Rules: make(map[string]Rule)}

	if err := json.Unmarshal(content, cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}
