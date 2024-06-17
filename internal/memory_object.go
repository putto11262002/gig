package common

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
)

// MemoryObject can be any object that is loaded into memory.
type MemoryObject struct {
	objectType ObjectType
	content    *bytes.Buffer
}

func NewObjectBuffer(objectType ObjectType, content []byte) *MemoryObject {
	return &MemoryObject{
		objectType: objectType,
		content:    bytes.NewBuffer(content),
	}
}

func (object *MemoryObject) Hash() ([]byte, error) {
	hashFn := sha1.New()
	if err := object.Encode(hashFn); err != nil {
		return nil, err
	}
	return hashFn.Sum(nil), nil
}

func (object *MemoryObject) Encode(w io.Writer) error {
	if _, err := w.Write([]byte(fmt.Sprintf("%s %d\x00", object.objectType, object.Size()))); err != nil {
		return err
	}
	if _, err := w.Write(object.Content()); err != nil {
		return err
	}
	return nil
}

func (object *MemoryObject) Write(p []byte) (int, error) {
	return object.content.Write(p)
}

func (object *MemoryObject) Read(p []byte) (int, error) {
	return object.content.Read(p)
}

func (object *MemoryObject) Size() int {
	return object.content.Len()
}

func (object *MemoryObject) Content() []byte {
	return object.content.Bytes()
}

func (object *MemoryObject) Type() ObjectType {
	return object.objectType
}

func DecodeMemoryObject(r io.Reader) (*MemoryObject, error) {
	decompR, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer decompR.Close()
	header, err := DecodeObjectHeader(decompR)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	var content bytes.Buffer

	if _, err := io.Copy(&content, decompR); err != nil {
		return nil, err
	}

	return &MemoryObject{
		objectType: header.Type,
		content:    &content,
	}, nil

}
