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

// FieldValueType represents the different types of field values supported.
type FieldValueType byte

const (
	// StringValueType demarks string value types.
	StringValueType FieldValueType = iota
)

// Term represents the value of a field.
type Term []byte

// Field represents a document field.
type Field struct {
	Name      []byte
	Value     Term
	ValueType FieldValueType
}

// Document represents a document to be indexed.
type Document struct {
	// ID is the natural identifier for the document.
	ID ID

	// IndexableFields contain the list of fields by which to index the document.
	IndexableFields []Field

	// StoredFields contains the list of fields to store supplemental
	// to the indexable fields.
	StoredFields []Field
}

// HashSize is the number of the hashed ID (currently set at 16 as we're
// using md5 Hashing).
const HashSize = 16

// Hash is the hash of a []byte, safe to store in a map.
type Hash [HashSize]byte

// HashFn computes the Hash for a given document ID.
type HashFn func(id ID) Hash
