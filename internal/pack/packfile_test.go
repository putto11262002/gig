package pack

import (
	"io"
	"testing"

	packfileFixture "github.com/codecrafters-io/git-starter-go/internal/pack/fixture"
)

func TestReadObject(t *testing.T) {

	packIter := packfileFixture.Packs()
	for {
		pack, ok := packIter.Next()
		if !ok {
			break
		}

		file, err := pack.Packfile()
		defer file.Close()

		packReader, err := NewPackfile(nil, file)
		if err != nil {
			t.Fatal(err)
		}

		if packReader.TotalObjects() != len(pack.PackEntries) {
			t.Fatalf("expected: totalObjects=%d\tactual: totalObjects=%d", len(pack.PackEntries), packReader.TotalObjects())
		}

		object := &PackObject{}

		for i := 0; ; i++ {
			offset, err := packReader.ReadObject(object)
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}

			expectedEntry := pack.PackEntries[i]

			if offset != expectedEntry.Offset || object.objectType != expectedEntry.ObjectType {
				t.Fatalf("expected: offset=%d, objectType=%s\tactual: offset=%d, objectType=%s", expectedEntry.Offset, expectedEntry.ObjectType, offset, object.objectType)
			}
		}
	}
}

func TestReadObjectAt(t *testing.T) {

	pack := packfileFixture.Pack()

	file, err := pack.Packfile()
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	packReader, err := NewPackfile(nil, file)
	if err != nil {
		t.Fatal(err)
	}

	currentObject := &PackObject{}
	previousObject := &PackObject{}
	var prevOffset int

	// Read off the first object
	offset, err := packReader.ReadObject(&PackObject{})
	if err != nil {
		t.Fatal(err)
	}
	prevOffset = offset

	for i := 1; i < packReader.TotalObjects(); i++ {

		offset, err := packReader.ReadObject(currentObject)
		if err != nil {
			t.Fatal(err)
		}

		expectedEntry := pack.PackEntries[i]

		if offset != expectedEntry.Offset || currentObject.objectType != expectedEntry.ObjectType {
			t.Fatalf("current\texpected: offset=%d, objectType=%s\tactual: offset=%d, objectType=%s", expectedEntry.Offset, expectedEntry.ObjectType, offset, currentObject.objectType)
		}

		// read the previous object using ReadObjectAt
		if err := packReader.ReadObjectAt(int64(prevOffset), previousObject); err != nil {
			t.Fatal(err)
		}

		expectedEntry = pack.PackEntries[i-1]

		if previousObject.objectType != expectedEntry.ObjectType {
			t.Fatalf("previous\texpected: objectType=%s\tactual: objectType=%s", expectedEntry.ObjectType, previousObject.objectType)
		}

		prevOffset = offset

	}

}
