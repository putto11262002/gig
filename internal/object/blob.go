package object

import (
	"bytes"
	"errors"
	"io"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

type Blob struct {
	buffer bytes.Buffer
}

func (b *Blob) Type() string {
	return "blob"
}

func (b *Blob) Size() int {
	return b.buffer.Len()
}

func NewBlobObj(r io.Reader) (*Blob, error) {
	var buffer bytes.Buffer
	_, err := io.Copy(&buffer, r)
	if err != nil {
		return nil, err
	}
	return &Blob{
		buffer,
	}, nil
}

func (b *Blob) String() string {
	return b.String()
}

func EncodeBlob(blob *Blob) (common.EncodedObject, error) {
	encodedBlob := common.NewGenericEncodedObject()
	encodedHeader := encodeObjectHeader(NewObjectHeader(blob.Type(), blob.buffer.Len()))
	_, err := encodedBlob.Write(encodedHeader)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(encodedBlob, &blob.buffer)
	if err != nil {
		return nil, err
	}
	return encodedBlob, nil
}

func DecodeBlob(encodedBlob common.EncodedObject) (*Blob, error) {
	header, s, err := decodeObjectHeader(encodedBlob)
	if err != nil {
		return nil, err
	}
	if header.Type != "blob" {
		return nil, errors.New("invalid blob object")
	}
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, bytes.NewReader(encodedBlob.Bytes()[s:]))
	if err != nil {
		return nil, err
	}
	return &Blob{
		buffer: buffer,
	}, nil
}
