package main

import (
	"os"
)

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func empty(v string) bool {
	return len(v) == 0
}
