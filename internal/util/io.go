package util

import (
	"errors"
	"io"
)

var (
	ErrReadLimitExceeded = errors.New("read limit exceeded")
)

// LimitedRead reads until EOF or limit. If the limit is reach and there is stll more bytes to be read,
// [ErrReadLimitExceeded] error is returned.
func LimitedRead(r io.Reader, w io.Writer, limit int) (int, error) {

	limitedReader := io.LimitReader(r, int64(limit))
	chunk := make([]byte, 1024)
	var bytesWritten int
	var readErr error
	var writeErr error
	var n int

	for {
		n, readErr = limitedReader.Read(chunk)

		if readErr != nil && readErr != io.EOF {
			return bytesWritten, readErr
		}

		n, writeErr = w.Write(chunk[:n])
		bytesWritten += n
		if writeErr != nil {
			return bytesWritten, writeErr
		}

		if readErr == io.EOF {
			break
		}

	}

	// Check if EOF is due to end of actual file or the file is larger than the limit
	// Check if there is still more data in r

	if _n, err := r.Read(make([]byte, 1)); _n > 0 {
		return n, ErrReadLimitExceeded
	} else if err != nil && err != io.EOF {
		return n, err
	}

	return bytesWritten, nil

}
