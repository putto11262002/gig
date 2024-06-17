package object

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	common "github.com/codecrafters-io/git-starter-go/internal"
	"github.com/codecrafters-io/git-starter-go/internal/filemode"
	"github.com/codecrafters-io/git-starter-go/internal/util"
)

type TreeEntry struct {
	Mode     filemode.FileMode
	Name     string
	Checksum []byte
}

func NewTreeEntry(m filemode.FileMode, name string, checksum []byte) *TreeEntry {
	return &TreeEntry{
		Mode:     m,
		Name:     name,
		Checksum: checksum,
	}
}
func encodeTreeEntry(treeObjEntry *TreeEntry) []byte {
	return []byte(fmt.Sprintf("%o %s\000%s", treeObjEntry.Mode, treeObjEntry.Name, treeObjEntry.Checksum))
}

// Type infer the object type from entry file mode
func (e *TreeEntry) Type() common.ObjectType {
	if e.Mode == filemode.Directory {
		return common.OBJ_TREE
	}
	return common.OBJ_BLOB
}

func (e *TreeEntry) String() string {
	return fmt.Sprintf("%06o  %s %x\t%s", e.Mode, e.Type(), e.Checksum, e.Name)
}

// Tree represents a tree object
type Tree struct {
	entries []TreeEntry
}

func (t *Tree) String() string {
	var s strings.Builder
	for _, entry := range t.entries {
		s.WriteString(entry.String())
		s.WriteString("\n")
	}
	return s.String()
}

// newTree create a new Tree object with empty entry slice
func newTree() *Tree {
	return &Tree{
		entries: []TreeEntry{},
	}
}

func (t *Tree) Type() common.ObjectType {
	return common.OBJ_TREE
}

// // Size return the size of the encoded tree object content (entries).
// // The size is not memonized. It is calculate every time the function is called,
// // which is a O(n) operation.
// // TODO: perhaps store size variable in the stsuct and update it every time a new entry is added
// func (t *Tree) Size() int {
// 	var size int
// 	for _, entry := range t.entries {
// 		size += len(encodeTreeEntry(&entry))
// 	}
// 	return size
// }

func (t *Tree) AppendEntry(entry TreeEntry) {
	t.entries = append(t.entries, entry)
}

func (t *Tree) TreeIter() *util.CollectionIter[TreeEntry] {
	return util.NewCollectionIter(t.entries)
}

// func WriteTree(fSys fs.FS, store store.Store) ([]byte, error) {
// 	entries, err := fs.ReadDir(fSys, ".")
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	treeObj := newTree()
// 	for _, entry := range entries {
// 		if entry.Name() == ".git" {
// 			continue
// 		}
// 		if entry.IsDir() {
// 			subFSys, err := fs.Sub(fSys, entry.Name())
// 			if err != nil {
// 				return nil, err
// 			}
// 			checksum, err := WriteTree(subFSys, store)
// 			if err != nil {
// 				return nil, err
// 			}
// 			fm, err := filemode.NewFomOSFileMode(entry.Type())
// 			if err != nil {
// 				if err == filemode.UnsupportedFileModeErr {
// 					continue
// 				}
// 				return nil, err
// 			}
// 			treeEntry := NewTreeEntry(fm, entry.Name(), checksum)
// 			treeObj.AppendEntry(*treeEntry)
// 			continue
// 		}
// 		file, err := fSys.Open(entry.Name())
// 		if err != nil {
// 			return nil, err
// 		}
// 		blob, err := NewBlobObj(file)
// 		if err != nil {
// 			return nil, err
// 		}
// 		encodedBlob, err := EncodeBlob(blob)
// 		if err != nil {
// 			return nil, err
// 		}
// 		err = store.WriteObject(encodedBlob)
// 		// TODO: perhaps ignore already exist error
// 		if err != nil {
// 			return nil, fmt.Errorf("cannot write object to store: %w", err)
// 		}
// 		checksum, _ := encodedBlob.Hash()
// 		fm, err := filemode.NewFomOSFileMode(entry.Type())
// 		if err != nil {
// 			if err == filemode.UnsupportedFileModeErr {
//
// 			}
// 		}
// 		treeEntry := NewTreeEntry(fm, entry.Name(), checksum)
// 		treeObj.AppendEntry(*treeEntry)
//
// 	}
// 	encodedTree, err := EncodeTree(treeObj)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = store.WriteObject(encodedTree)
// 	if err != nil {
// 		return nil, err
// 	}
// 	checksum, _ := encodedTree.Hash()
// 	return checksum, nil
// }

// func encodeTreeContent(tree *Tree) ([]byte, error) {
// 	var buffer bytes.Buffer
// 	for _, entry := range tree.entries {
// 		if _, err := buffer.Write(encodeTreeEntry(&entry)); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return buffer.Bytes(), nil
// }
//
// func EncodeTree(tree *Tree) (common.Object, error) {
// 	encodedTree := common.NewGenericEncodedObject()
// 	encodedContent, err := encodeTreeContent(tree)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if _, err := encodedTree.Write(encodeObjectHeader(NewObjectHeader(tree.Type(), len(encodedContent)))); err != nil {
// 		return nil, err
// 	}
// 	if _, err := encodedTree.Write(encodedContent); err != nil {
// 		return nil, err
// 	}
// 	return encodedTree, nil
// }

func DecodeTree(encodedObject common.Object) (*Tree, error) {
	tree := newTree()
	if encodedObject.Type() != common.OBJ_TREE {
		return nil, errors.New("Invalid tree object")
	}
	for {
		modeByte, err := util.ReadBefore(encodedObject, []byte(" "))
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		name, err := util.ReadBefore(encodedObject, []byte("\000"))
		if err != nil {
			return nil, err
		}

		checksum := make([]byte, 20)
		_, err = encodedObject.Read(checksum)

		// We may encounter EOF here but we are only checking for EOF at the beginning of the loop.
		// To reduce redundancy of EOF checking logic, we ignore EOF here
		if err != nil && err != io.EOF {
			return nil, err
		}

		mode, err := strconv.ParseInt(string(modeByte), 8, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse file mode: %w", err)
		}
		entry := TreeEntry{
			Mode:     filemode.FileMode(mode),
			Name:     string(name),
			Checksum: checksum,
		}
		tree.AppendEntry(entry)
	}

	return tree, nil
}

func (tree *Tree) ListEntryNames() string {
	var s strings.Builder
	for _, entry := range tree.entries {
		s.WriteString(entry.Name)
		s.WriteString("\n")
	}
	return s.String()
}
