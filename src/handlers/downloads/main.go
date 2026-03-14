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

func (manager *DownloadManager) DownloadFile(url string) error {

	// 1. check whether the url is valid or not
	if !isUrlValid(url) {
		return errors.New("Invalid URL")
	}

	task := DownloadTask {
		URL: url,
	}
	
	manager.queue = append(manager.queue, task)

	downloading := true

	client := http.Client {}
	headers := http.Header {}

	// first get headers
	req, _ := http.Get(url)

	if req.StatusCode != 206 {
			fmt.Println("Download doesn't support streams")
			downloading = false
	}

	// create a file
	contentDisposition := req.Header.Get("Content-Disposition")

	mediaType, metadata, _ := mime.ParseMediaType(contentDisposition); _ = mediaType
	filename := metadata["filename"]

	file, err := os.Create(manager.Destination + "/" + filename)

	if err != nil {
		fmt.Println(err)
		return err
	}

	writer := bufio.NewWriter(file); _ = writer

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
		responseHeader := res.Header
		contentRange := responseHeader.Get("Content-Range")
		var start, end, totalContentSize int64
		fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &end, &totalContentSize)
		chunkSize := end - start + 1
		fmt.Println(contentRange)

		// write data to the file
		bodyReader := res.Body; _ = bodyReader
		
		data := make([]byte, chunkSize)
		n, err := bodyReader.Read(data)
		nn, err := writer.Write(data[:n])

		if err != nil {
			fmt.Println(nn, err)
		}

		// fmt.Println(len(data), chunkSize)

		if end + 1 == totalContentSize || end == totalContentSize {
			downloading = false;

			fmt.Println("End of the content reached")

			continue
		}

		headers.Set("Range", fmt.Sprintf("bytes=%d-", end + 1))
	}

	return nil
}
