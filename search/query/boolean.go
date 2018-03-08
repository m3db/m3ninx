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

package query

import (
	"errors"

	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/search"
	"github.com/m3db/m3ninx/search/searcher"
)

var (
	errMustNotQuery = errors.New("queries which contain Must Not queries must contain at least one Must or Should query")
)

// booleanQuery is a compound query composed of must, should, and must not queries.
type booleanQuery struct {
	must    search.Query
	should  search.Query
	mustNot search.Query
}

// NewBooleanQuery returns a new compound query such that:
//   - resulting documents match ALL of the must queries
//   - resulting documents match AT LEAST ONE of the should queries
//   - resulting documents match NONE of the mustNot queries
func NewBooleanQuery(
	must []search.Query, should []search.Query, mustNot []search.Query,
) (search.Query, error) {
	// Any queries which include a MustNot query must also include either a Must or Should
	// query to constrain the results. In the future this restriction may be removed.
	if len(mustNot) > 0 && len(must) == 0 && len(should) == 0 {
		return nil, errMustNotQuery
	}

	var q booleanQuery
	if len(must) > 0 {
		q.must = newConjuctionQuery(must)
	}
	if len(should) > 0 {
		q.should = newDisjuctionQuery(should)
	}
	if len(mustNot) > 0 {
		q.mustNot = newConjuctionQuery(mustNot)
	}

	return &q, nil
}

func (q *booleanQuery) Searcher(rs index.Readers) (search.Searcher, error) {
	if q.must == nil && q.should == nil && q.mustNot == nil {
		l := len(rs)
		rs.Close() // Close the readers since the empty searcher does not take a reference to them.
		return searcher.NewEmptySearcher(l), nil
	}

	// If only must is non-empty then just return its Searcher directly.
	if q.must != nil && q.should == nil && q.mustNot == nil {
		return q.must.Searcher(rs)
	}
	// If only should is non-empty then just return its Searcher directly.
	if q.should != nil && q.must == nil && q.mustNot == nil {
		return q.should.Searcher(rs)
	}

	// Close the readers since we will pass a clone of them to each Searcher.
	defer rs.Close()

	// Construct must Searcher.
	clone, err := rs.Clone()
	if err != nil {
		return nil, err
	}
	must, err := q.must.Searcher(clone)
	if err != nil {
		return nil, err
	}

	// Construct should Searcher.
	clone, err = rs.Clone()
	if err != nil {
		return nil, err
	}
	should, err := q.should.Searcher(clone)
	if err != nil {
		return nil, err
	}

	// Construct mustNot Searcher.
	clone, err = rs.Clone()
	if err != nil {
		return nil, err
	}
	mustNot, err := q.mustNot.Searcher(clone)
	if err != nil {
		return nil, err
	}

	return searcher.NewBooleanSearcher(must, should, mustNot)
}
