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

package doc

// ID represents a document's natural identifier. It's treated as immutable
// by the indexing subsystem.
type ID []byte

// HashSize is the number bytes in the hash of an ID.
const HashSize = 16

// Hash is the hash of a []byte, safe to store in a map.
type Hash [HashSize]byte

// HashFn computes the Hash for a given document ID.
type HashFn func(id ID) Hash

// Field represents a field in a document. It is composed of a name and a value.
type Field struct {
	Name  []byte
	Value []byte
}

// Document represents a document to be indexed.
type Document struct {
	// ID is the natural identifier for the document.
	ID ID

	// Fields contains the list of fields by which to index the document.
	Fields []Field
}

// Iterator provides an iterator over a collection of documents.
type Iterator interface {
	// Next returns a bool indicating if the iterator has any more documents
	// to return.
	Next() bool

	// Current returns the current document.
	Current() Document

	// Err returns any errors encountered during iteration.
	Err() error

	// Error releases any internal resources used by the iterator.
	Close() error
}
