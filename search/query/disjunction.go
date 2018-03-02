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
)

// disjuctionQuery finds documents which match at least one of the given queries.
type disjuctionQuery struct {
	queries []search.Query
}

// NewDisjuctionQuery constructs a new DisjuctionQuery from the given queries.
func NewDisjuctionQuery(queries []search.Query) search.Query {
	return &disjuctionQuery{
		queries: queries,
	}
}

func (q *disjuctionQuery) Searcher(s index.Snapshot) (search.Searcher, error) {
	return searcher.NewDisjunctionSearcher(queries s), nil
}
