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
	// Insert inserts the given document into the index. The document is guaranteed to be
	// searchable once the Insert method returns.
	Insert(d doc.Document) error

	// Snapshot returns a point-in-time snapshot of the index that provides a consistent
	// view for searches.
	Snapshot() (Snapshot, error)

	// Seal marks the index as immutable. After Seal is called no more documents can be
	// inserted into the index.
	Seal() error

	// Close closes the index and releases any internal resources.
	Close() error
}

// Snapshot is a point-in-time snapshot of an index. It must be closed to
// release its internal readers.
type Snapshot interface {
	// Readers returns the set of readers contained in the snapshot. Each call to Readers
	// will return independent clones of the same Readers which the caller is responsible
	// for closing.
	Readers() (Readers, error)

	// Size returns the number of Readers in this snapshot.
	Size() (int, error)

	// Close closes the snapshot and releases any internal resources.
	Close() error
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

	// Clone returns an independent clone of this Reader. The returned reader must be
	// closed independently of this Reader.
	Clone() (Reader, error)

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

	if multiErr.Empty() {
		return nil
	}
	return multiErr
}

// Clone returns a clone of all of the internal Readers.
func (rs Readers) Clone() (Readers, error) {
	clones := make(Readers, 0, len(rs))
	for _, r := range rs {
		clone, err := r.Clone()
		if err != nil {
			// Close previously cloned readers.
			clones.Close()
			return nil, err
		}
		clones = append(clones, clone)
	}

	return clones, nil
}
