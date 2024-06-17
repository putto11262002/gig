package pack

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

func applyDelta(instruction []byte, base []byte) ([]byte, error) {
	instructionReader := bytes.NewReader(instruction)
	baseObjectSize, err := readVarint(instructionReader)
	if err != nil {
		return nil, err
	}

	targetObjectSize, err := readVarint(instructionReader)
	if err != nil {
		return nil, err
	}
	fmt.Printf("expected base %d base %d target %d\n", len(base), baseObjectSize, targetObjectSize)

	if int(baseObjectSize) != len(base) {
		return nil, errors.New("invalid base object size")
	}

	targetBuffer := bytes.NewBuffer(make([]byte, 0, targetObjectSize))
	for {
		if err := applyInstruction(instructionReader, base, targetBuffer); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	return targetBuffer.Bytes(), nil
}

func applyInstruction(instructionReader ByteReader, base []byte, target io.Writer) error {
	header, err := instructionReader.ReadByte()
	if err != nil {
		return err
	}
	isCopy := (header & (1 << 7)) != 0
	fmt.Printf("copy %v \n", isCopy)
	if isCopy {
		// Readingg offset
		offset, err := readSparseInt(header, 4, instructionReader)
		if err != nil {
			return err
		}
		size, err := readSparseInt(header>>4, 3, instructionReader)

		// Check if the offset and size are valid given the base object
		if offset < 0 || (offset+size) > len(base) {
			return errors.New("valid offset and size")
		}
		copyiedData := base[offset : offset+size]
		target.Write(copyiedData)
	} else {
		size := header & ((1 << 7) - 1)
		dataToAppend := make([]byte, size)
		if _, err := instructionReader.Read(dataToAppend); err != nil {
			return err
		}

		target.Write(dataToAppend)
	}
	return nil
}

func readSparseInt(presentIndex byte, presentIndexBits int, data io.ByteReader) (int, error) {
	value := 0

	for i := 0; i < presentIndexBits; i++ {
		zeroByte := presentIndex&(1<<(i)) == 0
		var byteToAppend byte
		if zeroByte {
			byteToAppend = 0
		} else {
			var err error
			byteToAppend, err = data.ReadByte()
			if err != nil {
				return value, err
			}
		}
		value |= int(byteToAppend) << (8 * i)
	}
	return value, nil
}
