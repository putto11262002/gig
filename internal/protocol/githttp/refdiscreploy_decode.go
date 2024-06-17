package githttp

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	common "github.com/codecrafters-io/git-starter-go/internal"
)

type RefDiscReployDecoder struct {
	parser  *PktLineDecoder
	decoded *RefDiscReply
}

func NewRefDiscReplyDecoder(src io.Reader) *RefDiscReployDecoder {
	return &RefDiscReployDecoder{
		parser:  NewPktLineParser(src),
		decoded: NewRefDiscReply(),
	}
}

func (d *RefDiscReployDecoder) Decode() (*RefDiscReply, error) {
	serviceSig, flush, err := d.parser.ReadPktLine()
	if err != nil {
		return nil, err
	} else if flush {
		return nil, io.EOF
	}
	if string(serviceSig) != "# service=git-upload-pack" {
		return nil, err
	}

	// Skip flush flush-pkt
	flushPkt := make([]byte, len(FLUSH_PKT))
	if _, err := d.parser.Read(flushPkt); err != nil {
		return nil, err
	}
	if !isFlushPkt(flushPkt) {
		if _, flush, err := d.parser.ReadPktLine(); err != nil {
			return nil, err
		} else if !flush {
			return nil, errors.New("expected flush-pkt")
		}
		return nil, errors.New("expected flush-pkt")
	}
	emptyList, err := d.decodeRefListHeader()
	if err != nil {
		return nil, err
	}
	if emptyList {
		return d.decoded, nil
	}

	if err := d.decodeRefList(); err != nil {
		return nil, err
	}
	return d.decoded, nil
}

func (d *RefDiscReployDecoder) decodeRefListHeader() (empty bool, err error) {
	line, flush, err := d.parser.ReadPktLine()
	if err != nil {
		return empty, fmt.Errorf("failed to read ref list header: %w", err)
	}
	if flush {
		return true, nil
	}
	encodedRef, encodedCap, found := bytes.Cut(line, NUL)
	if !found {
		return empty, errors.New("expected SP between object id and name")
	}
	capList, err := decodeCapList(encodedCap)
	if err != nil {
		return empty, err
	}
	d.decoded.caps = capList
	_, id, _, err := decodeRef(encodedRef)
	if err != nil {
		return empty, err
	}
	d.decoded.setHead(id)
	return empty, nil
}

func (d *RefDiscReployDecoder) decodeRefList() error {
	for {
		line, flush, err := d.parser.ReadPktLine()
		if err != nil {
			return err
		}
		if flush {
			break
		}
		name, id, peeled, err := decodeRef(line)
		if err != nil {
			return err
		}
		if peeled {
			d.decoded.addPeeledRef(name, id)
		} else {
			d.decoded.addRef(name, id)
		}
	}
	return nil
}

func decodeRef(encodedRef []byte) (string, common.Checksum, bool, error) {
	hexId, name, found := bytes.Cut(encodedRef, []byte(" "))
	if !found {
		return "", nil, false, errors.New("expected SP between object id and name")
	}
	if len(hexId) != BYTE_ID_LEN {
		return "", nil, false, errors.New("invalid id length")
	}

	id, err := hex.DecodeString(string(hexId))
	if err != nil {
		return "", nil, false, err
	}

	peeled := bytes.HasSuffix(name, []byte("^{}"))
	if peeled {
		name = bytes.TrimRight(name, "^{}")
	}
	return string(name), id, peeled, nil
}

func decodeCapList(encodedCapList []byte) (CapList, error) {
	return CapList(encodedCapList), nil
}
