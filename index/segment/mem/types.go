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

// termsDictionary is an internal interface for a mutable terms dictionary.
type termsDictionary interface {
	// Insert inserts the field with the given ID into the terms dictionary.
	Insert(field doc.Field, id postings.ID) error

	// MatchExact returns the postings list corresponding to documents which
	// match the given field and value exactly.
	MatchExact(field, value []byte) (postings.List, error)

	// MatchRegex returns the postings list corresponding to documents which
	// match the given field and regular expression pattern. Both the compiled
	// regular expression and the pattern it was generated from are provided
	// since some terms dictionaries can use the pattern to optimize their search.
	MatchRegex(field, pattern []byte, re *regexp.Regexp) (postings.List, error)
}
