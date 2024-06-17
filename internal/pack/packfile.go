package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	common "github.com/codecrafters-io/git-starter-go/internal"
	"github.com/codecrafters-io/git-starter-go/internal/hash"
	store "github.com/codecrafters-io/git-starter-go/internal/stor"
	"github.com/codecrafters-io/git-starter-go/internal/util"
)

const (
	BUFFER_SIZE = 2 * (1 << 20)
)

type PackObject struct {
	objectType   common.ObjectType
	content      []byte
	baseOffset   int    // Present if object type is ofs delta
	baseChecksum []byte // Present if object type is ref delta
	size         int    // The size of the uncompressed object
}

// Packfile represents a packfile in the store
type Packfile struct {
	buffer       []byte
	totalObjects int
	reader       *bytes.Reader
	readObjects  int
	checksum     []byte
}

func WritePackfile(r io.Reader, w io.Writer) ([]byte, error) {
	var buffer bytes.Buffer
	if _, err := util.LimitedRead(r, &buffer, BUFFER_SIZE); err != nil {
		return nil, err
	}
	if buffer.Len() < (HEADER_LEN + CHECKSUM_LEN) {
		return nil, errors.New("pack file too small")
	}

	if _, err := verifyHeader(buffer.Bytes()[:HEADER_LEN]); err != nil {
		return nil, fmt.Errorf("invalid header: %w", err)
	}

	hashFn := hash.New(hash.SHA1)
	if _, err := hashFn.Write(buffer.Bytes()[:buffer.Len()-CHECKSUM_LEN]); err != nil {
		return nil, err
	}

	calChecksum := hashFn.Sum(nil)
	if !bytes.Equal(calChecksum, buffer.Bytes()[buffer.Len()-CHECKSUM_LEN:]) {
		return nil, errors.New("failed to validate checksum")
	}
	if _, err := io.Copy(w, &buffer); err != nil {
		return nil, err
	}
	return calChecksum, nil
}

func NewPackfile(checksum []byte, file store.ReadOnlyFile) (*Packfile, error) {

	var buffer bytes.Buffer
	if _, err := util.LimitedRead(file, &buffer, BUFFER_SIZE); err != nil {
		return nil, err
	}

	reader := bytes.NewReader(buffer.Bytes())

	header := make([]byte, 12)
	if _, err := reader.Read(header); err != nil {
		return nil, err
	}

	numObjects, err := verifyHeader(header)
	if err != nil {
		return nil, err
	}

	return &Packfile{
		buffer:       buffer.Bytes(),
		totalObjects: numObjects,
		reader:       reader,
		checksum:     checksum,
	}, nil
}

// TotalObjects returns the number of objects in the packfile
func (p *Packfile) TotalObjects() int {
	return p.totalObjects
}

// ReadingOffset returns the current reading offset
func (p *Packfile) ReadingOffset() int {
	return int(p.reader.Size()) - p.reader.Len()
}

// ReadObjectAt start reading an object start offset. n is the number of bytes read.
//
// It does not effect the underlying reading offset
func (p *Packfile) ReadObjectAt(offset int64, object *PackObject) (err error) {
	prevOffset := p.ReadingOffset()

	if _, err := p.reader.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	if err = readObject(p.reader, object); err != nil {
		return err
	}

	// Move seeking offset back to where it was
	if _, err := p.reader.Seek(int64(prevOffset), io.SeekStart); err != nil {
		return err
	}

	return err
}

// ReadObject reads an object starting at the underlying reading offset
//
// offset is the offset of the object from the start of the file.
func (p *Packfile) ReadObject(object *PackObject) (offset int, err error) {

	if p.totalObjects-p.readObjects <= 0 {
		return offset, io.EOF
	}

	offset = p.ReadingOffset()

	err = readObject(p.reader, object)
	if err != nil {
		return offset, fmt.Errorf("failed to read object: %w", err)
	}

	p.readObjects++

	return offset, err
}

type ByteCounter struct {
	counter int
}

func (c *ByteCounter) Write(p []byte) (int, error) {
	c.counter += len(p)
	return len(p), nil
}

func (c *ByteCounter) Count() int {
	return c.counter
}

func readObject(r io.Reader, objectEntry *PackObject) error {
	objectType, size, err := readEntryHeader(r)
	if err != nil {
		return fmt.Errorf("failed to read object header: %w", err)
	}
	objectEntry.objectType = objectType
	objectEntry.size = int(size)
	if objectType == common.OBJ_OFS_DELTA {
		offset, err := readOffsetEncoding(r)
		if err != nil {
			return fmt.Errorf("failed to read ofs delta offset: %w", err)
		}
		objectEntry.baseOffset = int(offset)
	}

	if objectType == common.OBJ_REF_DELTA {
		baseChecksum := make([]byte, 20)
		if _, err := r.Read(baseChecksum); err != nil {
			return fmt.Errorf("failed to read ref delta checksum: %w", err)
		}
		objectEntry.baseChecksum = baseChecksum
	}
	data, err := readCompressedData(r)
	if err != nil {
		return fmt.Errorf("failed to read compressed object content: %w", err)
	}
	objectEntry.content = data
	return nil
}

func readEntryHeader(r io.Reader) (common.ObjectType, uint, error) {
	b, err := util.ReadByte(r)
	if err != nil {
		return 0, 0, err
	}
	value, more := readVarintByte(b)
	objectType := common.ObjectType((b >> 4) & 0x7)
	size := uint(b & 0xf)
	length := 4
	for more {
		b, err := util.ReadByte(r)
		if err != nil {
			return 0, 0, err
		}
		value, more = readVarintByte(b)
		size += uint(value) << length
		length += 7
	}
	return objectType, size, nil
}

func readOffsetEncoding(r io.Reader) (uint, error) {
	var offset uint
	var n uint

	for {
		b, err := util.ReadByte(r)
		if err != nil {
			return 0, err
		}
		value, more := readVarintByte(b)
		offset = offset<<7 | uint(value)
		n++
		if !more {
			break
		}
	}
	if n >= 2 {
		for i := uint(1); i < n; i++ {
			offset += i << (7 * i)
		}
	}
	return offset, nil
}

// verifyHeader decodes and verify the header. If the verification is successful a nil error is returned, otherwise return a non-nil error explaining the reason of failure
func verifyHeader(header []byte) (int, error) {
	if len(header) != HEADER_LEN {
		return 0, errors.New(fmt.Sprintf("expected header to be %d bytes", HEADER_LEN))
	}
	signature := header[:4]
	if !bytes.Equal(signature, []byte("PACK")) {
		return 0, errors.New("expected PACK")
	}
	// Version
	version := binary.BigEndian.Uint32(header[4:8])
	if version != 2 && version != 3 {
		return 0, errors.New("expected version 2 or 3")
	}
	// Number of objects
	numObjects := binary.BigEndian.Uint32(header[8:])
	return int(numObjects), nil
}

// readVarintByte returns the first seven bits as the value and check if there is more byte by examining if MSB is set
func readVarintByte(b byte) (value uint8, more bool) {
	return b & 0x7f, b&0x80 != 0
}
