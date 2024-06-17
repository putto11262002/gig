package pack

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	common "github.com/codecrafters-io/git-starter-go/internal"
	"github.com/codecrafters-io/git-starter-go/internal/hash"
	store "github.com/codecrafters-io/git-starter-go/internal/stor"
	"github.com/codecrafters-io/git-starter-go/internal/util"
)

type PackManageer interface {
	From(io.Reader) ([]byte, error)
	Object([]byte) (common.Object, error)
	ObjectExist([]byte) (bool, error)
	Unpack([]byte) error
}

type DefaulPackManager struct {
	store      store.Store
	indexfiles []*Indexfile
	packfiles  []*Packfile
}

func (manager *DefaulPackManager) getStore() store.Store {
	return manager.store
}

func New(store *store.FSStore) PackManageer {
	return &DefaulPackManager{
		store: store,
	}
}

type PackEntry struct {
	checksum common.Checksum
	offset   int
}

func NewPackEntry(checksum common.Checksum, offset int) PackEntry {
	return PackEntry{
		checksum: checksum,
		offset:   offset,
	}
}

func (manager *DefaulPackManager) Unpack(packfileChecksum []byte) error {
	file, err := manager.store.NewPackReader(hash.ChecksumToHex(packfileChecksum))
	if err != nil {
		return err
	}
	packfile, err := NewPackfile(packfileChecksum, file)
	if err != nil {
		return err
	}

	checksumToOffset := map[string]int{}

	for i := 0; i < packfile.TotalObjects(); i++ {
		objectEntry := &PackObject{}
		offset, err := packfile.ReadObject(objectEntry)
		if err != nil {
			return err
		}

		var targetObject common.Object

		// Entry is an undeltified object
		if !isDeltaObject(objectEntry.objectType) {
			targetObject = common.NewObjectBuffer(objectEntry.objectType, objectEntry.content)
		} else {

			// Etry is a deltified object
			var baseObjectBuffer []byte
			var objectType common.ObjectType

			stack := &Stack[*PackObject]{}
			// push object into stack
			cur := objectEntry
			for {
				ancestor := &PackObject{}
				fmt.Println("adding base")
				fmt.Printf("%x %d\n", cur.baseChecksum, cur.baseOffset)
				stack.Push(cur)
				if cur.objectType == common.OBJ_OFS_DELTA {
					err := packfile.ReadObjectAt(int64(offset-objectEntry.baseOffset), ancestor)
					if err != nil {
						return err
					}
				} else {
					baseOffset, ok := checksumToOffset[hash.ChecksumToHex(cur.baseChecksum)]
					if !ok {
						return errors.New("base object is not in current file")
					}
					err := packfile.ReadObjectAt(int64(baseOffset), ancestor)
					if err != nil {
						return err
					}

					if !isDeltaObject(ancestor.objectType) {
						baseObjectBuffer = ancestor.content
						objectType = ancestor.objectType
						break
					} else {
						cur = ancestor
					}
				}
			}

			// while the stock in not empty
			for !stack.Empty() {
				deltaObject, _ := stack.Pop()
				baseObjectBuffer, err = applyDelta(deltaObject.content, baseObjectBuffer)
				if err != nil {
					return err
				}
			}

			targetObject = common.NewObjectBuffer(objectType, baseObjectBuffer)
		}
		checksum, err := targetObject.Hash()
		if err != nil {
			return err
		}
		objectfile, err := manager.store.ObjectWriter(hash.ChecksumToHex(checksum))
		if err != nil {
			return err
		}
		comp := zlib.NewWriter(objectfile)
		targetObject.Encode(comp)
		comp.Close()
		objectfile.Close()
		checksumToOffset[hash.ChecksumToHex(checksum)] = offset

	}
	return nil

}

func (packagaeManager *DefaulPackManager) From(r io.Reader) ([]byte, error) {
	file, err := packagaeManager.store.NewPackWriter("")
	checksum, err := WritePackfile(r, file)
	if err != nil {
		return nil, err
	}

	file.Rename(hash.ChecksumToHex(checksum))

	file.Close()

	_file, err := packagaeManager.store.NewPackReader(hash.ChecksumToHex(checksum))
	if err != nil {
		return nil, err
	}
	defer _file.Close()

	packfile, err := NewPackfile(checksum, _file)

	if err != nil {
		return nil, err
	}

	indexfile, err := packagaeManager.store.NewPackIndexWriter(hash.ChecksumToHex(checksum))
	if err != nil {
		return nil, err
	}

	defer indexfile.Close()

	if err := writeIndex(packfile, indexfile); err != nil {
		return nil, err
	}

	return checksum, nil
}

