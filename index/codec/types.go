// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package codec

import "github.com/m3db/m3ninx/doc"

const (
	// FilenameFormat is the format of filenames used by the index.
	FilenameFormat = "%d.%s"
)

// FileType is an enum representing the different types of files created by an index.
type FileType uint32

const (
	// DocumentsFile contains the documents in an index.
	DocumentsFile FileType = iota
)

// Extension returns the extension for a file.
func (t FileType) Extension() string {
	switch t {
	case DocumentsFile:
		return "doc"
	default:
		return ""
	}
}

func (t FileType) String() string {
	switch t {
	case DocumentsFile:
		return "documents"
	default:
		return "unknown"
	}
}

// DocWriter is used to write a documents file.
type DocWriter interface {
	// Init initializes the DocWriter.
	Init() error

	// Write writes a document.
	Write(d doc.Document) error

	// Close closes the DocWriter.
	Close() error
}

// DocReader is used to read a documents file.
type DocReader interface {
	// Init initializes the DocReader.
	Init() error

	// Read reads a document.
	Read() (doc.Document, error)

	// Close closes the DocReader.
	Close() error
}
