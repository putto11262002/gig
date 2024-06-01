package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/git-starter-go/internal/object"
	"github.com/codecrafters-io/git-starter-go/internal/store"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		err := Init()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
		fmt.Println("Initialized git directory")
	case "cat-file":
		msg, err := CatFile()
		if err != nil {
			ExitWithError(err)
		}
		ExitWithMsg(msg)
	case "hash-object":
		msg, err := HashObject()
		if err != nil {
			ExitWithError(err)
		}
		ExitWithMsg(msg)
	case "ls-tree":
		msg, err := LsTree()
		if err != nil {
			ExitWithError(err)
		}
		ExitWithMsg(msg)
	case "write-tree":
		msg, err := WriteTree(os.Args[2:])
		if err != nil {
			ExitWithError(err)
		}
		ExitWithMsg(msg)
	default:
		ExitWithFormatErrorMsg("Unknown command %s", command)
	}
}

func ExitWithMsg(msg string) {
	fmt.Fprintf(os.Stdout, msg)
	os.Exit(0)
}

func ExitWithFormatMsg(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
	os.Exit(0)
}

func ExitWithErrorMsg(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func ExitWithFormatErrorMsg(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", format), a...)
	os.Exit(1)
}

func ExitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func Init() error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", dir, err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("failed writing file: %w\n", err)
	}
	return nil
}

// TODO return if file open succesfully if flag is -e. There no need to read and parse the file
func CatFile() (string, error) {
	usage := "useage: mygit cat-file (-e | -p) <object>"
	flagSet := flag.NewFlagSet("cat-file", flag.ExitOnError)
	printObjContent := flagSet.Bool("p", false, "pretty print <object> content")
	checkObjExist := flagSet.Bool("e", false, "check if <object> exists")
	flagSet.Parse(os.Args[2:])
	// Either one of the flags can be used, but not both simultaneously, nor can they both be absent
	if *printObjContent && *checkObjExist || (!*printObjContent && !*checkObjExist) {
		return "", errors.New(usage)
	}
	if flagSet.Arg(0) == "" {
		return "", errors.New(usage)
	}
	checksum, err := hex.DecodeString(flagSet.Arg(0))
	if err != nil {
		return "", err
	}
	store, err := store.NewDefaultStore()
	if err != nil {
		return "", err
	}
	encodedObject, err := store.ReadObject(checksum)
	if err != nil {
		return "", fmt.Errorf("failed to read object: %v", err)
	}

	genericObject, err := object.DecodeGenericObject(encodedObject)
	if err != nil {
		return "", err
	}

	if *printObjContent {
		return genericObject.String(), nil
	}
	return "", nil
}

func HashObject() (string, error) {
	usage := "usage: git hash-object [-w] <file>"
	flagSet := flag.NewFlagSet("hash-object", flag.ExitOnError)
	writeToFile := flagSet.Bool("w", false, "write the object into the object database")
	flagSet.Parse(os.Args[2:])
	if flagSet.Arg(0) == "" {
		return "", errors.New(usage)
	}
	file, err := os.Open(flagSet.Arg(0))
	if err != nil {
		return "", fmt.Errorf("failed opening file: %v", err)
	}
	defer file.Close()
	blobObj, err := object.NewBlobObj(file)
	if err != nil {
		return "", err
	}
	encodedObject, err := object.EncodeBlob(blobObj)
	if err != nil {
		return "", err
	}
	if *writeToFile {
		store, _ := store.NewDefaultStore()
		err := store.WriteObject(encodedObject)
		if err != nil {
			return "", err
		}
	}
	checksum, err := encodedObject.Hash()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x\n", checksum), nil
}

func LsTree() (string, error) {
	flagSet := flag.NewFlagSet("ls-tree", flag.ExitOnError)
	// nameOnly := flagSet.Bool("name-only", false, "list only file names")
	flagSet.Parse(os.Args[2:])
	usage := "usage: git ls-tree [options] <tree-ish>\n"
	if flagSet.Arg(0) == "" {
		var sBuilder strings.Builder
		sBuilder.WriteString(usage)
		flagSet.SetOutput(&sBuilder)
		flagSet.PrintDefaults()
		return "", errors.New(sBuilder.String())
	}
	checksumStr := flagSet.Arg(0)
	checksum, err := hex.DecodeString(checksumStr)
	if err != nil {
		return "", err
	}
	store, err := store.NewDefaultStore()
	if err != nil {
		return "", err
	}
	encodedObject, err := store.ReadObject(checksum)
	if err != nil {
		return "", err
	}
	tree, err := object.DecodeTree(encodedObject)
	return tree.String(), nil
}

func WriteTree(args []string) (string, error) {
	store, _ := store.NewDefaultStore()
	fSys := os.DirFS(".")
	checksum, err := object.WriteTree(fSys, store)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x\n", checksum), nil
}
