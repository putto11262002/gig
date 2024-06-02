package common

import (
	"bytes"
	"io"
)

// EncodedObject represents the encoded version of any object
type EncodedObject interface {
	io.Writer
	WriteString(s string) (int, error)
	NewReader() io.Reader
	Bytes() []byte
	Hash() ([]byte, error)
	// Size returns size of the object content in btyes
	Size() int
}

type GenericEncodeObject struct {
	content bytes.Buffer
	header  bytes.Buffer
	hash    []byte
}

func NewGenericEncodedObject() *GenericEncodeObject {
	return &GenericEncodeObject{
		content: bytes.Buffer{},
	}
}

func NewGenericEncodeObjectFromBuffer(buffer []byte) *GenericEncodeObject {
	return &GenericEncodeObject{
		content: *bytes.NewBuffer(buffer),
	}
}

// Size returns teh size of the encoded object content
func (eo *GenericEncodeObject) Size() int {
	return eo.content.Len()
}
func (eo *GenericEncodeObject) NewReader() io.Reader {
	return bytes.NewReader(eo.content.Bytes())
}

func (eo *GenericEncodeObject) Write(p []byte) (int, error) {
	if len(p) > 0 {
		eo.hash = nil
	}
	return eo.content.Write(p)
}

func (eo *GenericEncodeObject) WriteString(s string) (int, error) {
	return eo.Write([]byte(s))
}

func (eo *GenericEncodeObject) Read(p []byte) (int, error) {
	return eo.content.Read(p)
}

func (eo *GenericEncodeObject) Bytes() []byte {
	return eo.content.Bytes()
}

func (eo *GenericEncodeObject) Hash() ([]byte, error) {
	if eo.hash != nil {
		return bytes.Clone(eo.hash), nil
	}
	checksum, err := HashObject(eo)
	if err != nil {
		return nil, err
	}
	eo.hash = checksum
	return eo.hash, nil
}
