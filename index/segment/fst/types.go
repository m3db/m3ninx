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

package fst

import (
	"io"
	"regexp"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/postings"
)

const (
	magicNumber  = 0x6D33D0C5
	majorVersion = 1
	minorVersion = 0
)

// Reader represents a FST segment.
type Reader interface {
	segment.Segment

	// MatchRegexp returns a postings list over all documents which match the given
	// regular expression.
	MatchRegexp(field, regexp []byte, compiled *regexp.Regexp) (postings.List, error)

	// MatchAll returns a postings list for all documents known to the Reader.
	MatchAll() (postings.MutableList, error)

	// Docs returns an iterator over the documents whose IDs are in the provided
	// postings list.
	Docs(pl postings.List) (doc.Iterator, error)

	// AllDocs returns an iterator over the documents known to the Reader.
	AllDocs() (doc.Iterator, error)
}

// enforce Reader implements index.Reader's methods. Can't do this in-line as
// index.Reader and segment.Segment share methods.
var _ index.Reader = Reader(nil)

// Writer writes out a FST segment from the provided elements.
type Writer interface {
	// Clear clears the internal state of the Writer to allow re-use.
	Clear()

	// Reset sets the Writer to persist the provide segment.
	// NB(prateek): the provided segment must be a Sealed Mutable segment.
	Reset(s segment.MutableSegment) error

	// MajorVersion is the major version for the writer.
	MajorVersion() int

	// MinorVersion is the minor version for the writer.
	MinorVersion() int

	// Metadata returns metadata about the writer.
	Metadata() []byte

	// WriteDocumentsData writes out the documents data to the provided writer.
	WriteDocumentsData(w io.Writer) error

	// WriteDocumentsIndex writes out the documents index to the provided writer.
	// NB(prateek): this must be called after WriteDocumentsData().
	WriteDocumentsIndex(w io.Writer) error

	// WritePostingsOffsets writes out the postings offset file to the provided
	// writer.
	WritePostingsOffsets(w io.Writer) error

	// WriteFSTTerms writes out the FSTTerms file using the provided writer.
	// NB(prateek): this must be called after WritePostingsOffsets().
	WriteFSTTerms(w io.Writer) error

	// WriteFSTFields writes out the FSTFields file using the provided writer.
	// NB(prateek): this must be called after WriteFSTTerm().
	WriteFSTFields(w io.Writer) error
}
