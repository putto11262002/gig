package hash

import (
	"encoding/hex"
	"fmt"
)

type Checksum []byte

func ChecksumToHex(p []byte) string {
	return fmt.Sprintf("%x", p)
}

func ChecksumFromHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
