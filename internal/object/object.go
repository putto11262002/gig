package object

import (
	"bytes"
	"io"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

// Object represents a set of core functionalities a Git objects must implement.
type Object interface {
	// Return an [io.Reader] which reads raw object content. The data the reader returns is valid until the next modification to the object
	// NewReader() io.Reader
	// Type returns the type of the object
	Type() string
	// String returns the string representation of the object
	String() string
}

// GenericObject represents a generic object.
// It can represent any type objects that implements [Object] interface
type GenericObject struct {
	_type  string
	buffer bytes.Buffer
}

type ObjHeader struct {
	Size int
	Type string
}

func NewObjectHeader(t string, size int) *ObjHeader {
	return &ObjHeader{
		Type: t,
		Size: size,
	}
}

func (o *GenericObject) Type() string {
	return o._type
}

func (o *GenericObject) Size() int {
	return o.buffer.Len()
}

func (o *GenericObject) String() string {
	return o.buffer.String()
}

func DecodeGenericObject(encodedObject common.EncodedObject) (*GenericObject, error) {
	header, s, err := decodeObjectHeader(encodedObject)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, bytes.NewReader(encodedObject.Bytes()[s:]))
	if err != nil {
		return nil, err
	}
	return &GenericObject{
		buffer: buffer,
		_type:  header.Type,
	}, nil
}
