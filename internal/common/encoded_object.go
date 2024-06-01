package common

import (
	"bytes"
	"io"
)

// EncodedObject represents the encoded version of any object
type EncodedObject interface {
	io.Writer
	NewReader() io.Reader
	Bytes() []byte
	Hash() ([]byte, error)
}

type GenericEncodeObject struct {
	buffer bytes.Buffer
	hash   []byte
}

func NewGenericEncodedObject() *GenericEncodeObject {
	return &GenericEncodeObject{
		buffer: bytes.Buffer{},
	}
}

func NewGenericEncodeObjectFromBuffer(buffer []byte) *GenericEncodeObject {
	return &GenericEncodeObject{
		buffer: *bytes.NewBuffer(buffer),
	}
}

func (eo *GenericEncodeObject) NewReader() io.Reader {
	return bytes.NewReader(eo.buffer.Bytes())
}

func (eo *GenericEncodeObject) Write(p []byte) (int, error) {
	if len(p) > 0 {
		eo.hash = nil
	}
	return eo.buffer.Write(p)
}

func (eo *GenericEncodeObject) Read(p []byte) (int, error) {
	return eo.buffer.Read(p)
}

func (eo *GenericEncodeObject) Bytes() []byte {
	return eo.buffer.Bytes()
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
