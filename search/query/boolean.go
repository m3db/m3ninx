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

// import (
// 	"errors"

// 	"github.com/m3db/m3ninx/index"
// 	"github.com/m3db/m3ninx/postings"
// 	"github.com/m3db/m3ninx/search"
// )

// var (
// 	errMustNotQuery = errors.New("queries which contain Must Not queries must contain at least one Must or Should query")
// )

// // BooleanQuery is a compound query composed of must, should, and must not queries.
// type BooleanQuery struct {
// 	must    search.Query
// 	mustNot search.Query
// 	should  search.Query
// }

// // NewBooleanQuery returns a new compound query such that:
// //   - resulting documents match ALL of the must queries
// //   - resulting documents match AT LEAST ONE of the should queries
// //   - resulting documents match NONE of the mustNot queries
// func NewBooleanQuery(
// 	must []search.Query, should []search.Query, mustNot []search.Query,
// ) (search.Query, error) {
// 	// Any queries which include a Must Not query must also include either a Must or Should
// 	// query to constrain the results. In the future this restriction may be removed.
// 	if len(mustNot) > 0 && len(must) == 0 && len(should) == 0 {
// 		return nil, errMustNotQuery
// 	}

// 	var q BooleanQuery

// 	// TODO(jeromefroe): Investigate rewriting the queries to improve performance. For
// 	// example, if the sub queries are themselves boolean queries we can flatten them.
// 	if len(must) > 0 {
// 		q.must = newConjuctionQuery(must)
// 	}
// 	if len(should) > 0 {
// 		q.should = newDisjuctionQuery(should)
// 	}
// 	if len(mustNot) > 0 {
// 		q.mustNot = newConjuctionQuery(mustNot)
// 	}

// 	return &q, nil
// }

// // Execute returns an iterator over documents matching the boolean query.
// func (q *BooleanQuery) Execute(r index.Reader) (postings.List, error) {
// 	var mpl postings.MutableList

// 	if q.must != nil {
// 		pl, err := q.must.Execute(r)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// Fast path for when no documents match the Must query.
// 		if pl.IsEmpty() {
// 			return pl, nil
// 		}

// 		mpl = pl.Clone()
// 	}

// 	if q.should != nil {
// 		pl, err := q.should.Execute(r)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if mpl == nil {
// 			mpl = pl.Clone()
// 		} else {
// 			mpl.Intersect(pl)
// 		}

// 		// Fast path for when no documents match both the Must and Should queries.
// 		if mpl.IsEmpty() {
// 			return mpl, nil
// 		}
// 	}

// 	if q.mustNot != nil {
// 		pl, err := q.mustNot.Execute(r)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// pl contains documents which match the Must Not query and which therefore need to be
// 		// removed from the current set of matching documents.
// 		mpl.Difference(pl)
// 	}

// 	return mpl, nil
// }
