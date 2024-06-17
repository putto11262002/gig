package packfileFixture

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	common "github.com/codecrafters-io/git-starter-go/internal"
	store "github.com/codecrafters-io/git-starter-go/internal/stor"
)

type packEntry struct {
	ObjectType common.ObjectType
	Offset     int
}

type IndexTableEntry struct {
	Offset   int
	Checksum []byte
}

type pack struct {
	packfile          string
	indexfile         string
	PackEntries       []packEntry
	IndexTableEntries []IndexTableEntry
}

func (p *pack) Packfile() (store.ReadOnlyFile, error) {
	return readFile(p.packfile)
}

func (p *pack) Indexfile() (store.ReadOnlyFile, error) {
	return readFile(p.indexfile)
}

func readFile(name string) (store.ReadOnlyFile, error) {
	_, p, _, _ := runtime.Caller(0)
	dir, _ := filepath.Split(p)
	file, err := os.Open(path.Join(dir, name))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func Packs() *PackIter {
	return &PackIter{
		packs: packs,
		c:     0,
	}
}

func Pack() pack {
	if len(packs) < 1 {
		panic("no test data")
	}
	return packs[0]
}

type PackIter struct {
	packs []pack
	c     int
}

func (iter *PackIter) Next() (pack, bool) {
	if iter.c >= len(iter.packs) {
		return pack{}, false
	}
	p := iter.packs[iter.c]
	iter.c++
	return p, true
}

var packs = []pack{
	{
		packfile:  "pack-4de80fdcf1ea516333583db744c4a4d1dce005f5.pack",
		indexfile: "pack-4de80fdcf1ea516333583db744c4a4d1dce005f5.idx",
		PackEntries: []packEntry{

			{
				ObjectType: common.OBJ_COMMIT,
				Offset:     12,
			},
			{
				ObjectType: common.OBJ_COMMIT,
				Offset:     180,
			},
			{
				ObjectType: common.OBJ_COMMIT,
				Offset:     346,
			},

			{
				ObjectType: common.OBJ_COMMIT,
				Offset:     515,
			},
			{
				ObjectType: common.OBJ_BLOB,
				Offset:     652,
			},
			{
				ObjectType: common.OBJ_BLOB,
				Offset:     2031,
			},
			{
				ObjectType: common.OBJ_TREE,
				Offset:     2067,
			},
			{
				ObjectType: common.OBJ_TREE,
				Offset:     2145,
			},
			{
				ObjectType: common.OBJ_OFS_DELTA,
				Offset:     2222,
			},
			{
				ObjectType: common.OBJ_TREE,
				Offset:     2240,
			},
			{ObjectType: common.OBJ_TREE,
				Offset: 2287,
			},
			{
				ObjectType: common.OBJ_BLOB,
				Offset:     2334,
			},
		},
		IndexTableEntries: []IndexTableEntry{
			{Checksum: []byte{28, 255, 63, 137, 210, 25, 216, 154, 221, 59, 65, 194, 218, 127, 225, 34, 213, 44, 58, 9}, Offset: 12},
			{Checksum: []byte{184, 78, 167, 240, 57, 87, 75, 215, 172, 85, 67, 37, 154, 235, 81, 25, 19, 101, 198, 234},
				Offset: 180},
			{Checksum: []byte{107, 118, 203, 102, 49, 50, 168, 239, 55, 208, 121, 52, 242, 58, 172, 154, 96, 11, 90, 93},
				Offset: 346},
			{Checksum: []byte{100, 99, 108, 177, 56, 25, 192, 223, 19, 102, 17, 234, 90, 153, 129, 183, 254, 39, 221, 218},
				Offset: 515},
			{Checksum: []byte{56, 228, 194, 221, 174, 158, 116, 173, 64, 228, 100, 148, 33, 156, 100, 177, 142, 122, 242, 196}, Offset: 652},
			{Checksum: []byte{162, 59, 49, 24, 129, 109, 191, 250, 73, 107, 230, 206, 94, 206, 194, 239, 143, 64, 112, 194}, Offset: 2031},
			{Checksum: []byte{217, 30, 161, 5, 31, 225, 161, 9, 162, 237, 233, 133, 197, 109, 161, 92, 113, 23, 207, 32}, Offset: 2067},
			{Checksum: []byte{135, 163, 109, 83, 47, 26, 73, 86, 185, 183, 108, 115, 66, 149, 151, 219, 222, 247, 14, 129}, Offset: 2145},
			{Checksum: []byte{92, 48, 139, 99, 40, 71, 206, 101, 109, 231, 230, 115, 127, 56, 95, 84, 92, 179, 197, 100}, Offset: 2222},
			{Checksum: []byte{1, 233, 133, 232, 231, 16, 235, 217, 108, 194, 24, 73, 57, 14, 40, 54, 47, 196, 107, 64}, Offset: 2240},
			{Checksum: []byte{195, 184, 187, 16, 42, 254, 202, 134, 3, 125, 91, 93, 216, 156, 238, 176, 9, 14, 174, 157}, Offset: 2287},
			{Checksum: []byte{59, 24, 229, 18, 219, 167, 158, 76, 131, 0, 221, 8, 174, 179, 127, 142, 114, 139, 141, 173}, Offset: 2334},
		},
	},
}
