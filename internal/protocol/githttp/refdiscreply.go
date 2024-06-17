package githttp

import common "github.com/codecrafters-io/git-starter-go/internal"

type RefList map[string]common.Checksum

func NewRefList() RefList {
	return map[string]common.Checksum{}
}

type CapList string

type RefDiscReply struct {
	head      common.Checksum
	refs      RefList
	peeledRef RefList
	caps      CapList // TODO parse row cap list
}

func NewRefDiscReply() *RefDiscReply {
	return &RefDiscReply{
		refs:      NewRefList(),
		peeledRef: NewRefList(),
	}
}

func (r *RefDiscReply) addRef(name string, id common.Checksum) {
	r.refs[name] = id
}

func (r *RefDiscReply) addPeeledRef(name string, id common.Checksum) {
	r.peeledRef[name] = id
}

func (r *RefDiscReply) setHead(id common.Checksum) {
	r.head = id
}

func (r *RefDiscReply) Refs() RefList {
	return r.refs
}

func (r *RefDiscReply) Head() common.Checksum {
	return r.head
}
