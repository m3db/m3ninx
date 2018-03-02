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

package searcher

import (
	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/search"
)

type disjunctionSearcher struct {
	searchers []search.Searcher
	snapshot  index.Snapshot

	current doc.Document
	closed  bool
}

// NewDisjunctionSearcher returns a new Searcher for matching any.
func NewDisjunctionSearcher(ss []search.Searcher, s index.Snapshot) search.Searcher {
	return &disjunctionSearcher{
		searchers: ss,
		snapshot:  s,
	}
}

func (s *disjunctionSearcher) Next() bool            { return false }
func (s *disjunctionSearcher) Current() doc.Document { return doc.Document{} }
func (s *disjunctionSearcher) Err() error            { return nil }
func (s *disjunctionSearcher) Close() error          { return nil }
