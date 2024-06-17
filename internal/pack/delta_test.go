package pack

import (
	"bytes"
	"testing"
)

func TestDecodeSparseInt(t *testing.T) {
	presentIdx := byte(0b000101)
	data := []byte{0b11110000, 0b00110011}
	expected := 0x3300F0
	decoded, err := readSparseInt(presentIdx, 3, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if expected != decoded {
		t.Fatalf("expected: %b, actual: %b", expected, decoded)
	}

}
