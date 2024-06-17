package goit

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	common "github.com/codecrafters-io/git-starter-go/internal"
	object "github.com/codecrafters-io/git-starter-go/internal/obj"
	"github.com/codecrafters-io/git-starter-go/internal/pack"
	"github.com/codecrafters-io/git-starter-go/internal/protocol/githttp"
	store "github.com/codecrafters-io/git-starter-go/internal/stor"
)

const CLONE_USAGE = "mygit clone <git_url>"

func Clone(args []string) {
	flagSet := flag.NewFlagSet("clone", flag.ExitOnError)
	flagSet.Parse(args)
	if flagSet.Arg(0) == "" {
		fmt.Fprintln(os.Stderr, CLONE_USAGE)
		os.Exit(1)
	}

	gitUrl := flagSet.Arg(0)
	outputDir := flagSet.Arg(1)
	if outputDir == "" {
		outputDir = path.Base(gitUrl)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		ExitWithError(err)
	}

	if err := os.Chdir(outputDir); err != nil {
		ExitWithError(err)
	}

	client := githttp.NewGitHttpClient()
	baseContext := context.Background()

	store, err := store.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print("fetching refs...")
	refDiscReply, err := client.GetRefs(baseContext, gitUrl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	packReq := githttp.BuildPacReqFromRefDisc(refDiscReply)

	fmt.Print("\rfetching pack...")
	resBody, err := client.FetchPack(packReq, gitUrl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resBody.Close()

	if _, err := githttp.ReadSktLine(resBody); err != nil {
	}

	fmt.Print("\rwriting packfile...\n")
	packManager := pack.New(store)
	checksum, err := packManager.From(resBody)
	if err != nil {
		ExitWithError(err)
	}
	encodedHead, err := packManager.Object(refDiscReply.Head())
	if err != nil {
		ExitWithError(err)
	}
	head, err := object.DecodeCommit(encodedHead)
	if err != nil {
		ExitWithError(err)
	}
	if err := Checkout(packManager, head.Tree()); err != nil {
		ExitWithError(err)
	}
	fmt.Println("checkout")

	if err := packManager.Unpack(checksum); err != nil {
		ExitWithError(err)
	}

	os.Exit(0)
}

func ExitWithError(err error) {
	fmt.Fprintln(os.Stdout, err)
	os.Exit(1)
}

func Checkout(maanager pack.PackManageer, tree []byte) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return walkTree(maanager, tree, cwd, func(tn *TreeNode) error {

		if err := os.MkdirAll(path.Dir(tn.path), 0755); err != nil {
			return fmt.Errorf("create workspace directory: %w", err)
		}
		file, err := os.OpenFile(tn.path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("open workspace file: %w", err)
		}
		defer file.Close()
		encodedObject, err := maanager.Object(tn.checksum)
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, encodedObject); err != nil {
			return err
		}
		return nil
	})
}

type TreeNode struct {
	path     string
	checksum []byte
}

func walkTree(manager pack.PackManageer, treeChecksum []byte, pathPrefix string, f func(*TreeNode) error) error {
	encodedObject, err := manager.Object(treeChecksum)
	if err != nil {
		return err
	}

	tree, err := object.DecodeTree(encodedObject)
	if err != nil {
		return nil
	}

	entryIter := tree.TreeIter()
	for {

		entry, ok := entryIter.Next()
		if !ok {
			break
		}

		path := path.Join(pathPrefix, entry.Name)

		if entry.Type() == common.OBJ_BLOB {

			treeNode := &TreeNode{
				checksum: entry.Checksum,
				path:     path,
			}

			if err := f(treeNode); err != nil {
				return err
			}

		} else {

			walkTree(manager, entry.Checksum, path, f)

		}

	}
	return nil

}
