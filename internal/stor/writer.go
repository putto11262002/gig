package store

import (
	"bufio"
	"io"
)

type WriteReadFile interface {
	io.Writer
	io.Reader
	io.ReaderAt
	Rename(string) error
	Close() error
	Sync() error
}

type defaultWriter[T File] struct {
	file   T
	buffer *bufio.Writer
}

func (w *defaultWriter[T]) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

func (w *defaultWriter[T]) Rename(name string) error {
	return w.file.Rename(name)
}

func (w *defaultWriter[T]) Read(p []byte) (int, error) {
	return w.file.Read(p)
}

func (w *defaultWriter[T]) ReadAt(p []byte, offset int64) (int, error) {
	return w.file.ReadAt(p, offset)
}

func (w *defaultWriter[T]) Sync() error {
	return w.file.Sync()
}

func (w *defaultWriter[T]) Close() error {
	if err := w.buffer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}

func NewDefaultWriter[T File](file T) *defaultWriter[T] {
	return &defaultWriter[T]{
		file:   file,
		buffer: bufio.NewWriter(file),
	}
}
