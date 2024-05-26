package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
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
		flagSet := flag.NewFlagSet("cat-file", flag.ExitOnError)
		printObjContent := flagSet.Bool("p", false, "pretty print (-e | -p) <object> content")
		checkObjExist := flagSet.Bool("e", false, "check if <object> exists")
		flagSet.Parse(os.Args[2:])
		if *printObjContent && *checkObjExist || (!*printObjContent && !*checkObjExist) {
			fmt.Fprintf(os.Stderr, "useage: mygit cat-file (-e | -p) <object>")
			os.Exit(1)
		}
		if flagSet.Arg(0) == "" {
			fmt.Fprintf(os.Stderr, "useage: mygit cat-file (-e | -p) <object>")
			os.Exit(1)
		}
		checksum := flagSet.Arg(0)
		obj, err := GetBlobObject(checksum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
		fmt.Printf("%s", obj.content)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func Init() error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Error creating directory %s: %w", dir, err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("Error writing file: %w\n", err)
	}
	return nil
}

type Object struct {
	objType string
	size    int
	content []byte
}

func GetBlobObject(checksum string) (*Object, error) {
	file, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", checksum[:2], checksum[2:]))

	if err != nil {
		return nil, err
	}

	zlibReader, err := zlib.NewReader(file)
	defer zlibReader.Close()
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	io.Copy(&buff, zlibReader)

	objType, err := buff.ReadBytes(byte(' '))
	if err != nil {
		return nil, err
	}
	objType = bytes.Trim(objType, " ")

	sizeByte, err := buff.ReadBytes(byte(0))
	if err != nil {
		return nil, err
	}
	sizeByte = bytes.Trim(sizeByte, "\000")
	size, err := strconv.Atoi(string(sizeByte))
	if err != nil {
		return nil, err
	}

	buff.Truncate(size)
	content := buff.Bytes()

	return &Object{
		objType: string(objType),
		size:    size,
		content: content,
	}, nil
}
