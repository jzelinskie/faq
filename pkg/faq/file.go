package faq

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// FileInfo is a file that is read lazily from an io.Reader and caches the file
// bytes for future reads.
type FileInfo struct {
	path   string
	reader io.Reader
	data   []byte
	read   bool
}

// File is the interface that faq uses to read file contents, and get access to
// their path for file type detection.
type File interface {
	Contents() ([]byte, error)
	Path() string
}

// Contents returns the Contents of the file. After the first call, the results
// are cached.
func (info *FileInfo) Contents() ([]byte, error) {
	if !info.read {
		if readCloser, ok := info.reader.(io.ReadCloser); ok {
			defer readCloser.Close()
		}
		var err error
		info.data, err = ioutil.ReadAll(info.reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read file at %s: `%s`", info.path, err)
		}

		info.read = true
	}
	return info.data, nil
}

// Path returns the path to the file
func (info *FileInfo) Path() string {
	return info.path
}

// OpenFile returns a new FileInfo
func OpenFile(path string) (*FileInfo, error) {
	path = os.ExpandEnv(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file at %s: `%s`", path, err)
	}

	return &FileInfo{path: path, reader: file}, nil
}
