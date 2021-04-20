package faq

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// FileInfo is a file that is read lazily from an io.Reader and caches the file
// bytes for future reads.
type FileInfo struct {
	path      string
	file      io.ReadCloser
	bufReader *bufio.Reader
}

// File is the interface that faq uses to read file contents, and get access to
// their path for file type detection.
type File interface {
	Reader() *bufio.Reader
	Path() string
	Close() error
}

// Reader returns a bufio.Reader wrapping the file opened.
func (info *FileInfo) Reader() *bufio.Reader {
	return info.bufReader
}

// Path returns the path to the file
func (info *FileInfo) Path() string {
	return info.path
}

// Close closes the file.
func (info *FileInfo) Close() error {
	return info.file.Close()
}

// OpenFile returns a new FileInfo.
func OpenFile(path string) (*FileInfo, error) {
	path = os.ExpandEnv(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file at %s: `%s`", path, err)
	}

	return NewFile(path, file), nil
}

// NewFile returns a FileInfo from a given path and io.ReadCloser.
func NewFile(path string, file io.ReadCloser) *FileInfo {
	return &FileInfo{path, file, bufio.NewReader(file)}
}
