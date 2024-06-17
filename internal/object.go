package common

import (
	"fmt"
	"io"
	"strconv"

	"github.com/codecrafters-io/git-starter-go/internal/util"
)

// This file contains a common set of object constants and functions.

// ObjectType represents a set of all possible type of both deltified and normal objects.
type ObjectType uint8

const (
	OBJ_COMMIT ObjectType = iota + 1
	OBJ_TREE
	OBJ_BLOB
	OBJ_TAG
	RESERVED
	OBJ_OFS_DELTA
	OBJ_REF_DELTA
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
		return fmt.Sprint("delta")
	}
}

func ParseObjectType(s string) ObjectType {
	switch s {
	case "commit":
		return OBJ_COMMIT
	case "tree":
		return OBJ_TREE
	case "blob":
		return OBJ_BLOB
	case "tag":
		return OBJ_TAG
	default:
		panic(fmt.Sprintf("failed to parse: %s", s))
	}
}

type ObjectHeader struct {
	Size int
	Type ObjectType
}

type Object interface {
	Hash() ([]byte, error)
	Type() ObjectType
	Size() int
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Encode(w io.Writer) error
}

func EncodeObjectHeader(header ObjectHeader, w io.Writer) error {
	_, err := w.Write([]byte(fmt.Sprintf("%s %d\x00", header.Type, header.Size)))
	return err
}

// EncodeObject prepends object header to the object content.
func EncodeObject(objectType ObjectType, content []byte, w io.Writer) error {
	if _, err := w.Write([]byte(fmt.Sprintf("%s %d\x00", objectType, len(content)))); err != nil {
		return err
	}
	if _, err := w.Write(content); err != nil {
		return err
	}
	return nil
}

func DecodeObjectHeader(r io.Reader) (*ObjectHeader, error) {
	objectTypeBytes, err := util.ReadBefore(r, []byte(" "))
	if err != nil {
		return nil, fmt.Errorf("failed to read object type: %w", err)
	}
	objectType := ParseObjectType(string(objectTypeBytes))

	sizeBytes, err := util.ReadBefore(r, []byte("\x00"))
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(string(sizeBytes))
	if err != nil {
		return nil, err
	}
	return &ObjectHeader{
		Type: objectType,
		Size: size,
	}, nil

}
