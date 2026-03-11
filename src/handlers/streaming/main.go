package streaming

import (
	"encoding/json"
	"fmt"
	"homelab/src/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FileDetails struct {
	Name string `json:"name"`;
	IsDir bool `json:"is_dir"`
	Size int64 `json:"size"`
	Type string `json:"type"`
}

var VideoHeaders = map[string]string {
	".mp4": "video/mp4",
	".webm": "video/webm",
	".avi": "video/x-msvide",
	".mkv": "video/x-matroska",
	".mov": "video/quicktime",
}

const CHUNK_SIZE int64 = 128_000;

func SendChunk(writer http.ResponseWriter, request *http.Request) {

	var encoder = json.NewEncoder(writer)

	queryParams := request.URL.Query()

	path := utils.If(queryParams.Get("path") != "", queryParams.Get("path"), ".");

	file, err := os.OpenFile(path, os.O_RDONLY, 0444)

	if err != nil {
		encoder.Encode(err.Error())
		return
	}

	info, err := file.Stat()

	if err != nil {
		encoder.Encode(err.Error())
		return
	}

	metadata := GetFileMetadataFromInfo(info)

	if metadata.IsDir {
		writer.WriteHeader(400)
		fmt.Fprint(writer, metadata.Name, " is a directory")
		return
	}

	v := VideoHeaders[metadata.Type]
	
	if v == "" {
		writer.WriteHeader(400)
		fmt.Fprint(writer, metadata.Type, " isn't a valid video type")
		return
	}

	var start int64 = 0

	rangeHeader := request.Header.Get("Range")
	
	if rangeHeader != "" {
		ranges := rangeHeader[6:]
		startRange := strings.Split(ranges, "-")[0]
		start, err = strconv.ParseInt(startRange, 10, 64)

		if err != nil {
			encoder.Encode(err.Error())
			return
		}
	}

	totalContentSize := metadata.Size;
	end := min(totalContentSize, start + CHUNK_SIZE)

	writer.Header().Set("Content-Type", v)
	writer.Header().Set("Accept-Ranges", "bytes")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Keep-Alive", "timeout=5, max=100")
	writer.Header().Set(
		"Content-Range", 
		fmt.Sprintf("bytes %d-%d/%d", start, end - 1, totalContentSize),
	)

	var bytesToRead = end - start + 1;
	var bytes = make([]byte, bytesToRead)

	n, err := file.ReadAt(bytes, start); _ = n

	if err != nil {
		fmt.Println(err)
	}

	writer.WriteHeader(http.StatusPartialContent)
	writer.Write(bytes)
}

func GetFileMetadata(writer http.ResponseWriter, request *http.Request) {

	encoder := json.NewEncoder(writer)

	queryParams := request.URL.Query()

	path := utils.If(queryParams.Get("path") != "", queryParams.Get("path"), ".");

	file, err := os.OpenFile(path, os.O_RDONLY, 0444)

	if err != nil {
		encoder.Encode(err.Error())
		return
	}

	info, err := file.Stat()

	if err != nil {
		encoder.Encode(err.Error())
		return
	}

	fileMetaData := GetFileMetadataFromInfo(info)

	writer.Header().Set("Content-Type", "application/json")

	encoder.Encode(fileMetaData)
}

func GetFileMetadataFromInfo(info (os.FileInfo)) FileDetails {
	name := info.Name()
	ext := utils.If(info.IsDir(), "", filepath.Ext(name))

	return FileDetails {
			Name: name,
			Size: info.Size(),
			IsDir: info.IsDir(),
			Type: ext,
	}
}

func ListFiles(writer http.ResponseWriter, request *http.Request) {

	encoder := json.NewEncoder(writer)

	queryParams := request.URL.Query()

	path := utils.If(queryParams.Get("path") != "", queryParams.Get("path"), ".");

	dirEntires, err := os.ReadDir(path)

	writer.Header().Set("Content-Type", "application/json")

	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Println(err)
		encoder.Encode(err.Error())
		return
	}

	writer.WriteHeader(http.StatusOK)

	var dirs []FileDetails;

	for _, value := range dirEntires {
		info, _ := value.Info()

		dirs = append(dirs, GetFileMetadataFromInfo(info))
	}

	encoder.Encode(dirs);
}