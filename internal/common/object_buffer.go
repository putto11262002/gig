package common

import (
	"crypto/sha1"
	"fmt"
	"io"
)

type ObjectType uint8

const (
	OBJ_COMMIT ObjectType = iota + 1
	OBJ_TREE
	OBJ_BLOB
	OBJ_TAG
)

func (ot ObjectType) String() string {
	switch ot {
	case OBJ_COMMIT:
		return "commit"
	case OBJ_TREE:
		return "tree"
	case OBJ_BLOB:
		return "blob"
	case OBJ_TAG:
		return "tag"
	default:
		return string(ot)
	}
}

type ObjectBuffer struct {
	objectType ObjectType
	content    []byte
}

func NewObjectBuffer(objectType ObjectType, content []byte) (*ObjectBuffer, error) {
	return &ObjectBuffer{
		objectType: objectType,
		content:    content,
	}, nil
}

func (object *ObjectBuffer) Hash() (Checksum, error) {
	hashFn := sha1.New()
	if err := object.Encode(hashFn); err != nil {
		return nil, err
	}
	return hashFn.Sum(nil), nil
}

func (object *ObjectBuffer) Encode(w io.Writer) error {

	if _, err := w.Write([]byte(fmt.Sprintf("%s %d\x00", object.objectType, len(object.content)))); err != nil {
		return err
	}
	if _, err := w.Write(object.content); err != nil {
		return err
	}
	return nil
}

func (object *ObjectBuffer) Content() []byte {
	return object.content
}
