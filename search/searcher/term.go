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

type termSearcher struct {
	field, term []byte
	snapshot    index.Snapshot

	current doc.Document
	closed  bool
}

// NewTermSearcher returns a new Searcher for matching a term exactly.
func NewTermSearcher(s index.Snapshot, field, term []byte) (search.Searcher, error) {
	for _, r := range s.Readers() {
		pl, err := r.MatchTerm(field, term)
		if err != nil {

		}
	}

	return &termSearcher{
		field:    field,
		term:     term,
		snapshot: s,
	}
}

func (s *termSearcher) Next() (doc.Document, error) { return doc.Document{}, nil }
func (s *termSearcher) Close() error                { return nil }
