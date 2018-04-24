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
	"github.com/m3db/m3ninx/postings"

	"github.com/m3db/m3x/errors"
)

// Index is a collection of searchable documents.
type Index interface {
	Writer

	// Readers returns a set of readers representing a point-in-time snapshot of the index.
	Readers() (Readers, error)

	// Close closes the index and releases any internal resources.
	Close() error
}

// Writer is used to insert documents into an index.
type Writer interface {
	// Insert inserts the given document into the index and returns its ID. The document
	// is guaranteed to be searchable once the Insert method returns.
	Insert(d doc.Document) ([]byte, error)

	// InsertBatch inserts a batch of metrics into the index. The documents are guaranteed
	// to be searchable all at once when the Batch method returns.
	InsertBatch(b Batch) error
}

// Reader provides a point-in-time accessor to the documents in an index.
type Reader interface {
	// MatchTerm returns a postings list over all documents which match the given term.
	MatchTerm(field, term []byte) (postings.List, error)

	// MatchRegexp returns a postings list over all documents which match the given
	// regular expression.
	MatchRegexp(field, regexp []byte, compiled *regexp.Regexp) (postings.List, error)

	// Docs returns an iterator over the documents whose IDs are in the provided
	//  postings list.
	Docs(pl postings.List) (doc.Iterator, error)

	// Close closes the reader and releases any internal resources.
	Close() error
}

// Readers is a slice of Reader.
type Readers []Reader

// Close closes all of the Readers in rs.
func (rs Readers) Close() error {
	multiErr := errors.NewMultiError()
	for _, r := range rs {
		err := r.Close()
		if err != nil {
			multiErr = multiErr.Add(err)
		}
	}
	return multiErr.FinalError()
}
