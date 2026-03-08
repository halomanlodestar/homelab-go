package streaming

import (
	"fmt"
	"os"
)

func GetFileBuffer(path string, start uint) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0444);
	var bytes []byte

	if err != nil {
		fmt.Print("Unable to open the file specificed", err);
		return nil, err;
	}

	n, err := file.ReadAt(bytes, int64(start)); _ = n;

	if err != nil {
		return nil, err
	}

	return bytes, err
}