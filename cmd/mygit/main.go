package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strconv"

)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

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
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "useage: mygit cat-file <object>")
			os.Exit(1)
		}
		checksum := os.Args[2] 
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
		size: size,
		content: content,
	}, nil
}
