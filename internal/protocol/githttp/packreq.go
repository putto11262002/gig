package githttp

import (
	"fmt"
	"io"

	common "github.com/codecrafters-io/git-starter-go/internal"
)

type PackReq struct {
	want map[string]bool
	done bool
}

func NewPackReq() *PackReq {
	return &PackReq{
		want: map[string]bool{},
	}
}

func (pr *PackReq) AddWant(id common.Checksum) {
	pr.want[fmt.Sprintf("%x", id)] = true
}

func (pr *PackReq) DelWant(id common.Checksum) {
	delete(pr.want, fmt.Sprintf("%x", id))
}

func (pr *PackReq) IsInWant(id common.Checksum) bool {
	_, ok := pr.want[fmt.Sprintf("%x", id)]
	return ok
}

func (pr *PackReq) Done() {
	pr.done = true
}

func (pr *PackReq) Encode(w io.Writer) error {
	encoder := NewPktLineEncoder(w)
	for id, _ := range pr.want {
		if _, err := encoder.WriteLineString(fmt.Sprintf("want %s", id)); err != nil {
			return err
		}
	}
	if err := encoder.WriteFlush(); err != nil {
		return err
	}
	if pr.done {
		if _, err := encoder.WriteLineString("done"); err != nil {
			return err
		}
	}
	return nil
}

func BuildPacReqFromRefDisc(refDisc *RefDiscReply) *PackReq {
	packReq := NewPackReq()
	for _, id := range refDisc.Refs() {
		packReq.AddWant(id)
	}
	packReq.Done()
	return packReq
}
