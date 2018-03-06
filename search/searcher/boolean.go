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
	"errors"

	"github.com/m3db/m3ninx/postings"
	"github.com/m3db/m3ninx/search"
)

var (
	errMustNotSearcher = errors.New("a boolean searcher which contains a Must Not searcher must also contain either a Must or Should searcher")
)

type booleanSearcher struct {
	must    search.Searcher
	should  search.Searcher
	mustNot search.Searcher
	len     int

	idx  int
	curr postings.List

	closed bool
	err    error
}

// NewBooleanSearcher returns a new Searcher for exectuting boolean queries. It is not
// safe for concurrent access. If any of the Searchers are nil they are ignored when
// finding matching documents.
func NewBooleanSearcher(must, should, mustNot search.Searcher) (search.Searcher, error) {
	// Any queries which include a MustNot query must also include either a Must or Should
	// query to constrain the results. In the future this restriction may be removed.
	if mustNot != nil && must == nil && should == nil {
		return nil, errMustNotSearcher
	}

	l := -1
	if must != nil {
		l = must.Len()
	}

	if should != nil {
		if l == -1 {
			l = should.Len()
		} else if l != should.Len() {
			search.Searchers{must, mustNot, should}.Close()
			return nil, errSearchersLengthsUnequal
		}
	}

	// l must be defined when mustNot is not nil since we require must or should be defined
	// when mustNot is defined.
	if mustNot != nil && l != mustNot.Len() {
		search.Searchers{must, mustNot, should}.Close()
		return nil, errSearchersLengthsUnequal
	}

	return &booleanSearcher{
		must:    must,
		should:  should,
		mustNot: mustNot,
		len:     l,
		idx:     -1,
	}, nil
}

func (s *booleanSearcher) Next() bool {
	if s.closed || s.err != nil || s.idx == s.len-1 {
		return false
	}

	var mpl postings.MutableList
	s.idx++
	if s.must != nil {
		if !s.must.Next() {
			err := s.must.Err()
			if err == nil {
				err = errSearcherTooShort
			}
			s.err = err
			return false
		}
		pl := s.must.Current()

		// Fast path for when no documents match the Must query.
		if pl.IsEmpty() {
			s.curr = pl
			return true
		}

		mpl = pl.Clone()
	}

	if s.should != nil {
		if !s.should.Next() {
			err := s.should.Err()
			if err == nil {
				err = errSearcherTooShort
			}
			s.err = err
			return false
		}
		pl := s.should.Current()

		if mpl == nil {
			mpl = pl.Clone()
		} else {
			mpl.Intersect(pl)
		}

		// Fast path for when no documents match both the Must and Should queries.
		if mpl.IsEmpty() {
			s.curr = pl
			return true
		}
	}

	if s.mustNot != nil {
		if !s.mustNot.Next() {
			err := s.mustNot.Err()
			if err == nil {
				err = errSearcherTooShort
			}
			s.err = err
			return false
		}
		pl := s.mustNot.Current()

		// We require that either the Must or Should query is not nil so we are ensured that
		// mpl is not nil here.
		mpl.Difference(pl)
	}

	s.curr = mpl
	return true
}

func (s *booleanSearcher) Current() postings.List {
	return s.curr
}

func (s *booleanSearcher) Len() int {
	return s.len
}

func (s *booleanSearcher) Err() error {
	return s.err
}

func (s *booleanSearcher) Close() error {
	if s.closed {
		return errSearcherClosed
	}
	s.closed = true
	return search.Searchers{s.must, s.should, s.mustNot}.Close()
}
