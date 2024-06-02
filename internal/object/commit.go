package object

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

type Author struct {
	name     string
	email    string
	date     time.Time
	timezone string
}

func NewAuthor(name string, email string) *Author {
	now := time.Now()
	_, offset := now.Zone()
	return &Author{
		name:     name,
		email:    email,
		date:     now,
		timezone: fmt.Sprintf("%d", offset),
	}
}

type Commit struct {
	tree          common.Checksum
	parents       []common.Checksum
	author        Author
	message       *bytes.Buffer
	encodedCommit *bytes.Buffer
}

func NewCommit(tree common.Checksum, author Author, parents []common.Checksum) *Commit {
	return &Commit{
		tree:          tree,
		message:       &bytes.Buffer{},
		parents:       parents,
		encodedCommit: nil,
		author:        author,
	}
}

func (c *Commit) SetAuthor(author Author) {
	c.author = author
}

func (c *Commit) AddParent(parent common.Checksum) {
	c.parents = append(c.parents, parent)
}

func (c *Commit) WriteMessage(m string) (int, error) {
	return c.message.WriteString(m)
}

func (c *Commit) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("tree %x\n", c.tree))
	for _, parent := range c.parents {
		s.WriteString(fmt.Sprintf("parent %x\n", parent))
	}
	if (Author{}) != c.author {
		s.WriteString(fmt.Sprintf("author %s <%s> %d %s\n", c.author.name, c.author.email, c.author.date.Unix(), c.author.timezone))
		s.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n", c.author.name, c.author.email, c.author.date.Unix(), c.author.timezone))
	}
	s.WriteString("\n")
	s.Write(c.message.Bytes())
	return s.String()

}

func (c *Commit) Type() string {
	return "commit"
}

func encodeCommitContent(commit *Commit) ([]byte, error) {
	var encodedContent bytes.Buffer
	if _, err := encodedContent.WriteString(fmt.Sprintf("tree %s\n", commit.tree.HexString())); err != nil {
		return nil, err
	}
	for _, parent := range commit.parents {
		if _, err := encodedContent.WriteString(fmt.Sprintf("parent %s\n", parent.HexString())); err != nil {
			return nil, err
		}
	}
	if _, err := encodedContent.WriteString(fmt.Sprintf("author %s <%s> %d %s\n", commit.author.name, commit.author.email, commit.author.date.Unix(), commit.author.timezone)); err != nil {
		return nil, err
	}
	if _, err := encodedContent.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n", commit.author.name, commit.author.email, commit.author.date.Unix(), commit.author.timezone)); err != nil {
		return nil, err
	}
	if _, err := encodedContent.WriteString("\n"); err != nil {
		return nil, err
	}
	if _, err := encodedContent.Write(commit.message.Bytes()); err != nil {
		return nil, err
	}
	if _, err := encodedContent.WriteString("\n"); err != nil {
		return nil, err
	}
	return encodedContent.Bytes(), nil
}

func EncodeCommit(commit *Commit) (common.EncodedObject, error) {
	encodedCommit := common.NewGenericEncodedObject()
	encodedContent, err := encodeCommitContent(commit)
	if err != nil {
		return nil, err
	}
	if _, err := encodedCommit.Write(encodeObjectHeader(NewObjectHeader(commit.Type(), len(encodedContent)))); err != nil {
		return nil, err
	}
	if _, err := encodedCommit.Write(encodedContent); err != nil {
		return nil, err
	}
	return encodedCommit, err
}

// func DecodeCommit(encodedCommit common.EncodedObject) (*Commit, error) {
//
// }
