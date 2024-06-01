package store

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

func TestWriteObject(t *testing.T) {
	if os.Getenv("TEST_IO") == "" {
		t.Skip()
	}
	encodedObject := common.NewGenericEncodeObjectFromBuffer([]byte("blob 11\000hello world"))
	err := os.Chdir(t.TempDir())
	if err != nil {
		t.Fatalf("failed to change working directory to test temp dir: %v", err)
	}
	store, err := Init()
	if err != nil {
		t.Fatalf("failed to create new store: %v", err)
	}
	err = store.WriteObject(encodedObject)
	if err != nil {
		t.Fatalf("failed to write object: %v", err)
	}

	checksum, err := encodedObject.Hash()
	if err != nil {
		t.Fatalf("failed to has object: %v", err)
	}
	file, err := os.Open(path.Join(store.rootDir, "objects", fmt.Sprintf("%x", checksum[:1]), fmt.Sprintf("%x", checksum[1:])))
	if err != nil {
		t.Fatalf("cannot open object file: %v", err)
	}
	defer file.Close()
	zlibR, err := zlib.NewReader(file)
	if err != nil {
		t.Fatalf("cannot wrap file with zlib decoder: %v", err)
	}
	defer zlibR.Close()
	fileContent, err := io.ReadAll(zlibR)
	if err != nil {
		t.Fatalf("cannot read object file: %v", err)
	}
	if !bytes.Equal(encodedObject.Bytes(), fileContent) {
		t.Fatalf("expected(file content): %s, actual(file content): %s", encodedObject.Bytes(), fileContent)
	}

}
