package object

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/git-starter-go/internal/common"
	testing_helper "github.com/codecrafters-io/git-starter-go/internal/test"
)

func TestEncodeBlob(t *testing.T) {
	for name, testCase := range map[string]struct {
		input         *Blob
		encodedObject common.EncodedObject
		err           error
	}{
		"encode hello world blob": {
			input: &Blob{
				buffer: *bytes.NewBuffer([]byte("hello world")),
			},
			encodedObject: common.NewGenericEncodeObjectFromBuffer([]byte("blob 11\000hello world")),
			err:           nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			encodedObject, err := EncodeBlob(testCase.input)
			testing_helper.AssertReturnError(t, testCase.err, err)
			if !bytes.Equal(testCase.encodedObject.Bytes(), encodedObject.Bytes()) {
				t.Fatalf("expected: %s, actual: %s", testCase.encodedObject.Bytes(), encodedObject.Bytes())
			}
		})
	}
}
