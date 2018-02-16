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
	"github.com/m3db/m3ninx/idx/index"
	"github.com/m3db/m3ninx/idx/search"
)

// disjuctionQuery finds documents which match at least one of the given queries.
type disjuctionQuery struct {
	queries []search.Query
}

// newDisjuctionQuery constructs a new DisjuctionQuery from the given queries.
func newDisjuctionQuery(queries []search.Query) search.Query {
	return &disjuctionQuery{
		queries: queries,
	}
}

// Execute returns an iterator over documents matching the disjunction query.
func (q *disjuctionQuery) Execute(r index.Reader) (index.PostingsList, error) {
	if len(q.queries) == 0 {
		return nil, nil
	}

	pl, err := q.queries[0].Execute(r)
	if err != nil {
		return nil, err
	}

	// Fast path for when we have only one internal query so we can avoid cloning the
	// postings list.
	if len(q.queries) == 1 {
		return pl, nil
	}

	mpl := pl.Clone()
	for _, qy := range q.queries[1:] {
		pl, err := qy.Execute(r)
		if err != nil {
			return nil, err
		}

		// TODO(jeromefroe): Instead of unioning in order we should instead get all the
		// postings lists, clone the largest one, and then union them in order of
		// decreasing size.
		mpl.Union(pl)
	}

	return mpl, nil
}
