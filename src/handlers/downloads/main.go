/*
This package provides functionality for downloading files from the internet.
It defines a DownloadTask struct that contains information about the file to be downloaded, such as its URL, expected size, save path, and callback functions for completion and error handling.
The package also maintains a queue of download tasks, allowing for asynchronous processing of downloads.
*/
package downloads

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

/*
Maintains a queue of download tasks. The DownloadFile function adds a new task to the queue, and a separate goroutine can be used to process the queue and perform the actual downloads.
*/
var downloadQueue = make(chan DownloadTask, 100)

func DownloadFile(url string) {
	task := DownloadTask{
		URL: url,
	}

	downloadQueue <- task
}
