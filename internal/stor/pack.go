package store

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type PackFile struct {
	file
}

func (f *PackFile) Rename(checksome string) error {
	return f.file.Rename(path.Join(path.Dir(f.Name()), fmt.Sprintf("pack-%s.pack", checksome)))
}

type PackIndexFile struct {
	file
}

func (f *PackIndexFile) Rename(checksome string) error {
	return f.file.Rename(fmt.Sprintf("pack-%s.idx", checksome))
}

func (store *FSStore) NewPackWriter(key string) (WriteReadFile, error) {
	if key == "" {
		key = fmt.Sprintf("%d", time.Now().Unix())
	}
	f, err := os.OpenFile(path.Join(store.rootDir, PACK_PREFIX, fmt.Sprintf("pack-%s.pack", key)), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &PackFile{
		file{f},
	}, nil
}

func (store *FSStore) NewPackReader(checksum string) (ReadOnlyFile, error) {
	f, err := os.OpenFile(path.Join(store.rootDir, PACK_PREFIX, fmt.Sprintf("pack-%s.pack", checksum)), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &fileReader[*PackFile]{
		file: &PackFile{file{f}},
	}, nil
}

func (store *FSStore) NewPackIndexWriter(key string) (WriteReadFile, error) {
	if key == "" {
		key = fmt.Sprintf("%d", time.Now().Unix())
	}
	f, err := os.OpenFile(path.Join(store.rootDir, PACK_PREFIX, fmt.Sprintf("pack-%s.idx", key)), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &PackIndexFile{
		file{f},
	}, nil
}

func (store *FSStore) NewPackIndexReader(checksum string) (ReadOnlyFile, error) {
	f, err := os.OpenFile(path.Join(store.rootDir, PACK_PREFIX, fmt.Sprintf("pack-%s.idx", checksum)), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &fileReader[*PackIndexFile]{
		file: &PackIndexFile{file{f}},
	}, nil
}

func (store *FSStore) ListPackIndices() ([]string, error) {
	entries, err := os.ReadDir(path.Join(store.rootDir, PACK_PREFIX))
	if err != nil {
		return nil, err
	}
	indexFiles := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".idx") {
			checksum := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), "pack-"), ".idx")
			indexFiles = append(indexFiles, checksum)
		}
	}
	return indexFiles, nil
}

func (store *FSStore) ListPacks() ([]string, error) {
	entries, err := os.ReadDir(path.Join(store.rootDir, PACK_PREFIX))
	if err != nil {
		return nil, err
	}
	packFiles := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pack") {
			checksum := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), "pack-"), ".pack")
			packFiles = append(packFiles, checksum)
		}
	}
	return packFiles, nil
}
