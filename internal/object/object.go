package object

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

// GetObjectTypeFromFileMode `
func GetObjectTypeFromFileMode(mode []byte) []byte {
	if bytes.Equal(mode, []byte("040000")) {
		return []byte("tree")
	} else {
		return []byte("blob")
	}
}

// readObject reads and decompress the object that matches the checksum.
// The raw object is returned as a slice of bytes
func ReadObject(checksum []byte) ([]byte, error) {
	file, err := os.Open(fmt.Sprintf(".git/objects/%x/%x", checksum[:1], checksum[1:]))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	zlibR, err := zlib.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer zlibR.Close()
	buffer, err := io.ReadAll(zlibR)
	return buffer, nil
}
