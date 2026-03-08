package main

import (
	"fmt"
	"homelab/src/handlers/streaming"
	"net/http"
)

func main() {

	http.HandleFunc("/", streaming.ListFiles)
	http.HandleFunc("/file", streaming.SendChunk)

	fmt.Println("Listening at 4080")
	http.ListenAndServe(":4080", nil);
}