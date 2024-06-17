package store

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	DIR           = ".git"
	OBJECT_PREFIX = "objects"
	PACK_PREFIX   = "objects/pack"
	REF_PREFIX    = "refs"
	HEAD          = "refs/HEAD"
)

var (
	ErrStoreNotExist = errors.New("store not found")
)

// FSStore represents a content-addressable database
type Store interface {
	NewPackReader(checksum string) (ReadOnlyFile, error)
	NewPackWriter(checksum string) (WriteReadFile, error)
	NewPackIndexReader(checksum string) (ReadOnlyFile, error)
	NewPackIndexWriter(checkum string) (WriteReadFile, error)
	ListPacks() ([]string, error)
	ListPackIndices() ([]string, error)
	ObjectReader(string) (ReadOnlyFile, error)
	ObjectWriter(string) (WriteReadFile, error)
}

type FSStore struct {
	rootDir string
}

type File interface {
	io.ReadSeeker
	io.ReaderAt
	io.Writer
	io.ReadCloser
	Sync() error
	Rename(string) error
}

type file struct {
	*os.File
}

func (f *file) Sync() error {
	return f.Sync()
}

func (f *file) Rename(path string) error {
	return os.Rename(f.Name(), path)
}

// New() returns a store object representing an existing store on disk.
// The existing store is discovered by walking up the file system tree until a [DIR] is found.
func New() (*FSStore, error) {
	rootDir, err := findStoreDir()
	if err != nil {
		return nil, err
	}
	return &FSStore{
		rootDir: rootDir,
	}, nil
}

func Init() (*FSStore, error) {
	if err := os.Mkdir(DIR, 0755); err != nil {
		return nil, err
	}

	requiredSubDirs := []string{"objects", "refs", "objects/pack"}
	for _, subDir := range requiredSubDirs {
		path := path.Join(DIR, subDir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed creating directory %s: %w", path, err)
		}
	}

	requiredFiles := map[string][]byte{"HEAD": []byte("ref: refs/heads/main\n")}
	for name, content := range requiredFiles {
		fullPath := path.Join(DIR, name)
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return nil, fmt.Errorf("failed writing file: %w\n", err)
		}
	}

	// cwd, err := os.Getwd()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get cwd: %w", err)
	// }

	return &FSStore{
		rootDir: path.Join("./", DIR),
	}, nil

}

func findStoreDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	numVisitDir := len(strings.Split(cwd, "/"))
	var relativePath strings.Builder
	for i := 0; i < numVisitDir; i++ {
		p := path.Join(relativePath.String(), DIR)
		stat, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				relativePath.WriteString("../")
				continue
			}
			return "", err
		}

		if stat.IsDir() {
			abs, err := filepath.Abs(p)
			if err != nil {
				return "", err
			}
			return abs, nil
		}

		relativePath.WriteString("../")
	}
	return "", ErrStoreNotExist
}
