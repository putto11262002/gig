package hash

import (
	"crypto/sha1"
	"hash"
	"io"
)

type Hash hash.Hash

type hashAlgo string

const (
	SHA1 hashAlgo = "sha1"
)

type Hasher struct {
	hashFn Hash
}

func New(algo hashAlgo) (hash Hash) {
	switch algo {
	case SHA1:
		return sha1.New()
	default:
		panic("invalid hashing algorithm")
	}
}

type HashWriter struct {
	hash Hash
	w    io.Writer
}

func (hw *HashWriter) Write(p []byte) (int, error) {
	return hw.w.Write(p)
}

// HashWriter writes to the underlying writer while updating the hash state with the written data
func NewHashWriter(w io.Writer, algo hashAlgo) *HashWriter {
	hash := New(algo)
	mw := io.MultiWriter(w, hash)
	return &HashWriter{
		hash: hash,
		w:    mw,
	}
}

func (hw *HashWriter) Sum(p []byte) []byte {
	return hw.hash.Sum(p)
}
