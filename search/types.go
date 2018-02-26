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
	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/postings"
)

// Query is a search query for documents.
type Query interface {
	// Execute executes the query against the given reader and returns a postings
	// list of documents matching the query.
	Execute(r index.Reader) (postings.List, error)
}

// Searcher executes a query against a snapshot.
type Searcher interface {
	// Search executes a query against the given snapshot and returns an iterator
	// over the documents matching the query.
	Search(s index.Snapshot, q Query) (doc.Iterator, error)
}
