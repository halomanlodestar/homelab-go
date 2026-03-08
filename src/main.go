package main

import (
	"fmt"
	"homelab/src/handlers"
	"net/http"
)

func main() {

	http.HandleFunc("/", handlers.ListFiles)
	http.HandleFunc("/file", handlers.SendChunk)

	fmt.Println("Listening at 4080")
	http.ListenAndServe(":4080", nil);
}