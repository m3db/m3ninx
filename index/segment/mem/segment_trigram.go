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
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"
	xlog "github.com/m3db/m3x/log"

	"github.com/uber-go/atomic"
)

type trigramIndex struct {
	opts     Options
	logger   xlog.Logger
	docIDGen *atomic.Uint32

	// internal docID -> document
	documentsLock sync.RWMutex
	documents     map[segment.DocID]document // TODO(prateek): measure perf impact of slice v map here

	// field (Name+Value) -> postingOffset
	trigramTermsDict *trigramTermsDictionary

	// TODO(prateek): add a delete documents bitmap to optimise fetch
}

// NewTrigram returns a new trigram in-memory index segment.
func NewTrigram(opts Options) Segment {
	td := newTrigramTermsDictionary(opts, newPostingsManager)
	triTD := td.(*trigramTermsDictionary)
	return &trigramIndex{
		opts:     opts,
		logger:   opts.InstrumentOptions().Logger(),
		docIDGen: atomic.NewUint32(0),

		documents:        make(map[segment.DocID]document, opts.InitialCapacity()),
		trigramTermsDict: triTD,
	}
}

func (i *trigramIndex) Insert(d doc.Document) error {
	return i.insertDocument(document{
		Document: d,
		docID:    segment.DocID(i.docIDGen.Inc()),
	})
}

func (i *trigramIndex) insertDocument(doc document) error {
	// insert document into master doc id -> doc map
	i.documentsLock.Lock()
	i.documents[doc.docID] = doc
	i.documentsLock.Unlock()

	// insert each of the indexed fields into the reverse index
	for _, field := range doc.Fields {
		if err := i.trigramTermsDict.Insert(field, doc.docID); err != nil {
			return err
		}
	}
	return nil
}

func (i *trigramIndex) Query(query segment.Query) ([]doc.Document, error) {
	if err := validateQuery(query); err != nil {
		return nil, err
	}

	// compile validation function
	matchPredicate, err := queryMatcher(query).Fn()
	if err != nil {
		return nil, err
	}

	// order filters to ensure the first filter has no-negation
	filters := orderFiltersByNonNegated(query.Filters)
	sort.Sort(filters)

	var (
		candidateDocIds segment.PostingsList
	)

	// TODO(prateek): option to do this in parallel
	// TODO(prateek): a filter results cache of some kind
	for _, filter := range filters {
		filter := filter

		// TODO(prateek): can pool fetchedIds ...
		fetchedIds, err := i.trigramTermsDict.Fetch(
			filter.FieldName, filter.FieldValueFilter, filter.Regexp)
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

		if filter.Negate {
			// NB(prateek): unfortunately, we cannot use the trigram filtering to negate responses
			// because they're unable to authoratatively test if a query matches something or not.
			// So we have to union here.
			candidateDocIds.Union(fetchedIds)
		} else {
			candidateDocIds.Intersect(fetchedIds)
		}

		// early terminate if we don't have any docs in candidate set
		if candidateDocIds.IsEmpty() {
			return nil, nil
		}
	}

	var (
		numRetrieved = 0
		numFiltered  = 0
		docs         = make([]doc.Document, 0, candidateDocIds.Size())
		iter         = candidateDocIds.Iter()
	)

	// retrieve all the filtered document ids
	// TODO(prateek): can do this in parallel
	for ; iter.Next(); numRetrieved++ {
		id := iter.Current()
		d, ok := i.fetchDocument(id)
		if !ok {
			continue
		}
		if !matchPredicate(d) {
			numFiltered++
			continue
		}
		docs = append(docs, d)
	}
	i.logger.Debugf("query: %+v, retrieved %d, filtered %d", query, numRetrieved, numFiltered)
	// TODO(prateek): emit histogram for % of #filtered/#retrieved
	// TODO(prateek): emit histogram for % of #retrieved/#total-segment-cardinality
	// TODO(prateek): emit histogram for % of #(retrieved-filtered)/#total-segment-cardinality

	return docs, nil
}

func (i *trigramIndex) fetchDocument(id segment.DocID) (doc.Document, bool) {
	i.documentsLock.RLock()
	d, ok := i.documents[id]
	i.documentsLock.RUnlock()
	return d.Document, ok
}

func (i *trigramIndex) Update(d doc.Document) error {
	panic("not implemented")
}

func (i *trigramIndex) Delete(d doc.Document) error {
	panic("not implemented")
}

func (i *trigramIndex) Iter() segment.Iter {
	panic("not implemented")
}

func (i *trigramIndex) Size() uint32 {
	panic("not implemented")
}

func (i *trigramIndex) ID() segment.ID {
	panic("not implemented")
}

func (i *trigramIndex) Optimize() error {
	panic("not implemented")
}

// abstractions required to perform validation of results returned by
// trigram terms dictionary, as it's results are first class approximations.

type queryMatcher segment.Query

// Fn generates a matchPredicate for the given query.
func (q queryMatcher) Fn() (matchPredicate, error) {
	if q.Conjunction != segment.AndConjunction {
		return nil, fmt.Errorf("unsupport conjuction")
	}

	fns := make([]matchPredicate, 0, len(q.Filters)+len(q.SubQueries))
	for _, f := range q.Filters {
		fn, err := filterMatcher(f).Fn()
		if err != nil {
			return nil, err
		}
		fns = append(fns, fn)
	}

	// compose earlier generated fns
	return func(d doc.Document) bool {
		for _, fn := range fns {
			if !fn(d) {
				return false
			}
		}
		return true
	}, nil
}

type filterMatcher segment.Filter

// Fn generates a matchPredicate for the given filter.
func (f filterMatcher) Fn() (matchPredicate, error) {
	if !f.Regexp {
		return func(d doc.Document) bool {
			for _, field := range d.Fields {
				// find field with correct name
				if bytes.Compare(f.FieldName, field.Name) == 0 {
					if f.Negate {
						// ensure field value does not match
						return bytes.Compare(field.Value, f.FieldValueFilter) != 0
					}
					// ensure field value does matches
					return bytes.Compare(field.Value, f.FieldValueFilter) == 0
				}
			}
			// i.e. didn't find field, so we should allow this doc based on the filter negation
			return f.Negate
		}, nil
	}

	re, err := regexp.Compile(string(f.FieldValueFilter))
	if err != nil {
		return nil, err
	}
	return func(d doc.Document) bool {
		for _, field := range d.Fields {
			// find field with correct name
			if bytes.Compare(f.FieldName, field.Name) == 0 {
				if f.Negate {
					// ensure field value does not match
					return !re.Match(field.Value)
				}
				// ensure field value does matches
				return re.Match(field.Value)
			}
		}
		// i.e. didn't find field, so we should allow this doc based on the filter negation
		return f.Negate
	}, nil
}
