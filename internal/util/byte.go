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
