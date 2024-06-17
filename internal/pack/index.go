package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"

	common "github.com/codecrafters-io/git-starter-go/internal"
	"github.com/codecrafters-io/git-starter-go/internal/hash"
	store "github.com/codecrafters-io/git-starter-go/internal/stor"
)

const (
	FIRST_LEVEL_ENTRY_SIZE = 4
	MEMORY_LIMIT           = 4 * (1 << 10)

	HEADER_ENTRY_SIZE = 4
	HEADER_ENTRIES    = 256
	HEADER_SIZE       = HEADER_ENTRIES * HEADER_ENTRY_SIZE
	TABLE_OFFSET      = HEADER_SIZE

	OFFSET_SIZE      = 4
	TABLE_ENTRY_SIZE = OFFSET_SIZE + CHECKSUM_LEN
)

// LevelEntry represents an entry in the second layer of the fan-out table
type LevelEntry struct {
	checksum []byte
	offset   uint32
}

type Indexfile struct {
	file     store.ReadOnlyFile
	packfile []byte
}

func NewIndexfile(packChecksum []byte, file store.ReadOnlyFile) *Indexfile {
	return &Indexfile{
		file:     file,
		packfile: packChecksum,
	}
}

func writeIndex(packfile *Packfile, w io.Writer) error {
	hashWriter := hash.NewHashWriter(w, hash.SHA1)

	if err := writeFanoutTable(packfile, hashWriter); err != nil {
		return err
	}

	if _, err := hashWriter.Write(packfile.checksum); err != nil {
		return err
	}
	indexChecksum := hashWriter.Sum(nil)
	w.Write(indexChecksum)
	return nil
}

func (index *Indexfile) search(checksum []byte) (offset int, found bool, err error) {
	return search(checksum, index.file)
}

func search(targetChecksum []byte, index io.ReaderAt) (offset int, found bool, err error) {

	var readOffset int
	readSize := 4
	if targetChecksum[0] > 0 {
		readOffset = int(targetChecksum[0]) - 1
		readSize *= 2
	}
	readOffset *= 4

	chunk := make([]byte, readSize)
	if _, err := index.ReadAt(chunk, int64(readOffset)); err != nil {
		return offset, found, err
	}

	numSeachEntries := binary.BigEndian.Uint32(chunk[:4])
	if targetChecksum[0] > 0 {
		numSeachEntries = binary.BigEndian.Uint32(chunk[4:]) - numSeachEntries
	}

	if numSeachEntries < 1 {
		return offset, found, nil
	}

	var searchOffset uint32
	if targetChecksum[0] == 0 {
		searchOffset = 256 * 4
	} else {
		searchOffset = 256*4 + binary.BigEndian.Uint32(chunk[:4])*24
	}
	searchEntries := make([]byte, numSeachEntries*24)

	if _, err := index.ReadAt(searchEntries, int64(searchOffset)); err != nil {
		return offset, found, err
	}

	for i := 0; i < int(numSeachEntries); i++ {
		offset := binary.BigEndian.Uint32(searchEntries[i*24 : i*24+4])
		checksum := searchEntries[i*24+4 : i*24+24]
		if bytes.Equal(checksum, targetChecksum) {
			return int(offset), true, nil
		}
	}

	return offset, found, nil
}

func writeFanoutTable(r *Packfile, w io.Writer) error {
	firstLevel := [256]uint32{}
	secondLevel := make([]LevelEntry, r.TotalObjects())

	checksumToOffset := map[string]int{}

	for i := 0; i < r.TotalObjects(); i++ {
		objectEntry := &PackObject{}
		offset, err := r.ReadObject(objectEntry)
		if err != nil {

			return err
		}

		fmt.Printf("type: %s\n", objectEntry.objectType)

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
					err := r.ReadObjectAt(int64(offset-objectEntry.baseOffset), ancestor)
					if err != nil {
						return err
					}
				} else {
					baseOffset, ok := checksumToOffset[hash.ChecksumToHex(cur.baseChecksum)]
					if !ok {
						return errors.New("base object is not in current file")
					}
					err := r.ReadObjectAt(int64(baseOffset), ancestor)
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
		checksumToOffset[hash.ChecksumToHex(checksum)] = offset

		// Add checksum and offset to the map

		secondLevel[i] = LevelEntry{offset: uint32(offset), checksum: checksum}
		fmt.Printf("%x %d %d\n", checksum, offset, objectEntry.size)
		firstLevel[int(checksum[0])]++

	}

	// Sort the second level entry by the checksum
	slices.SortFunc(secondLevel, func(i, j LevelEntry) int {
		return bytes.Compare(i.checksum, j.checksum)
	})

	for i := 1; i < len(firstLevel); i++ {
		firstLevel[i] = firstLevel[i] + firstLevel[i-1]
	}

	for _, accCount := range firstLevel {
		if err := writeUnit32BigEndian(accCount, w); err != nil {
			return err
		}

	}

	for _, entry := range secondLevel {
		if err := writeUnit32BigEndian(entry.offset, w); err != nil {
			return err
		}
		if _, err := w.Write(entry.checksum); err != nil {
			return err
		}
	}

	return nil
}

