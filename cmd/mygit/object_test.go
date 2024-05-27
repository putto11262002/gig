package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
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
				t.Fatalf("failed to read from reader")
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

func TestDecodeBlobObject(t *testing.T) {
	t.Run("decode helloworld", func(t *testing.T) {
		expectedSize := 9
		expectedContent := "helloworld"
		encoded := []byte(fmt.Sprintf("blob %d\000%s", expectedSize, expectedContent))
		var decodedContent bytes.Buffer
		n, err := DecodeBlobObject(bytes.NewReader(encoded), &decodedContent)
		if err != nil {
			t.Errorf("failed to decode blob object: %v", err)
		}
		if expectedSize != n {
			t.Errorf("expected=%d, actual=%d", expectedSize, n)
		}
		if !bytes.Equal(decodedContent.Bytes(), []byte(expectedContent)) {
			t.Errorf("expected=%s, actual=%s", expectedContent, decodedContent.Bytes())
		}
	})
	t.Run("code invalid prefix", func(t *testing.T) {
		encoded := []byte("random 12\000content")
		var decodedContent bytes.Buffer
		_, err := DecodeBlobObject(bytes.NewReader(encoded), &decodedContent)
		if err == nil {
			t.Error("expect error to be returned")
		}

		if !strings.Contains(err.Error(), "invalid format") {
			t.Errorf("expected invalid format error to be return")
		}
	})
}

func TestEncodeBlobObject(t *testing.T) {
	expected := []byte("blob 9\000helloworld")
	input := []byte("helloworld")
	var encoded bytes.Buffer
	if err := EncodeBlobObject(bytes.NewReader(input), &encoded); err != nil {
		t.Errorf("failed to encode: %v", err)
	}
	if bytes.Equal(expected, encoded.Bytes()) {
		t.Errorf("expected=%v, actual=%v", expected, encoded.Bytes())
	}
}
