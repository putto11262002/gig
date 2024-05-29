package util

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
)

func TestReadUntil(t *testing.T) {
	t.Run("read and return all the bytes until the first end is encounter", func(t *testing.T) {
		r := bytes.NewReader([]byte("helloword"))
		expected := []byte("hel")
		end := byte('l')
		actual, err := ReadUntil(r, end)
		if err != nil {
			t.Fatalf("return non-nil error: %v", err)
		}
		if !bytes.Equal(expected, actual) {
			t.Fatalf("expected=%s, actual=%s", expected, actual)
		}
	})

	t.Run("return an EOF error when EOF is encounter before end", func(t *testing.T) {
		r := bytes.NewReader([]byte("helloword"))
		end := byte('a')
		_, err := ReadUntil(r, end)
		if err != io.EOF {
			t.Fatalf("return non-EOF error: %v", err)
		}
	})
}

func generateRandomAsci(n int) ([]byte, error) {
	var buffer bytes.Buffer
	for i := 0; i < n-1; i++ {
		if err := buffer.WriteByte(byte(rand.Intn(127))); err != nil {
			return nil, err
		}
	}
	buffer.WriteByte(byte(127))
	return buffer.Bytes(), nil
}

func BenchmarkReadUntil(b *testing.B) {
	data, err := generateRandomAsci(100000000)
	if err != nil {
		b.Fatalf("failed to generate test data")
	}
	r := bytes.NewReader(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadUntil(r, byte(127))
	}
}
