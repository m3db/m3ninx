// Copyright (c) 2017 Uber Technologies, Inc.
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

package mem

import (
	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"
)

type sequentialSearcher struct {
	queryable queryable
	pool      segment.PostingsListPool
}

// newSequentialSearcher returns a new sequential searcher.
func newSequentialSearcher(
	q queryable,
	pool segment.PostingsListPool,
) searcher {
	return &sequentialSearcher{
		queryable: q,
		pool:      pool,
	}
}

func (s *sequentialSearcher) Query(query segment.Query) ([]doc.Document, error) {
	// TODO: timeout/early termination once we know we're done

	var candidateDocIds segment.PostingsList
	// TODO: support parallel fetching across segments/filters
	for _, filter := range query.Filters {
		fetchedIds, _, err := s.queryable.Filter(filter.FieldName, filter.FieldValueFilter, filter.Regexp)
		if err != nil {
			return nil, err
		}

		// i.e. we don't have any documents for the given filter, can early terminate entire fn
		if fetchedIds == nil {
			return nil, nil
		}

		if candidateDocIds == nil {
			candidateDocIds = fetchedIds
			continue
		}

		// update candidate set
		// TODO: evaluate perf impact of retrieving all candidate docIDs, waiting till end,
		// sorting by size and then doing the intersection
		candidateDocIds.Intersect(fetchedIds)

		// early terminate if we don't have any docs in candidate set
		if candidateDocIds.IsEmpty() {
			return nil, nil
		}
	}

	// TODO: once we support multiple segments, we'll have to merge results
	return s.queryable.Fetch(candidateDocIds)
}
