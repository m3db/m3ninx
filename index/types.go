// Copyright (c) 2017 Uber Technologies, Inc.
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

package index

import (
	"github.com/m3db/m3ninx/doc"
)

// Readable is a writable index.
type Readable interface {
	// Fetch retrieves the list of documents matching the given criterion.
	Fetch(opts FetchOptions) []doc.Document
}

// Writable is a writable index.
type Writable interface {
	// Insert inserts the given documents with provided fields.
	Insert(d doc.Document) error

	// Update updates the given document.
	Update(d doc.Document) error

	// Delete deletes the given ID.
	Delete(i doc.ID) error

	// Merge merges the given indexes.
	Merge(r ...Readable) error
}

// Writer represents an index writer.
type Writer interface {
	Writable

	// Open prepares the writer to accept writes.
	Open() error

	// Close closes the writer.
	Close() error
}

// Reader represents an index reader.
type Reader interface {
	Readable

	// Open prepares the reader to accept reads.
	Open() error

	// Close closes the reader.
	Close() error
}

// DocID is document identifier internal to each index.
type DocID uint32

// FetchOptions is a group of criterion to filter documents by.
type FetchOptions struct {
	IndexFieldFilters []Filter
	Limit             int64
}

// Filter is a field value filter.
type Filter struct {
	// FiledName is the name of an index field associated with a document.
	FieldName []byte
	// FieldValueFilter is the filter applied to field values for the given name.
	FieldValueFilter []byte
	// Negate specifies if matches should be negated.
	Negate bool
	// Regexp specifies whether the FieldValueFilter should be treated as a Regexp
	// Note: RE2 only, no PCRE (i.e. no backreferences)
	Regexp bool
}
