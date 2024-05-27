package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestIfTheWrappedReaderClearsTheReadByesFromTheBuffer(t *testing.T) {
	t.Run("After one read the content of buffer remain unchanged", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte("First line\nSecond line\nThrid line\n"))

		originalBuffer := make([]byte, buffer.Len())
		copy(buffer.Bytes(), originalBuffer)
		reader := bytes.NewReader(buffer.Bytes())
		// Keep reading until the end
		chunk := make([]byte, 5)
		for {
			n, err := reader.Read(chunk)
			if err != nil && err != io.EOF {
				t.Errorf("failed to read from reader")
			}
			if n == 0 {
				break
			}
			os.Stdout.Write(chunk[:n])
		}

		for i, _ := range originalBuffer {
			if originalBuffer[i] != buffer.Bytes()[i] {
				t.Error("The buffer has changed")
			}
		}
	})
}
