/*
This package provides functionality for downloading files from the internet.
It defines a DownloadTask struct that contains information about the file to be downloaded, such as its URL, expected size, save path, and callback functions for completion and error handling.
The package also maintains a queue of download tasks, allowing for asynchronous processing of downloads.
*/
package downloads

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"mime"
	"net/http"
	"os"
	"time"
)

const (
	KiloByte = 1024
	MegaByte = KiloByte * 1024
	GigaByte = MegaByte * 1024
)

type DownloadTask struct {
	URL             string
	TotalSize       int64
	SavePath        string
	DownloadedBytes int64
	FileName        string
}

type DownloadManager struct {
	tasks       map[string]*DownloadTask
	Destination string
	updated     chan bool
}

func (m *DownloadManager) UpdateProgress() {}

func TestDownload(w http.ResponseWriter, r *http.Request) {

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	manager := DownloadManager{
		Destination: cwd + "/downloads",
	}

	manager.DownloadFile("http://localhost:4080/file?path=files/file1.mp4")
}

func isUrlValid(_ string) bool {
	return true
}

func (m *DownloadManager) PollDownloadProgress() {
	for {

		if u := <-m.updated; !u {
			continue
		}

		for _, v := range m.tasks {
			fmt.Printf("\r%s - %.0fMB/%.0fMB", v.FileName, math.Abs(convertToMb(v.DownloadedBytes)), convertToMb(v.TotalSize))
		}
	}
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	query := r.URL.Query()
	url := query.Get("url")

	manager := DownloadManager{
		Destination: cwd + "/downloads",
		tasks:       map[string]*DownloadTask{},
		updated:     make(chan bool),
	}

	go manager.PollDownloadProgress()

	manager.DownloadFile(url)
}

func (m *DownloadManager) DownloadFromStream(url string, writer *bufio.Writer) error {

	downloading := true

	client := http.Client{}
	headers := http.Header{}

	for downloading {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		_ = req
		req.Header = headers

		if err != nil {
			downloading = false
		}

		res, err := client.Do(req)

		if err != nil {
			downloading = false
		}

		contentRange := res.Header.Get("Content-Range")
		var start, end, totalContentSize int64
		fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &end, &totalContentSize)
		chunkSize := end - start + 1

		bodyReader := res.Body

		data := make([]byte, chunkSize)
		n, err := bodyReader.Read(data)
		nn, err := writer.Write(data[:n])

		if err != nil {
			fmt.Println(nn, err)
		}

		if end+1 == totalContentSize || end == totalContentSize {
			downloading = false

			fmt.Println("End of the content reached")

			continue
		}

		headers.Set("Range", fmt.Sprintf("bytes=%d-", end+1))

		m.tasks[url].DownloadedBytes -= int64(n)
		m.updated <- true
	}

	return nil
}

func convertToMb(bytes int64) float64 {
	return float64(bytes) / float64(MegaByte)
}

func (m *DownloadManager) DownloadFromBulk(url string, writer *bufio.Writer) error {
	res, err := http.Get(url)

	if err != nil {
		fmt.Println("Error making GET request:", err)
		return err
	}

	const CHUNK_SIZE = 30 * MegaByte

	bodyReader := res.Body
	data := make([]byte, CHUNK_SIZE)

	for m.tasks[url].DownloadedBytes < m.tasks[url].TotalSize {
		n, err := bodyReader.Read(data)

		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}

		_, err = writer.Write(data[:n])

		if err != nil {
			fmt.Println("Error writing to file:", err)
			return err
		}

		m.tasks[url].DownloadedBytes -= int64(n)
		m.updated <- true
	}

	return nil
}

func (manager *DownloadManager) DownloadFile(url string) error {

	if !isUrlValid(url) {
		return errors.New("Invalid URL")
	}

	req, _ := http.Get(url)

	contentDisposition := req.Header.Get("Content-Disposition")
	contentType := req.Header.Get("Content-Type")

	mediaType, metadata, _ := mime.ParseMediaType(contentDisposition)
	_ = mediaType
	filename := metadata["filename"]
	exts, err := mime.ExtensionsByType(contentType)

	if err != nil {
		fmt.Println("Error parsing content type:", err)
	}

	if filename == "" {
		filename = fmt.Sprintf("%d", time.Now().Second())
	}

	if len(exts) == 0 || err != nil {
		exts = append(exts, "")
	}

	filePath := manager.Destination + "/" + filename + exts[0]

	manager.tasks[url] = &DownloadTask{
		URL:             url,
		TotalSize:       req.ContentLength,
		SavePath:        filePath,
		DownloadedBytes: 0,
		FileName:        filename,
	}

	file, err := os.Create(filePath)

	if err != nil {
		fmt.Println(err)
		return err
	}

	writer := bufio.NewWriter(file)
	_ = writer

	switch req.StatusCode {
	case 206:
		fmt.Println("Downloading from stream")
		go manager.DownloadFromStream(url, writer)

	case 200:
		fmt.Println("Downloading from bulk")
		go manager.DownloadFromBulk(url, writer)

	case 403:
		fmt.Println("Request blocked by the server")
		os.Remove(filePath)
	}

	return nil
}
