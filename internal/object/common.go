package object

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

// EncodeObjectHeader encodes the object header according the Git object header encoding format
func encodeObjectHeader(obj Object) []byte {
	return []byte(fmt.Sprintf("%s %d\000", obj.Type(), obj.Size()))
}

// DecodeObjectHeader decodes [ObjHeader] from an encoded object. s is the offset first byte of the object content
func decodeObjectHeader(encodedObject common.EncodedObject) (h *ObjHeader, s int, err error) {
	buffer := encodedObject.Bytes()
	header, _, found := bytes.Cut(buffer, []byte("\000"))
	if !found {
		return nil, 0, errors.New("invalid object format")
	}
	typeBytes, sizeBytes, found := bytes.Cut(header, []byte(" "))
	if !found {
		return nil, 0, errors.New("Invalid header format")
	}
	size, err := strconv.Atoi(string(sizeBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("cannot parse size: %w", err)
	}
	return &ObjHeader{
		Type: string(typeBytes),
		Size: size,
	}, len(header) + 1, nil
}
