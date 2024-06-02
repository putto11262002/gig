package plumbing

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/codecrafters-io/git-starter-go/internal/common"
	"github.com/codecrafters-io/git-starter-go/internal/object"
	"github.com/codecrafters-io/git-starter-go/internal/store"
)

func CommitTree(args []string) (output string, err error) {
	// Check if the command arguement is before the flag (command <arguement> <flags>...) . Otherwise, assume command is after flag (command <flags>... <arguement>)
	argBeforeflag := len(args) > 0 && !strings.HasPrefix(args[0], "-")
	usage := "mygit commit-tree [(-p <parent>)...] [(-m <message>)...] <tree>"
	flagSet := flag.NewFlagSet("commit-tree", flag.ExitOnError)
	// TODO - both p and m should be a list
	parent := flagSet.String("p", "", "id of the parent commit object")
	message := flagSet.String("m", "", "commit message")
	if argBeforeflag {
		flagSet.Parse(args[1:])
	} else {
		flagSet.Parse(args)
	}
	var treeChecksumStr string
	if argBeforeflag {
		treeChecksumStr = args[0]
	} else {
		treeChecksumStr = flagSet.Arg(0)
	}
	if treeChecksumStr == "" {
		return output, errors.New(usage)
	}
	if *message == "" {
		return output, errors.New("commit message must not be empty")
	}
	treeChecksum, err := common.NewChecksumFromHexString(treeChecksumStr)
	if err != nil {
		return output, fmt.Errorf("invalid tree checksum: %w", err)
	}
	store, err := store.NewDefaultStore()
	if err != nil {
		return output, err
	}
	// Look up tree from the object store
	encodeObject, err := store.ReadObject(treeChecksum)
	if err != nil {
		return output, fmt.Errorf("failed to read tree object: %w", err)
	}
	_, err = object.DecodeTree(encodeObject)
	if err != nil {
		return output, fmt.Errorf("failed to deocde tree object: %w", err)
	}
	var parents []common.Checksum
	if *parent != "" {
		parentChecksum, err := common.NewChecksumFromHexString(*parent)
		if err != nil {
			return "", err
		}
		encodedParent, err := store.ReadObject(parentChecksum)
		if err != nil {
			return "", err
		}
		genericObject, err := object.DecodeGenericObject(encodedParent)
		if err != nil {
			return "", err
		}
		if genericObject.Type() != "commit" {
			return "", errors.New("invalid parent commit object")
		}
		parents = append(parents, parentChecksum)
	}
	author := object.NewAuthor("Some Author", "example@domain.com")
	commit := object.NewCommit(treeChecksum, *author, parents)
	commit.WriteMessage(*message)
	encodedCommit, err := object.EncodeCommit(commit)
	if err != nil {
		return "", err
	}
	if err := store.WriteObject(encodedCommit); err != nil {
		return "", err
	}
	checksum, _ := encodedCommit.Hash()
	return fmt.Sprintf("%x\n", checksum), nil
}
