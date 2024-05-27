package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
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
		obj, err := GetBlobObject([]byte(checksum))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
		fmt.Printf("%s", obj.content)
	case "hash-object":
		usage := "usage: git hash-object [-w] <file>"
		flagSet := flag.NewFlagSet("hash-object", flag.ExitOnError)
		writeToDB := flagSet.Bool("w", false, "write the object into the object database")
		flagSet.Parse(os.Args[2:])
		if flagSet.Arg(0) == "" {
			ExitWithErrorMsg(usage)
		}
		file, err := os.Open(flagSet.Arg(0))
		if err != nil {
			ExitWithFormatErrorMsg("failed opening file: %v", err)
		}
		defer file.Close()
		var buffer bytes.Buffer
		FormatBlobObject(file, &buffer)
		checksum, err := HashBlobObject(bytes.NewReader(buffer.Bytes()))
		if err != nil {
			ExitWithFormatErrorMsg("failed  hashing data: %v", err)
		}
		if *writeToDB {
			if err := WriteObjectToStore(bytes.NewReader(buffer.Bytes()), checksum); err != nil {
				ExitWithFormatErrorMsg("failed to write object to file: %v", err)
			}
		}
		ExitWithFormatMsg("%x\n", checksum)
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

// WriteObjectToStore compress the object using the zlib compression format and write the object to the store with the file naming convention specified in https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
func WriteObjectToStore(r io.Reader, checksum []byte) error {
	err := os.Mkdir(fmt.Sprintf(".git/objects/%x", checksum[:1]), 0744)
	if err != nil {
		return fmt.Errorf("failed to create object directory: %w", err)
	}
	file, err := os.OpenFile(fmt.Sprintf(".git/objects/%x/%x", checksum[:1], checksum[1:]), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create object file: %w", err)
	}
	defer file.Close()
	zlibW := zlib.NewWriter(file)
	defer zlibW.Close()
	io.Copy(zlibW, r)
	return nil
}

// WriteBlobObject reads data r, hash the data with hashFn and write the hash checksum to w.
// If r and/or w implement [io.ReadCloser] and [io.WriteCloser] the caller are responsible for caller Close.
func HashBlobObject(r io.Reader) ([]byte, error) {
	hashFn := sha1.New()
	chunk := make([]byte, 1024)
	for {
		n, err := r.Read(chunk)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		_, err = hashFn.Write(chunk[:n])
		if err != nil {
			return nil, err
		}
		if err == io.EOF {
			break
		}
	}
	checksum := hashFn.Sum(nil)
	return checksum, nil
}

// FormatBlob format the blob object data as as specified in [Git Object Spec].
// The caller is responsible for calling Close if r or/and w implements [io.ReadCloser] and [io.WriteCloser].
// [Git Object Spec]: https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
func FormatBlobObject(r io.Reader, w io.Writer) error {
	buffer, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	// Write the blob header "blob <size>\000"
	fmt.Fprintf(w, "blob %d\000", len(buffer))
	// Write content
	w.Write(buffer)
	return nil
}

type Object struct {
	objType string
	size    int
	content []byte
}

func GetBlobObject(checksum []byte) (*Object, error) {
	file, err := os.Open(getObjFileName(checksum))

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

func getObjFileName(checksum []byte) string {
	return fmt.Sprintf(".git/objects/%s/%s", checksum[:2], checksum[2:])
}
