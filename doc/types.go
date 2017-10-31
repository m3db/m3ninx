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

import "crypto/md5"

// ID represents a document's natural identifier.
type ID []byte

// FieldValueType represents the different types of field values supported.
type FieldValueType int

const (
	// UnknownValueType marks un-specified value types.
	UnknownValueType FieldValueType = iota

	// StringValueType demarks string value types.
	StringValueType
)

// Term represents the name/value of a field.
type Term []byte

// Field represents a document field.
type Field struct {
	Name      Term
	Value     Term
	ValueType FieldValueType
}

// Document represents a document to be indexed.
type Document interface {
	// ID is the natural identifier for the document.
	ID() ID

	// IndexFields contain the list of fields to index by.
	IndexFields() []Field

	// Bytes returns a serialized representation of the document. It may
	// comprise contents in addition to the ID/Index fields.
	Bytes() []byte
}

// Hash is the hash of a []byte, safe to store in a map.
type Hash [md5.Size]byte

// IDHashFn computes the Hash for a given document ID.
type IDHashFn func(id ID) Hash

// TermHashFn computes the Hash for a given term.
type TermHashFn func(t Term) Hash
