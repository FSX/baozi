package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	HttpAddr  string
	HttpsAddr string
	CertPath  string
	Hsts      string
	Verbose   bool
	Hosts     map[string][]string
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

	return cfg
}

func loadjson(filepath string) *Config {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &Config{Hosts: make(map[string][]string)}

	if err := json.Unmarshal(content, cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}
