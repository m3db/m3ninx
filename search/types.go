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

package search

import (
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/postings"

	"github.com/m3db/m3x/errors"
)

// Query is a search query for documents.
type Query interface {
	// Searcher returns a Searcher for executing the query over a collection of Readers.
	Searcher(rs index.Readers) (Searcher, error)
}

// Searcher executes a query against a collection of Readers. It is an iterator which
// returns the postings lists of the documents it matches for each segment. It is not
// safe for concurrent access.
type Searcher interface {
	// Next returns the whether the iterator has another postings list.
	Next() bool

	// Current returns the current postings list. It is only safe to call Current immediately
	// after a call to Next confirms there are more postings lists remaining.
	Current() postings.List

	// Len returns the number of Readers that the Searcher is searching over.
	Len() int

	// Err returns any errors encountered during iteration.
	Err() error

	// Close closes the iterator.
	Close() error
}

// Searchers is a slice of Searcher.
type Searchers []Searcher

// Close closes all of the Searchers in ss.
func (ss Searchers) Close() error {
	multiErr := errors.NewMultiError()
	for _, s := range ss {
		err := s.Close()
		if err != nil {
			multiErr.Add(err)
		}
	}

	if multiErr.Empty() {
		return nil
	}
	return multiErr
}
