package store

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

// Store represents a content-addressable filesystem-based store
type Store interface {
	// ReadObj reads and decompresses the encoded object from the object store and write the encoded object to w.
	ReadObject([]byte) (common.EncodedObject, error)
	// WriteObj compresses and writes the encoded object data from r.
	WriteObject(common.EncodedObject) error
}

// Init initialise a store in the current directory.
// After the initialisation the .git directory, which is the store directory, will be created.
func Init() (*DefaultStore, error) {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed creating directory %s: %w", dir, err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return nil, fmt.Errorf("failed writing file: %w\n", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get cwd: %w", err)
	}
	return &DefaultStore{
		rootDir: path.Join(cwd, ".git"),
	}, nil
}

type DefaultStore struct {
	rootDir string
}

func NewDefaultStore() (*DefaultStore, error) {
	return &DefaultStore{
		rootDir: "./.git",
	}, nil
}

func (s *DefaultStore) ReadObject(checksum []byte) (common.EncodedObject, error) {
	file, err := os.Open(path.Join(s.rootDir, fmt.Sprintf("objects/%x/%x", checksum[:1], checksum[1:])))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	zlibR, err := zlib.NewReader(file)
	if err != nil {
		return nil, err
	}
	zlibR.Close()
	encodedObject := common.NewGenericEncodedObject()
	_, err = io.Copy(encodedObject, zlibR)
	if err != nil {
		return nil, err
	}
	return encodedObject, nil
}

func (s *DefaultStore) WriteObject(encodedObject common.EncodedObject) error {
	checksum, err := encodedObject.Hash()
	if err != nil {
		return fmt.Errorf("failed to hash object: %w", err)
	}
	err = os.Mkdir(s.getPath("objects", fmt.Sprintf("%x", checksum[:1])), 0744)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create object directory: %v", err)
	}
	file, err := os.OpenFile(fmt.Sprintf(".git/objects/%x/%x", checksum[:1], checksum[1:]), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		// The object already exist in the store
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("failed to create object file: %w", err)
	}
	defer file.Close()
	zlibW := zlib.NewWriter(file)
	defer zlibW.Close()
	if _, err := io.Copy(zlibW, encodedObject.NewReader()); err != nil {
		return err
	}
	return nil
}

// getPath joins the p into a single path and prefix it with store root.
func (s *DefaultStore) getPath(p ...string) string {
	return path.Join(s.rootDir, path.Join(p...))
}
