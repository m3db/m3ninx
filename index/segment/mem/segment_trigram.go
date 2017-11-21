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
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"
	xlog "github.com/m3db/m3x/log"

	"github.com/uber-go/atomic"
)

type trigramSegment struct {
	opts     Options
	logger   xlog.Logger
	docIDGen *atomic.Uint32

	// internal docID -> document
	documentsLock sync.RWMutex
	documents     map[segment.DocID]document // TODO(prateek): measure perf impact of slice v map here

	// field (Name+Value) -> docIDs
	trigramTermsDict *trigramTermsDictionary

	searcher searcher
	// TODO(prateek): add a delete documents bitmap to optimise fetch
}

// NewTrigram returns a new trigram in-memory index segment.
func NewTrigram(opts Options) Segment {
	td := newTrigramTermsDictionary(opts)
	triTD := td.(*trigramTermsDictionary)
	seg := &trigramSegment{
		opts:     opts,
		logger:   opts.InstrumentOptions().Logger(),
		docIDGen: atomic.NewUint32(0),

		documents:        make(map[segment.DocID]document, opts.InitialCapacity()),
		trigramTermsDict: triTD,
	}

	searcher := newSequentialSearcher(seg, unionNegationFn, opts.PostingsListPool())
	seg.searcher = searcher
	return seg
}

func (i *trigramSegment) Insert(d doc.Document) error {
	return i.insertDocument(document{
		Document: d,
		docID:    segment.DocID(i.docIDGen.Inc()),
	})
}

func (i *trigramSegment) Query(query segment.Query) ([]doc.Document, error) {
	return i.searcher.Query(query)
}

func (i *trigramSegment) insertDocument(doc document) error {
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

func (i *trigramSegment) Filter(f segment.Filter) (segment.PostingsList, matchPredicate, error) {
	fn, err := filterMatcher(f).Fn()
	if err != nil {
		return nil, nil, err
	}
	docs, err := i.trigramTermsDict.Fetch(f.FieldName, f.FieldValueFilter, f.Regexp)
	if err != nil {
		return nil, nil, err
	}
	return docs, fn, nil
}

func (i *trigramSegment) Fetch(
	p segment.PostingsList,
	filterFn matchPredicate,
) ([]doc.Document, error) {
	var (
		numRetrieved = 0
		numFiltered  = 0
		docs         = make([]doc.Document, 0, p.Size())
		iter         = p.Iter()
	)

	// retrieve all the filtered document ids
	// TODO(prateek): can do this in parallel
	for ; iter.Next(); numRetrieved++ {
		id := iter.Current()
		d, ok := i.fetchDocument(id)
		if !ok {
			return nil, fmt.Errorf("unknown doc-id: %d", id)
		}
		if !filterFn(d) {
			numFiltered++
			continue
		}
		docs = append(docs, d)
	}
	i.logger.Debugf("query: %+v, retrieved %d, filtered %d", numRetrieved, numFiltered)
	// TODO(prateek): emit histogram for % of #filtered/#retrieved
	// TODO(prateek): emit histogram for % of #retrieved/#total-segment-cardinality
	// TODO(prateek): emit histogram for % of #(retrieved-filtered)/#total-segment-cardinality

	return docs, nil
}

func (i *trigramSegment) fetchDocument(id segment.DocID) (doc.Document, bool) {
	i.documentsLock.RLock()
	d, ok := i.documents[id]
	i.documentsLock.RUnlock()
	return d.Document, ok
}

func (i *trigramSegment) Size() uint32 {
	i.documentsLock.RLock()
	size := len(i.documents)
	i.documentsLock.RUnlock()
	return uint32(size)
}

func (i *trigramSegment) Update(d doc.Document) error {
	panic("not implemented")
}

func (i *trigramSegment) Delete(d doc.Document) error {
	panic("not implemented")
}

func (i *trigramSegment) Iter() segment.Iter {
	panic("not implemented")
}

func (i *trigramSegment) ID() segment.ID {
	panic("not implemented")
}

func (i *trigramSegment) Optimize() error {
	panic("not implemented")
}

// abstractions required to perform validation of results returned by
// trigram terms dictionary, as it's results are first class approximations.

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
