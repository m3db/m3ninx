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
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/search"
	"github.com/m3db/m3ninx/search/searcher"
)

// conjuctionQuery finds documents which match at least one of the given queries.
type conjuctionQuery struct {
	queries []search.Query
}

// newConjuctionQuery constructs a new query which matches documents which match all
// of the given queries.
func newConjuctionQuery(queries []search.Query) search.Query {
	qs := make([]search.Query, 0, len(queries))
	for _, query := range queries {
		// Merge conjunction queries into slice of top-level queries.
		q, ok := query.(*conjuctionQuery)
		if ok {
			qs = append(qs, q.queries...)
			continue
		}

		qs = append(qs, query)
	}
	return &conjuctionQuery{
		queries: qs,
	}
}

func (q *conjuctionQuery) Searcher(rs index.Readers) (search.Searcher, error) {
	switch l := len(q.queries); l {
	case 0:
		l := len(rs)

		// Close the readers since the empty searcher does not take a reference to them.
		rs.Close()

		return searcher.NewEmptySearcher(l), nil

	case 1:
		// If there is only a single query we can return the Searcher for just that query.
		return q.queries[0].Searcher(rs)

	default:
	}

	ss := make(search.Searchers, 0, len(q.queries))
	for _, q := range q.queries {
		clone, err := rs.Clone()
		if err != nil {
			ss.Close()
			return nil, err
		}

		s, err := q.Searcher(clone)
		if err != nil {
			ss.Close()
			return nil, err
		}
		ss = append(ss, s)
	}

	// Close the readers since we pass a clone of them to each of the internal Searchers.
	rs.Close()

	return searcher.NewConjunctionSearcher(ss)
}
