package common

import (
	"crypto/sha1"
	"io"
)

// HashObject hash the encoded object with the sha1 hash algorithm
func HashObject(obj EncodedObject) ([]byte, error) {
	hashFn := sha1.New()
	_, err := io.Copy(hashFn, obj.NewReader())
	if err != nil {
		return nil, err
	}
	return hashFn.Sum(nil), nil
}
