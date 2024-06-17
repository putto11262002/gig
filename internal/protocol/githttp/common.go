package githttp

import (
	"bytes"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

var (
	FLUSH_PKT   = []byte{48, 48, 48, 48}
	ZERO_ID     = []byte("0000000000000000000000000000000000000000")
	SP          = []byte(" ")
	HEX_ID_LEN  = 20
	BYTE_ID_LEN = 40
	NUL         = []byte("\x00")
)

func isZeroId(id common.Checksum) bool {
	return bytes.Equal(id, ZERO_ID)
}

func isFlushPkt(b []byte) bool {
	return bytes.Equal(FLUSH_PKT, b)
}
