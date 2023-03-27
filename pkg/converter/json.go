package converter

import (
	"io"
	"os"
)

func JsonToBytes(file string) ([]byte, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	return io.ReadAll(jsonFile)
}
