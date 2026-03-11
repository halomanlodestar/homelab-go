package main

import (
	"fmt"
	"homelab/src/handlers/downloads"
	"homelab/src/handlers/streaming"
	"log"
	"net/http"
	"os"
)

func main() {
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	manager := downloads.DownloadManager {
		Destination: cwd,
	}

	manager.DownloadFile("")

	http.HandleFunc("/", streaming.ListFiles)
	http.HandleFunc("/file", streaming.SendChunk)

	fmt.Println("Listening at 4080")
	http.ListenAndServe(":4080", nil);
}