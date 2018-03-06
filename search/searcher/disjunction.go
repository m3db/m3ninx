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
	"github.com/m3db/m3ninx/postings"
	"github.com/m3db/m3ninx/search"
)

type disjunctionSearcher struct {
	searchers search.Searchers
	len       int

	idx  int
	curr postings.List

	closed bool
	err    error
}

// NewDisjunctionSearcher returns a new Searcher which matches documents which are matched
// by any of the given Searchers. It is not safe for concurrent access.
func NewDisjunctionSearcher(ss search.Searchers) (search.Searcher, error) {
	var length int
	if len(ss) > 0 {
		length = ss[0].Len()
		for _, s := range ss[1:] {
			if s.Len() != length {
				ss.Close()
				return nil, errSearchersLengthsUnequal
			}
		}
	}

	return &disjunctionSearcher{
		searchers: ss,
		len:       length,
		idx:       -1,
	}, nil
}

func (s *disjunctionSearcher) Next() bool {
	if s.closed || s.err != nil || s.idx == s.len-1 {
		return false
	}

	var pl postings.MutableList
	s.idx++
	for _, sr := range s.searchers {
		if !sr.Next() {
			err := sr.Err()
			if err == nil {
				err = errSearcherTooShort
			}
			s.err = err
			return false
		}

		// TODO: Sort the iterators so that we take the union in order of decreasing size.
		curr := sr.Current()
		if pl == nil {
			pl = curr.Clone()
		} else {
			pl.Union(curr)
		}
	}
	s.curr = pl

	return true
}

func (s *disjunctionSearcher) Current() postings.List {
	return s.curr
}

func (s *disjunctionSearcher) Len() int {
	return s.len
}

func (s *disjunctionSearcher) Err() error {
	return s.err
}

func (s *disjunctionSearcher) Close() error {
	if s.closed {
		return errSearcherClosed
	}
	s.closed = true
	return s.searchers.Close()
}
