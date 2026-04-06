package main

import (
	"fmt"
	"homelab/internals/handlers/downloads"
	"homelab/internals/handlers/streaming"
	"net/http"
)

func main() {

	PORT := ":4080"

	http.HandleFunc("/", streaming.ListFiles)
	http.HandleFunc("/file", streaming.SendChunk)
	http.HandleFunc("/try", downloads.TestDownload)
	http.HandleFunc("/download", downloads.DownloadFileHandler)

	fmt.Println("Server is running on", PORT)

	http.ListenAndServe(PORT, nil)
}
