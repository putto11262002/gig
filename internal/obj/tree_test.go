package object

import (
	"fmt"
	"os"
	"testing"

	common "github.com/codecrafters-io/git-starter-go/internal"
)

func TestDecodeTree(t *testing.T) {
	file, err := os.Open("8c6c162402df682bfd378b07f9f42ddf331474")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	encodedObject, err := common.DecodeMemoryObject(file)
	tree, err := DecodeTree(encodedObject)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v", tree)
}
