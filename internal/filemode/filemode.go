package filemode

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type FileMode uint32

const (
	Directory  FileMode = 0040000
	Regular    FileMode = 0100644
	Executable FileMode = 0100755
	Symlink    FileMode = 0120000
)

var UnsupportedFileModeErr = errors.New("Unsupported file mode")

func NewFomOSFileMode(m fs.FileMode) (FileMode, error) {
	if m.IsDir() {
		return Directory, nil
	}
	if m&os.ModeSymlink != 0 {
		return Symlink, nil
	}
	if m.IsRegular() {

		if m&0100 != 0 {
			return Executable, nil
		}
		return Regular, nil
	}
	return 0, UnsupportedFileModeErr
}

func (fm FileMode) IsDir() bool {
	return fm == Directory
}

func (fm FileMode) isFile() bool {
	return fm == Regular || fm == Executable || fm == Symlink
}

func (fm FileMode) String() string {
	return fmt.Sprintf("%06o", fm)
}
