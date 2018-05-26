package main

import (
	"fmt"
)

func main() {
	fmt.Println("Parsing the specified flags ...")
	InitFlags()

	fmt.Println("Loading the provided json hosts file from (" + *HOSTS_FILE + ") ...")
	if err := InitHostsList(); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Watching the hosts file for any change to hot reload it ...")
	go func() {
		WatchChanges()
	}()

	fmt.Println("Running the HTTP server on address (" + *HTTP_ADDR + ") ...")
	fmt.Println("Running the HTTPS (HTTP/2) server on address (" + *HTTPS_ADDR + ") ...")
	fmt.Println(InitServer())
}
