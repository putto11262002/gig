package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/git-starter-go/internal/object"
	"github.com/codecrafters-io/git-starter-go/internal/util"
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
	printObjContent := flagSet.Bool("p", false, "pretty print (-e | -p) <object> content")
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
	buffer, err := object.ReadObject(checksum)
	if err != nil {
		return "", fmt.Errorf("failed to read object: %v", err)
	}
	bufferReader := bytes.NewReader(buffer)

	// Read object type
	_, err = util.ReadUntil(bufferReader, byte(' '))
	if err != nil {
		return "", err
	}

	// Read size (if you are going to use size remove the trailing null byte
	_size, err := util.ReadUntil(bufferReader, byte(0))
	_size = bytes.Trim(_size, "\000")
	size, err := strconv.Atoi(string(_size))
	if err != nil {
		return "", err
	}

	if *printObjContent {
		return fmt.Sprintf("%s", buffer[len(buffer)-size:]), nil
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
	buffer, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	size := len(buffer)
	hashFn := sha1.New()
	// Write header
	header := []byte(fmt.Sprintf("blob %d\000", size))
	hashFn.Write(header)
	hashFn.Write(buffer)
	checksum := hashFn.Sum(nil)

	if *writeToFile {
		err = os.Mkdir(fmt.Sprintf(".git/objects/%x", checksum[:1]), 0744)
		if err != nil {
			return "", fmt.Errorf("failed to create object directory: %v", err)
		}
		objectFile, err := os.OpenFile(fmt.Sprintf(".git/objects/%x/%x", checksum[:1], checksum[1:]), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to create object file: %w", err)
		}
		defer objectFile.Close()
		zlibW := zlib.NewWriter(objectFile)
		defer zlibW.Close()
		zlibW.Write(header)
		zlibW.Write(buffer)
	}
	return fmt.Sprintf("%x\n", checksum), nil
}

func LsTree() (string, error) {
	flagSet := flag.NewFlagSet("ls-tree", flag.ExitOnError)
	nameOnly := flagSet.Bool("name-only", false, "list only file names")
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
	buffer, err := object.ReadObject(checksum)
	if err != nil {
		return "", err
	}
	bufferReader := bytes.NewReader(buffer)
	objectType, err := util.ReadUntil(bufferReader, byte(' '))
	if err != nil {
		return "", err
	}
	objectType = bytes.Trim(objectType, " ") // Remove trailing space

	if !bytes.Equal([]byte("tree"), objectType) {
		return "", errors.New("not a tree object")
	}

	// Read size (if you are going to use size remove the trailing null byte
	_, err = util.ReadUntil(bufferReader, byte(0))

	var sBuilder strings.Builder
	for {
		mode, err := util.ReadUntil(bufferReader, byte(' '))
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		mode = bytes.Trim(mode, " ")
		name, err := util.ReadUntil(bufferReader, byte(0))
		if err != nil {
			return "", err
		}
		name = bytes.Trim(name, "\000")
		checksum := make([]byte, 20)
		_, err = bufferReader.Read(checksum)
		if err != nil && err != io.EOF {
			return "", err
		}

		if !*nameOnly {
			sBuilder.Write(mode)
			sBuilder.WriteByte(byte(' '))

			sBuilder.Write(object.GetObjectTypeFromFileMode(mode))
			sBuilder.WriteByte(byte(' '))

			encodeChecksum := hex.NewEncoder(&sBuilder)
			encodeChecksum.Write(checksum)
			sBuilder.WriteByte(byte('\t'))
		}
		sBuilder.Write(name)
		sBuilder.WriteByte(byte('\n'))

		if err == io.EOF {
			break
		}
	}

	return sBuilder.String(), nil
}
