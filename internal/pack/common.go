package pack

import (
	"compress/zlib"
	"io"

	common "github.com/codecrafters-io/git-starter-go/internal"
)

const (
	HEADER_LEN   = 12
	CHECKSUM_LEN = 20
	// The number of bits used to store values
	SIZE_ENCODING_DATA_BITS       = 7
	SIZE_ENCODING_FLAG_MASK  uint = 1 << 7
	SIZE_ENCONDING_DATA_MASK uint = ^SIZE_ENCODING_FLAG_MASK
)

func isDeltaObject(ot common.ObjectType) bool {
	return ot == common.OBJ_OFS_DELTA || ot == common.OBJ_REF_DELTA
}

func decompressData(entryBuffer io.Reader) ([]byte, error) {
	zlibDecom, err := zlib.NewReader(entryBuffer)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(zlibDecom)
	if err != nil {
		return nil, err
	}
	zlibDecom.Close()
	return data, nil
}

type ByteReader interface {
	io.Reader
	io.ByteReader
}

func readCompressedData(entryBuffer io.Reader) ([]byte, error) {
	zlibDecom, err := zlib.NewReader(entryBuffer)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(zlibDecom)
	if err != nil {
		return nil, err
	}
	zlibDecom.Close()
	return data, nil
}

func readVarint(encodedPackBuffer io.ByteReader) (value uint, err error) {
	length := 0
	for {
		b, err := encodedPackBuffer.ReadByte()
		if err != nil {
			return value, err
		}
		byteValue, more := readVarintByte(b)
		value += uint(byteValue) << length
		length += 7
		if !more {
			break
		}
	}
	return value, err
}
