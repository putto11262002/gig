package common

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
)

// Checksum represents a 20 bytes sha1 hash checksum of an object
type Checksum []byte

func NewChecksumFromHexString(s string) (Checksum, error) {
	return hex.DecodeString(s)
}

func (c Checksum) HexString() string {
	return fmt.Sprintf("%x", c)
}

// HashObject hash the encoded object with the sha1 hash algorithm
func HashObject(obj EncodedObject) ([]byte, error) {
	hashFn := sha1.New()
	_, err := io.Copy(hashFn, obj.NewReader())
	if err != nil {
		return nil, err
	}
	return hashFn.Sum(nil), nil
}
