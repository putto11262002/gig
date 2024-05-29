package util

import (
	"bytes"
	"io"
)

// ReadTo reads r until the first end is encountered and return p which contains all the bytes read up until end (including end)
// If EOF is reached before end is encountered EOF error is
// returned and the p will contain the bytes that have been read so far
func ReadUntil(r io.ByteReader, end byte) (p []byte, err error) {
	var buffer bytes.Buffer
	for {
		b, err := r.ReadByte()
		if err != nil {
			return buffer.Bytes(), err
		}
		buffer.WriteByte(b)
		if b == end {
			return buffer.Bytes(), nil
		}
	}
}