// newPackfile returns a complete packfile from store
func (manager *DefaulPackManager) newPackfile(checksum []byte) (*Packfile, error) {
	file, err := manager.store.NewPackReader(hash.ChecksumToHex(checksum))
	if err != nil {
		return nil, err
	}
	return NewPackfile(checksum, file)
}

func (manager *DefaulPackManager) newIndexfile(packfile *Packfile) (*Indexfile, error) {
	file, err := manager.store.NewPackIndexReader(hash.ChecksumToHex(packfile.checksum))
	if err != nil {
		return nil, err
	}
	return &Indexfile{
		file:     file,
		packfile: packfile.checksum,
	}, nil
}

func (manager *DefaulPackManager) IndexfilesIter() (iter *util.CollectionIter[*Indexfile], err error) {

	if manager.indexfiles == nil {
		hexChecksums, err := manager.store.ListPackIndices()
		if err != nil {
			return nil, err
		}

		for _, hexChecksum := range hexChecksums {
			file, err := manager.store.NewPackIndexReader(hexChecksum)
			if err != nil {
				return nil, err
			}
			checksum, err := hash.ChecksumFromHex(hexChecksum)
			if err != nil {
				return nil, err
			}
			indexfile := NewIndexfile(checksum, file)
			manager.indexfiles = append(manager.indexfiles, indexfile)
		}
	}

	return util.NewCollectionIter(manager.indexfiles), nil
}

func (manager *DefaulPackManager) searchIndex(checksum []byte) (offset int, packfile []byte, ok bool, err error) {

	indexfileIter, err := manager.IndexfilesIter()
	if err != nil {
		return offset, packfile, ok, err
	}

	for {
		indexfile, _ok := indexfileIter.Next()
		if !_ok {
			break
		}
		offset, ok, err = indexfile.search(checksum)
		if ok {
			packfile = indexfile.packfile
			break
		}
	}
	return offset, packfile, ok, err
}

func (manager *DefaulPackManager) ObjectExist(checksum []byte) (ok bool, err error) {
	_, _, ok, err = manager.searchIndex(checksum)
	return ok, err
}

func (manager *DefaulPackManager) Object(checksum []byte) (object common.Object, err error) {
	offset, packfileChecksum, ok, err := manager.searchIndex(checksum)
	if err != nil {
		return nil, err
	}

	if !ok {
		return object, err
	}

	// read corresponding packfile and retrun the object
	packfile, err := manager.packfile(packfileChecksum)

	packObject := &PackObject{}
	if err := packfile.ReadObjectAt(int64(offset), packObject); err != nil {
		return nil, err
	}

	if !isDeltaObject(packObject.objectType) {
		return common.NewObjectBuffer(packObject.objectType, packObject.content), nil
	}
	// Etry is a deltified object
	var baseObjectBuffer []byte
	var objectType common.ObjectType

	stack := &Stack[*PackObject]{}
	// push object into stack
	cur := packObject
	for {
		ancestor := &PackObject{}
		fmt.Println("adding base")
		fmt.Printf("%x %d\n", cur.baseChecksum, cur.baseOffset)
		stack.Push(cur)
		if cur.objectType == common.OBJ_OFS_DELTA {
			// TODO: change from obejctEntry.baseOffset to cur.baseOffset in other spot
			err := packfile.ReadObjectAt(int64(offset-cur.baseOffset), ancestor)
			if err != nil {
				return nil, err
			}
		} else {
			baseOffset, _, ok, err := manager.searchIndex(cur.baseChecksum)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errors.New("base object is not in current file")
			}
			err = packfile.ReadObjectAt(int64(baseOffset), ancestor)
			if err != nil {
				return nil, err
			}

			if !isDeltaObject(ancestor.objectType) {
				baseObjectBuffer = ancestor.content
				objectType = ancestor.objectType
				break
			} else {
				cur = ancestor
			}
		}
	}

	// while the stock in not empty
	for !stack.Empty() {
		deltaObject, _ := stack.Pop()
		baseObjectBuffer, err = applyDelta(deltaObject.content, baseObjectBuffer)
		if err != nil {
			return nil, err
		}
	}

	object = common.NewObjectBuffer(objectType, baseObjectBuffer)

	return object, nil
}

func (manager *DefaulPackManager) packfile(checksum []byte) (packfile *Packfile, err error) {

	for _, _packfile := range manager.packfiles {
		if bytes.Equal(_packfile.checksum, checksum) {
			packfile = _packfile
			break
		}
	}
	if packfile != nil {
		return packfile, nil
	}

	packfile, err = manager.newPackfile(checksum)
	if err != nil {
		return nil, err
	}

	manager.packfiles = append(manager.packfiles, packfile)
	return packfile, nil

}
