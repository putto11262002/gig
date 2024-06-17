package common

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/codecrafters-io/git-starter-go/internal/hash"
)

const (
	CHECKSUM_LEN = 20
)

// Checksum represents a 20 bytes sha1 hash checksum of an object
type Checksum []byte

func NewChecksumFromHexString(s string) (Checksum, error) {
	return hex.DecodeString(s)
}

func (c Checksum) HexString() string {
	return fmt.Sprintf("%x", c)
}

func (c Checksum) Equal(o Checksum) bool {
	return bytes.Equal(c, o)
}

func HashObject(objectType ObjectType, content []byte) ([]byte, error) {
	hashFn := hash.New(hash.SHA1)
	if err := EncodeObject(objectType, content, hashFn); err != nil {
		return nil, err
	}
	return hashFn.Sum(nil), nil
}
