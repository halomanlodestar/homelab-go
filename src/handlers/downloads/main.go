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
	"mime"
	"net/http"
	"os"
	"time"
)

type DownloadTask struct {

	/*
	The URL of the file to be downloaded. This is the only required field.
	*/
	URL string

	/*
	The expected size of the file in bytes. This can be used to show progress.
	*/
	expectedSize int64

	/*
	The path where the file should be saved. If not provided, the file will be saved in the current directory with its original name.
	*/
	savePath string
	
	/*
	A callback function that will be called when the download is complete. The function will receive the path to the downloaded file as an argument.
	*/
	onComplete func(string)

	/*
	A callback function that will be called if there is an error during the download. The function will receive the error as an argument.
	*/
	onError func(error)

	/*
	The number of bytes that have been downloaded so far. This can be used to show progress.
	*/
	downloadedBytes int64
}

type DownloadManager struct {
	queue []DownloadTask
	Destination string
}

func TestDownload(w http.ResponseWriter, r *http.Request) {

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	manager := DownloadManager {
		Destination: cwd + "/downloads",
	}
	
	manager.DownloadFile("http://localhost:4080/file?path=files/file1.mp4")
}

func isUrlValid(_ string) bool {
	return true
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

	manager := DownloadManager {
		Destination: cwd + "/downloads",
	}
	
	manager.DownloadFile(url)
}

func DownloadFromStream(url string, writer *bufio.Writer) error {

	downloading := true

	client := http.Client {}
	headers := http.Header {}
	
	for downloading {
		req, err := http.NewRequest(http.MethodGet, url, nil); _ = req
		req.Header = headers

		if err != nil {
			downloading = false
		}

		res, err := client.Do(req)

		if err != nil {
			downloading = false
		}
		
		// current range headers
		contentRange := res.Header.Get("Content-Range")
		var start, end, totalContentSize int64
		fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &end, &totalContentSize)
		chunkSize := end - start + 1

		// write data to the file
		bodyReader := res.Body;
		
		data := make([]byte, chunkSize)
		n, err := bodyReader.Read(data)
		nn, err := writer.Write(data[:n])

		if err != nil {
			fmt.Println(nn, err)
		}

		if end + 1 == totalContentSize || end == totalContentSize {
			downloading = false;

			fmt.Println("End of the content reached")

			continue
		}

		headers.Set("Range", fmt.Sprintf("bytes=%d-", end + 1))
	}

	return nil
}

func DownloadFromBulk(url string, writer *bufio.Writer) error {
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	contentLength := res.ContentLength

	bodyReader := res.Body
	data := make([]byte, contentLength)
	n, err := bodyReader.Read(data)
	writer.Write(data[:n])

	return nil
}

func (manager *DownloadManager) DownloadFile(url string) error {

	if !isUrlValid(url) {
		return errors.New("Invalid URL")
	}

	task := DownloadTask {
		URL: url,
	}
	
	manager.queue = append(manager.queue, task)

	// first get headers
	req, _ := http.Get(url)

	contentDisposition := req.Header.Get("Content-Disposition")
	contentType := req.Header.Get("Content-Type")

	mediaType, metadata, _ := mime.ParseMediaType(contentDisposition); _ = mediaType
	filename := metadata["filename"]
	exts, err := mime.ExtensionsByType(contentType)

	if filename == "" {
		filename = fmt.Sprintf("%d", time.Now().Second())
	}

	if len(exts) == 0 || err != nil {
		exts = append(exts, "")
	}

	filePath := manager.Destination + "/" + filename + exts[0]

	file, err := os.Create(filePath)

	if err != nil {
		fmt.Println(err)
		return err
	}

	writer := bufio.NewWriter(file); _ = writer

	switch req.StatusCode {
	case 206:
		DownloadFromStream(url, writer)
		
	case 200:
		DownloadFromBulk(url, writer)
	
	case 403:
		fmt.Println("Request blocked by the server")
		os.Remove(filePath)
	}

	return nil
}
