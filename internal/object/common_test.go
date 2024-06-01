package object

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codecrafters-io/git-starter-go/internal/common"
)

func TestEncodeObjectHeader(t *testing.T) {
	object := GenericObject{
		_type:  "blob",
		buffer: *bytes.NewBuffer([]byte{10, 21, 45}),
	}
	expected := []byte("blob 3\000")
	actual := encodeObjectHeader(&object)
	if !bytes.Equal(expected, actual) {
		t.Fatalf("expected: %s, actual: %s", expected, actual)
	}
}

func TestDecodeHeader(t *testing.T) {
	for name, testCase := range map[string]struct {
		input  common.EncodedObject
		header ObjHeader
		s      int
		err    error
	}{
		"decode valid header: blob 11\000hello world": {
			input: common.NewGenericEncodeObjectFromBuffer([]byte("blob 11\000hello world")),
			header: ObjHeader{
				Type: "blob",
				Size: 11,
			},
			s:   8,
			err: nil,
		},
	} {

		t.Run(name, func(t *testing.T) {
			header, s, err := decodeObjectHeader(testCase.input)
			if testCase.err != nil && err == nil {
				t.Fatalf("expected(err): non-nil, actual(err): nil")
			}
			if testCase.err != nil && err != nil && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.err.Error())) {
				t.Fatalf("expected(err.Error()): to contain %s, actual(err.Error()): %s", testCase.err, testCase.err)
			}
			if testCase.err == nil && err != nil {
				t.Fatalf("expected(err): nil, actual(err): %s", err)
			}
			if s != testCase.s {
				t.Fatalf("expected(s): %d, actual(s): %d", testCase.s, s)
			}
			if testCase.header.Type != header.Type || testCase.header.Size != header.Size {
				t.Fatalf("exepected(header): %+v, actual(header): %+v", testCase.header, header)
			}
		})
	}
}
