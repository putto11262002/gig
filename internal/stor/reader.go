package store

import (
	"bufio"
	"io"
)

// ReadOnlyFile reads the underly file sequentially underlying file using buffered IO
type ReadOnlyFile interface {
	io.ReaderAt
	io.ReadCloser
}

type bufferedReader[T File] struct {
	file   T
	buffer *bufio.Reader
}

func (r *bufferedReader[T]) Read(p []byte) (int, error) {
	return r.buffer.Read(p)
}

func (r *bufferedReader[T]) Close() error {
	return r.file.Close()
}
func newDefaultReader[T File](file T) *bufferedReader[T] {
	return &bufferedReader[T]{
		file:   file,
		buffer: bufio.NewReader(file),
	}
}

type fileReader[T File] struct {
	file T
}

func (r *fileReader[T]) Read(p []byte) (int, error) {
	return r.file.Read(p)
}

func (r *fileReader[T]) ReadAt(p []byte, offset int64) (int, error) {
	return r.file.ReadAt(p, offset)
}

func (r *fileReader[T]) Close() error {
	return r.file.Close()
}
func newFileReader[T File](file T) *fileReader[T] {
	return &fileReader[T]{
		file: file,
	}
}
