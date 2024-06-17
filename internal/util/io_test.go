package util

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"testing"
)

// GenerateRandomByteSlice generates a random slice of bytes with a specified length n.
// It returns the generated slice and an error if n is less than or equal to zero.
func GenerateRandomByteSlice(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("n must be a positive integer")
	}

	randomSlice := make([]byte, n)

	_, err := rand.Read(randomSlice)
	if err != nil {
		return nil, err
	}

	return randomSlice, nil
}

func setupLimitedRead(n int) (source []byte, r io.Reader, w io.Writer, err error) {
	source, err = GenerateRandomByteSlice(n)
	if err != nil {
		return source, r, w, err
	}
	r = bytes.NewReader(source)
	w = bytes.NewBuffer(make([]byte, 0, len(source)))
	return source, r, w, err
}

func TestLimitedRead(t *testing.T) {
	t.Run("data is less than the limit", func(t *testing.T) {
		sourceSize := 10
		source, err := GenerateRandomByteSlice(sourceSize)
		if err != nil {
			t.Fatalf("failed to generate test data: %v", err)
		}
		r := bytes.NewReader(source)
		w := bytes.NewBuffer(make([]byte, 0, len(source)))

		limit := sourceSize + 1
		n, err := LimitedRead(r, w, limit)
		if err != nil {
			t.Fatalf("error reading: %v", err)
		}
		if n != sourceSize {
			t.Fatalf("expected to read %d bytes but read %d bytes", sourceSize, n)
		}

		if !bytes.Equal(source, w.Bytes()) {
			t.Fatalf("expected to read:\n%v\nbut read:\n%v", source, w.Bytes())
		}

	})
	t.Run("data is equals to the limit", func(t *testing.T) {

		sourceSize := 10
		source, err := GenerateRandomByteSlice(sourceSize)
		if err != nil {
			t.Fatalf("failed to generate test data: %v", err)
		}
		r := bytes.NewReader(source)
		w := bytes.NewBuffer(make([]byte, 0, len(source)))

		limit := sourceSize
		n, err := LimitedRead(r, w, limit)
		if err != nil {
			t.Fatalf("error reading: %v", err)
		}
		if n != sourceSize {
			t.Fatalf("expected to read %d bytes but read %d bytes", sourceSize, n)
		}

		if !bytes.Equal(source, w.Bytes()) {
			t.Fatalf("expected to read:\n%v\nbut read:\n%v", source, w.Bytes())
		}
	})
	t.Run("data is greater than the limit", func(t *testing.T) {
		sourceSize := 10
		source, err := GenerateRandomByteSlice(sourceSize)
		if err != nil {
			t.Fatalf("failed to generate test data: %v", err)
		}
		r := bytes.NewReader(source)
		w := bytes.NewBuffer(make([]byte, 0, len(source)))

		limit := 0
		_, err = LimitedRead(r, w, limit)
		if err == nil {
			t.Fatalf("expected an error %v but got nil error", ErrReadLimitExceeded)
		}

	})

}