func writeUnit32BigEndian(i uint32, w io.Writer) error {
	chunk := make([]byte, 4)
	binary.BigEndian.PutUint32(chunk, i)
	if _, err := w.Write(chunk); err != nil {
		return err
	}
	return nil
}

// readSecondLevelEntry reads a secondary entry from r
func readSecondLevelEntry(r io.Reader) (entry LevelEntry, err error) {
	offsetBytes := make([]byte, 4)
	if _, err := r.Read(offsetBytes); err != nil {
		return entry, err
	}
	entry.offset = binary.BigEndian.Uint32(offsetBytes)
	checksumBytes := make([]byte, common.CHECKSUM_LEN)
	if _, err := r.Read(checksumBytes); err != nil {
		return entry, err
	}
	return entry, err
}

func readUint32BigEndian(r io.Reader) (uint32, error) {
	chunk := make([]byte, 4)
	if _, err := r.Read(chunk); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(chunk), nil
}

type Stack[T any] struct {
	list []T
}

func (stack *Stack[T]) Push(v T) {
	stack.list = append(stack.list, v)
}

func (stack *Stack[T]) Len() int {
	return len(stack.list)
}

func (stack *Stack[T]) Empty() bool {
	return stack.Len() < 1
}

func (stack *Stack[T]) Pop() (v T, ok bool) {
	if stack.Empty() {
		return v, ok
	}

	v = stack.list[stack.Len()-1]
	stack.list = stack.list[:stack.Len()-1]
	return v, true
}

// switch objectEntry.objectType {
// case common.OBJ_OFS_DELTA:
// 	fmt.Println("ofs")
// 	baseObjectEntry := &PackObject{}
// 	err := r.ReadObjectAt(int64(offset-objectEntry.baseOffset), baseObjectEntry)
// 	if err != nil {
// 		return err
// 	}
// 	targetObject, err := applyDelta(objectEntry.content, baseObjectEntry.content)
// 	if err != nil {
// 		return err
// 	}
// 	checksum, err = common.HashObject(baseObjectEntry.objectType, targetObject)
// 	if err != nil {
// 		return err
// 	}
// case common.OBJ_REF_DELTA:
// 	fmt.Println("refs")
// 	var found bool
// 	var baseOffset int64
// 	for _, entry := range secondLevel {
// 		if bytes.Equal(entry.checksum, objectEntry.baseChecksum) {
// 			baseOffset = int64(entry.offset)
// 			found = true
// 			break
// 		}
// 	}
// 	fmt.Printf("lookg up %x %d\n", objectEntry.baseChecksum, baseOffset)
// 	if !found {
// 		return errors.New("base object not found in current pack file")
// 	}
//
// 	baseObjectEntry := &PackObject{}
// 	err := r.ReadObjectAt(int64(baseOffset), baseObjectEntry)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("length %d", len(baseObjectEntry.content))
// 	targetObject, err := applyDelta(objectEntry.content, baseObjectEntry.content)
// 	if err != nil {
// 		return err
// 	}
// 	checksum, err = common.HashObject(baseObjectEntry.objectType, targetObject)
// 	if err != nil {
// 		return err
// 	}
// default:
// 	checksum, err = common.HashObject(objectEntry.objectType, objectEntry.content)
// 	if err != nil {
// 		return err
// 	}
// }
