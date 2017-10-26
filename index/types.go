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

import "github.com/m3db/m3ninx/doc"

// Writer represents an index writer.
type Writer interface {
	// Insert inserts the given documnet with provided fields.
	Insert(d doc.Document) error

	// Delete deletes the given ID.
	Delete(i doc.ID) error

	// Merge merges the given indexes.
	Merge(r ...Reader) error
}

// Reader represents an index reader.
type Reader interface {
	// Fetch retrieves the list of documents matching the given criterion.
	Fetch(opts FetchOptions) []doc.Document
}

// FetchOptions is a group of criterion to filter documents.
type FetchOptions struct {
	Filters         []Filter
	Limit           int64
	SkipID          bool
	SkipBytes       bool
	SkipIndexFields bool
}

// Filter is a field value filter.
type Filter struct {
	FieldName        []byte
	FieldValueFilter []byte
	Negate           bool
	Regexp           bool // RE2 only, no PCRE
}
