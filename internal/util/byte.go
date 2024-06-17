package util

import (
	"bytes"
	"io"
)

/*
ReadUtil reads r until the first occurrence of end and returns p, which contains all the bytes read up until end (including end).
If EOF is reached before encountering end, an EOF error is returned, and p will contain the bytes that have been read so far.
*/
func ReadUntil(r io.Reader, end []byte) (p []byte, err error) {
	var buffer bytes.Buffer
	b := make([]byte, 1)
	for {
		_, err := r.Read(b)
		if err != nil {
			return buffer.Bytes(), err
		}
		buffer.Write(b)
		if bytes.Equal(b, end) {
			return buffer.Bytes(), nil
		}
	}
}

// ReadBefore reads bytes from a Reader (r) until the first occurrence of the
// specified end sequence. It returns the bytes read **before** the end sequence.
//
// If EOF is reached before encountering the end sequence, an EOF error is
// returned. In this case, p will contain all the bytes that have been read
// so far (up to the EOF).
func ReadBefore(r io.Reader, end []byte) (p []byte, err error) {
	var buffer bytes.Buffer
	b := make([]byte, 1)
	for {
		n, err := r.Read(b)
		if err != nil && n < 1 {
			return buffer.Bytes(), err
		}
		if bytes.Equal(b, end) {
			break
		}
		buffer.Write(b)
	}
	return buffer.Bytes(), nil
}

// ReadByte reads and return a byte from the read
func ReadByte(r io.Reader) (byte, error) {
	b := make([]byte, 1)
	n, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, io.EOF
	}
	return b[0], nil
}
