package object

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	common "github.com/codecrafters-io/git-starter-go/internal"
	"github.com/codecrafters-io/git-starter-go/internal/hash"
	"github.com/codecrafters-io/git-starter-go/internal/util"
)

type Actor struct {
	name     string
	email    string
	date     time.Time
	timezone string
}

func NewAuthor(name string, email string) *Actor {
	now := time.Now()
	_, offset := now.Zone()
	return &Actor{
		name:     name,
		email:    email,
		date:     now,
		timezone: fmt.Sprintf("%d", offset),
	}
}

type Commit struct {
	tree      []byte
	parents   [][]byte
	author    Actor
	committer Actor
	size      int
	message   string
}

func (c *Commit) Tree() []byte {
	return c.tree
}

func (c *Commit) SetAuthor(author Actor) {
	c.author = author
}

func (c *Commit) SetCommitter(commiter Actor) {
	c.committer = commiter
}

func (c *Commit) AddParent(parent []byte) {
	c.parents = append(c.parents, parent)
}

func (c *Commit) SetMessage(m string) {
	c.message = m
}
func (c *Commit) SetTree(tree []byte) {
	c.tree = tree
}

func (c *Commit) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("tree %x\n", c.tree))

	for _, parent := range c.parents {
		s.WriteString(fmt.Sprintf("parent %x\n", parent))
	}

	if (Actor{}) != c.author {
		s.WriteString(fmt.Sprintf("author %s <%s> %d %s\n", c.author.name, c.author.email, c.author.date.Unix(), c.author.timezone))
	}

	if c.committer != (Actor{}) {
		s.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n", c.committer.name, c.committer.email, c.committer.date.Unix(), c.committer.timezone))
	}

	s.WriteString("\n")
	s.WriteString(c.message)

	return s.String()
}

func (c *Commit) Type() common.ObjectType {
	return common.OBJ_COMMIT
}

func (c *Commit) Encode(w io.Writer) error {
	var buffer bytes.Buffer
	if _, err := buffer.WriteString(fmt.Sprintf("tree %s\n", hash.ChecksumToHex(c.tree))); err != nil {
		return err
	}
	for _, parent := range c.parents {
		if _, err := buffer.WriteString(fmt.Sprintf("parent %s\n", hash.ChecksumToHex(parent))); err != nil {
			return err
		}
	}
	if c.author != (Actor{}) {
		if _, err := buffer.WriteString(fmt.Sprintf("author %s <%s> %d %s\n", c.author.name, c.author.email, c.author.date.Unix(), c.author.timezone)); err != nil {
			return err
		}
	}
	if c.committer != (Actor{}) {

		if _, err := buffer.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n", c.author.name, c.author.email, c.author.date.Unix(), c.author.timezone)); err != nil {
			return err
		}
	}
	if _, err := buffer.WriteString("\n"); err != nil {
		return err
	}
	if _, err := buffer.WriteString(c.message); err != nil {
		return err
	}
	if _, err := buffer.WriteString("\n"); err != nil {
		return err
	}
	return nil
}

func DecodeCommit(encodedObject common.Object) (*Commit, error) {
	if encodedObject.Type() != common.OBJ_COMMIT {
		return nil, errors.New("invalid object type")
	}
	commit := &Commit{}
	commit.size = encodedObject.Size()
	for {
		line, err := util.ReadBefore(encodedObject, []byte("\n"))
		if err != nil {
			return nil, err
		}
		if len(line) < 1 {
			break
		}

		prefix, data, found := bytes.Cut(line, []byte(" "))
		if !found {
			return nil, errors.New("invalid line encoding")
		}
		switch string(prefix) {
		case "tree":

			tree, err := hash.ChecksumFromHex(string(data))
			if err != nil {
				return nil, err
			}
			commit.SetTree(tree)

		case "parent":

			parent, err := hash.ChecksumFromHex(string(data))
			if err != nil {
				return nil, err
			}
			commit.AddParent(parent)

		case "author":

			author, err := decodeActor(data)
			if err != nil {
				return nil, err
			}
			commit.SetAuthor(*author)

		case "commiter":

			commiter, err := decodeActor(data)
			if err != nil {
				return nil, err
			}

			commit.SetCommitter(*commiter)

		}
	}

	message, err := io.ReadAll(encodedObject)
	if err != nil {
		return nil, err
	}
	message = bytes.TrimSuffix(message, []byte("\n"))
	commit.SetMessage(string(message))
	return commit, nil
}

func decodeActor(b []byte) (*Actor, error) {
	re := regexp.MustCompile(`([\w\s]+) <([\w\._%+-]+@[\w\.-]+\.[a-zA-Z]{2,})> (\d+) ([\+\-]\d+)`)

	// Search for matches
	matches := re.FindSubmatch(b)
	if len(matches) != 5 {
		return nil, errors.New("invalid actor")
	}

	name, email, date, timezone := matches[1], matches[2], matches[3], matches[4]
	dateUnix, err := strconv.ParseInt(string(date), 10, 64)
	if err != nil {
		return nil, err
	}
	return &Actor{
		name:     string(name),
		email:    string(email),
		date:     time.Unix(dateUnix, 0),
		timezone: string(timezone),
	}, nil

}
