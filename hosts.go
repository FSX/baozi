package main

import (
	"encoding/json"
	"os"
)

var (
	HOSTS = make(map[string][]string)
)

// Load the hosts file into memory
func InitHostsList() error {
	f, e := os.OpenFile(*HOSTS_FILE, os.O_RDONLY|os.O_CREATE, 0666)
	if e != nil {
		return e
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		return err
	}
	if finfo.Size() == 0 {
		return nil
	}
	return json.NewDecoder(f).Decode(&HOSTS)
}
