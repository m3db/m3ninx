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
	"github.com/m3db/m3ninx/idx/doc"
)

// Index is a collection of searchable documents.
type Index interface {
	Writer

	// Snapshot returns a point-in-time snapshot of the index that provides
	// a consistent view for searches.
	Snapshot() (Snapshot, error)
}

// Writer is the interface for adding documents to an index.
type Writer interface {
	// Open initializes the index for use.
	Open() error

	// Insert inserts the given document into the index. The document is
	// guaranteed to be searchable once the Insert method returns.
	Insert(d doc.Document) error

	// Closes closes the index.
	Close() error
}

// Snapshot is a point-in-time snapshot of an index. It must be closed to
// release the internal readers.
type Snapshot interface {
	// Readers returns the set of readers contained in the snapshot.
	Readers() []Reader

	// Close closes the snapshot and releases any internal resources.
	Close() error
}

// Reader provides a point-in-time accessor to the documents in an index.
type Reader interface {
	// MatchTerm returns a postings list of the IDs of all documents which
	// exactly match the given field and value.
	MatchTerm(field, value []byte) (PostingsList, error)

	// MatchRegex returns a postings list of the IDs of all documents which
	// match the given field and regular expression.
	MatchRegex(field, pattern []byte) (PostingsList, error)

	// Docs returns an iterator over the documents in the segment corresponding
	// to the provided postings list. If fields is non-empty then only the
	// specified fields are returned in the documents.
	Docs(pl PostingsList, fields [][]byte) (doc.Iterator, error)
}

// DocID is the unique identifier of a document in the index.
// TODO: Use a uint64, currently we only use uint32 because we use
// Roaring Bitmaps for the implementation of the postings list.
type DocID uint32

// PostingsList is a collection of docIDs. The interface only supports
// immutable methods.
type PostingsList interface {
	// Contains returns whether the specified id is contained in the set.
	Contains(id DocID) bool

	// IsEmpty returns whether the set is empty. Some posting lists have
	// an optimized implementation for determing if they are empty which
	// is faster than calculating the size of the postings list.
	IsEmpty() bool

	// Size returns the numbers of ids in the set.
	Size() uint64

	// Iterator returns an iterator over the known values.
	Iter() PostingsIterator

	// Clone returns a copy of the set.
	Clone() MutablePostingsList
}

// MutablePostingsList is a PostingsList implementation which also supports
// mutable operations.
type MutablePostingsList interface {
	PostingsList

	// Insert inserts the given id to set.
	Insert(i DocID) error

	// Intersect modifies the receiver set to contain only those ids by both sets.
	Intersect(other PostingsList) error

	// Difference modifies the receiver set to remove any ids contained by both sets.
	Difference(other PostingsList) error

	// Union modifies the receiver set to contain ids containted in either of the original sets.
	Union(other PostingsList) error

	// Reset resets the internal state of the PostingsList.
	Reset()
}

// PostingsIterator is an iterator over a set of docIDs.
type PostingsIterator interface {
	// Current returns the current DocID.
	Current() DocID

	// Next returns whether we have another DocID.
	Next() bool
}
