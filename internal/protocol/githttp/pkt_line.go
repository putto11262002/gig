package githttp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/codecrafters-io/git-starter-go/internal/util"
)

const (
	MAX_LINE_DATA = 65516
)

// Parser for pkt-line formatted data
type PktLineDecoder struct {
	src io.Reader
}

type PktLine []byte

// NewPktLineParser initialise a new pkt-line parser. src is the raw data that is to be parsed
func NewPktLineParser(src io.Reader) *PktLineDecoder {
	return &PktLineDecoder{
		src: src,
	}
}

func (p *PktLineDecoder) Skip(s int) (int, error) {
	return p.src.Read(make([]byte, s))
}

func (p *PktLineDecoder) Read(b []byte) (int, error) {
	return p.src.Read(b)
}

func (p *PktLineDecoder) ReadPktLine() (PktLine, bool, error) {
	line, err := util.ReadBefore(p.src, []byte("\n"))
	if err != nil && err != io.EOF {
		return nil, false, fmt.Errorf("failed to read line: %w", err)
	}
	if bytes.Equal(line, FLUSH_PKT) {
		return nil, true, nil
	}
	_, err = strconv.ParseInt(string(line[:4]), 16, 64)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse pkt-len: %w", err)
	}
	data := line[4:]
	return data, false, nil
}

func ReadSktLine(r io.Reader) (PktLine, error) {
	line, err := util.ReadBefore(r, []byte("\n"))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}
	_, err = strconv.ParseInt(string(line[:4]), 16, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pkt-len: %w", err)
	}
	data := line[4:]
	return data, nil
}

type PktLineEncoder struct {
	dest io.Writer
}

func NewPktLineEncoder(dest io.Writer) *PktLineEncoder {
	return &PktLineEncoder{
		dest: dest,
	}
}

func (e *PktLineEncoder) WriteLine(p []byte) (int, error) {
	if len(p) > MAX_LINE_DATA {
		return 0, errors.New("data exceeds max line data")
	}

	pktLen := fmt.Sprintf("%04x", len(p)+5)
	line := append([]byte(pktLen), p...)
	line = append(line, '\n')

	n, err := e.dest.Write(line)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (e *PktLineEncoder) WriteLineString(s string) (int, error) {
	return e.WriteLine([]byte(s))
}

func (e *PktLineEncoder) WriteFlush() error {
	_, err := e.dest.Write(FLUSH_PKT)
	return err
}
