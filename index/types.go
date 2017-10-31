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
	"regexp"

	"github.com/m3db/m3ninx/doc"
)

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

// DocID is document identifier internal to each index.
type DocID uint64

// PostingList represents an efficient mapping from doc.Term -> []DocID
type PostingList interface {
	// Insert inserts a reference to document `i` for term `t`.
	Insert(t doc.Term, i DocID)

	// Remove removes the reference (if any) from document `i` to term `t`.
	Remove(t doc.Term, i DocID)

	// Terms retrieves all known terms.
	Terms() []doc.Term

	// NumTerms returns the cardinality of known terms.
	NumTerms() int

	// DocIDs retrieves all known DocIDs.
	DocIDs() []DocID

	// NumDocs returns the cardinality of known docs.
	NumDocs() int

	// Fetch retrieves all documents which are associated with term `t`.
	Fetch(t doc.Term, negate bool) []DocID

	// FetchFuzzy retrieves all documents with terms satisfying the given regexp.
	FetchFuzzy(r *regexp.Regexp, negate bool) []DocID
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
