package pack

import (
	"bytes"
	"math/rand"
	"testing"

	packfileFixture "github.com/codecrafters-io/git-starter-go/internal/pack/fixture"
)

func TestWriteFanoutTable(t *testing.T) {

	packIter := packfileFixture.Packs()
	for {
		pack, ok := packIter.Next()
		if !ok {
			break
		}
		packfile, err := pack.Packfile()
		if err != nil {
			t.Fatal(err)
		}
		packfileReader, err := NewPackfile(nil, packfile)
		if err != nil {
			t.Fatal(err)
		}
		var table bytes.Buffer
		if err := writeFanoutTable(packfileReader, &table); err != nil {
			t.Fatal(err)
		}

		// Compare table entries
		indexfile, err := pack.Indexfile()
		if err != nil {
			t.Fatal(err)
		}

		expectedTable := make([]byte, packfileReader.TotalObjects()*TABLE_ENTRY_SIZE+HEADER_SIZE)
		if _, err := indexfile.ReadAt(expectedTable, 0); err != nil {
			t.Fatal(err)
		}

		// Compare header
		if table.Len() < (HEADER_SIZE) {
			t.Fatalf("expecte header length to be %d bytes", 256*4)
		}

		for i := 0; i < HEADER_SIZE; i += HEADER_ENTRY_SIZE {
			if !bytes.Equal(table.Bytes()[i:i+4], expectedTable[i:i+4]) {
				t.Fatalf("invalid header entry at byte %d", i)
			}
		}

		// Compare table
		for i := TABLE_OFFSET; i < packfileReader.TotalObjects(); i += TABLE_ENTRY_SIZE {
			if !bytes.Equal(table.Bytes()[i:i+TABLE_ENTRY_SIZE], expectedTable[i:i+TABLE_ENTRY_SIZE]) {
				t.Fatalf("invalid table entry at byte %d", i)
			}
		}

		// Check if there are no more bytes
		if len(table.Bytes()) != (packfileReader.TotalObjects()*TABLE_ENTRY_SIZE + HEADER_SIZE) {
			t.Fatalf("table too large")
		}

	}
}

func TestSearch(t *testing.T) {
	packIter := packfileFixture.Packs()
	for {
		pack, ok := packIter.Next()
		if !ok {
			break
		}
		indexfile, err := pack.Indexfile()
		if err != nil {
			t.Fatal(err)
		}

		// Should return offset of objects that exist in the index
		t.Run("Should report offset and found when object exist in the file", func(t *testing.T) {
			for _, indexEntry := range pack.IndexTableEntries {
				offset, found, err := search(indexEntry.Checksum, indexfile)
				if err != nil {
					t.Fatal(err)
				}
				if !found {
					t.Fatalf("expected object %x to be found", indexEntry.Checksum)
				}
				if offset != indexEntry.Offset {
					t.Fatalf("expected object %x offset to be at %d but got %d", indexEntry.Checksum, indexEntry.Offset, offset)
				}
			}
		})

		t.Run("Should not report found when object does not exist in the file", func(t *testing.T) {

			for {
				random := make([]byte, 20)
				for i := range random {
					random[i] = byte(rand.Intn(256))
				}

				for _, indexEntry := range pack.IndexTableEntries {
					if bytes.Equal(indexEntry.Checksum, random) {
						continue
					}
				}

				_, found, err := search(random, indexfile)
				if err != nil {
					t.Fatal(err)
				}
				if found {
					t.Fatalf("do not expect object %x to be found", random)
				}
				break

			}
		})

	}
}
