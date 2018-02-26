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

package mem

import (
	"regexp"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/postings"
)

// termsDict is an internal interface for a mutable terms dictionary.
type termsDict interface {
	// Insert inserts the field with the given ID into the terms dictionary.
	Insert(field doc.Field, id postings.ID) error

	// MatchExact returns the postings list corresponding to documents which match the
	// given field name and value exactly.
	MatchExact(name, value []byte) (postings.List, error)

	// MatchRegex returns the postings list corresponding to documents which match the
	// given name field and regular expression pattern. Both the compiled regular expression
	// and the pattern it was generated from are provided so terms dictionaries can
	// optimize their searches.
	MatchRegex(name, pattern []byte, re *regexp.Regexp) (postings.List, error)
}

// readableSegment is an internal interface for reading from a segment.
type readableSegment interface {
	// matchExact returns the postings list of documents which match the given name and
	// value exactly.
	matchExact(name, value []byte) (postings.List, error)

	// matchRegex returns the postings list of documents which match the given name and
	// regular expression.
	matchRegex(name, pattern []byte, re *regexp.Regexp) (postings.List, error)

	// getDoc returns the document associated with the given ID.
	getDoc(id postings.ID) (doc.Document, error)
}
