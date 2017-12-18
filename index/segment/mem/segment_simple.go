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
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"

	"github.com/uber-go/atomic"
)

// TODO(prateek): investigate impact of native heap
type simpleSegment struct {
	opts     Options
	id       segment.ID
	docIDGen *atomic.Uint32

	// internal docID -> document
	docs struct {
		sync.RWMutex
		values map[segment.DocID]document // TODO(prateek): measure perf impact of slice v map here
	}

	// field (Name+Value) -> postingsManagerOffset
	termsDict termsDictionary

	searcher searcher
	// TODO(prateek): add a delete documents bitmap to optimise fetch
}

type document struct {
	doc.Document
	docID      segment.DocID
	tombstoned bool
}

// New returns a new in-memory index.
func New(id segment.ID, opts Options) (Segment, error) {
	seg := &simpleSegment{
		opts:      opts,
		id:        id,
		docIDGen:  atomic.NewUint32(0),
		termsDict: newSimpleTermsDictionary(opts),
	}
	seg.docs.values = make(map[segment.DocID]document, opts.InitialCapacity())
	searcher := newSequentialSearcher(seg, differenceNegationFn, opts.PostingsListPool())
	seg.searcher = searcher
	return seg, nil
}

// TODO(prateek): consider copy semantics for input data, esp if we can store it on the native heap
func (i *simpleSegment) Insert(d doc.Document) error {
	return i.insertDocument(document{
		Document: d,
		docID:    segment.DocID(i.docIDGen.Inc()),
	})
}

func (i *simpleSegment) insertDocument(doc document) error {
	// insert document into master doc id -> doc map
	i.docs.Lock()
	i.docs.values[doc.docID] = doc
	i.docs.Unlock()

	// insert each of the indexed fields into the reverse index
	// TODO: current implementation allows for partial indexing. Evaluate perf impact of not doing that.
	for _, field := range doc.Fields {
		if err := i.termsDict.Insert(field, doc.docID); err != nil {
			return err
		}
	}

	return nil
}

func (i *simpleSegment) Query(query segment.Query) (segment.ResultsIter, error) {
	ids, pendingFn, err := i.searcher.Query(query)
	if err != nil {
		return nil, err
	}

	return newResultsIter(ids, pendingFn, i), nil
}

func (i *simpleSegment) Filter(
	f segment.Filter,
) (segment.PostingsList, matchPredicate, error) {
	docs, err := i.termsDict.Fetch(f.FieldName, f.FieldValueFilter, f.Regexp)
	return docs, nil, err
}

func (i *simpleSegment) Delete(d doc.Document) error {
	panic("not implemented")
}

func (i *simpleSegment) Size() uint32 {
	panic("not implemented")
}

func (i *simpleSegment) ID() segment.ID {
	return i.id
}

func (i *simpleSegment) Options() Options {
	return i.opts
}

func (i *simpleSegment) FetchDocument(id segment.DocID) (document, bool) {
	i.docs.RLock()
	d, ok := i.docs.values[id]
	i.docs.RUnlock()
	return d, ok
}
