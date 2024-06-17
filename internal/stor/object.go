package store

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path"
)

type ObjectFile struct {
	file
}

func (f *ObjectFile) Rename(checksum string) error {
	objectPath := path.Join(path.Dir(path.Dir(f.Name())), checksum[:2], checksum[2:])
	if err := os.MkdirAll(path.Dir(objectPath), 0755); err != nil && !os.IsExist(err) {
		return err
	}
	return f.file.Rename(objectPath)
}

func (store *FSStore) ObjectReader(checksum string) (ReadOnlyFile, error) {
	_file, err := os.Open(path.Join(store.rootDir, OBJECT_PREFIX, checksum[:2], checksum[2:]))
	if err != nil {
		return nil, err
	}
	return &fileReader[*ObjectFile]{
		&ObjectFile{file{_file}},
	}, nil
}

func (store *FSStore) ObjectWriter(checksum string) (WriteReadFile, error) {
	p := path.Join(store.rootDir, OBJECT_PREFIX)
	if checksum == "" {
		b := make([]byte, 10)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		p = path.Join(p, "temp", hex.EncodeToString(b))
	} else {
		p = path.Join(p, checksum[:2], checksum[2:])
	}

	if err := os.MkdirAll(path.Dir(p), 0755); err != nil {
		return nil, err
	}

	_file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &ObjectFile{file{_file}}, nil
}
