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
	"fmt"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"

	"github.com/uber-go/atomic"
)

// TODO(prateek): investigate impact of native heap
type memIndex struct {
	opts     Options
	docIDGen *atomic.Uint32

	// internal docID -> document
	documentsLock sync.RWMutex
	documents     map[segment.DocID]document // TODO(prateek): measure perf impact of slice v map here

	// field (Name+Value) -> postingsManagerOffset
	termsDict termsDictionary

	searcher searcher
	// TODO(prateek): add a delete documents bitmap to optimise fetch
}

// New returns a new in-memory index.
func New(opts Options) (Segment, error) {
	seg := &memIndex{
		opts:     opts,
		docIDGen: atomic.NewUint32(0),

		documents: make(map[segment.DocID]document, opts.InitialCapacity()),
		termsDict: newSimpleTermsDictionary(opts, newPostingsManager),
	}

	searcher := newSequentialSearcher(seg, opts.PostingsListPool())
	seg.searcher = searcher
	return seg, nil
}

// TODO(prateek): consider copy semantics for input data, esp if we can store it on the native heap
func (i *memIndex) Insert(d doc.Document) error {
	return i.insertDocument(document{
		Document: d,
		docID:    segment.DocID(i.docIDGen.Inc()),
	})
}

func (i *memIndex) insertDocument(doc document) error {
	// insert document into master doc id -> doc map
	i.documentsLock.Lock()
	i.documents[doc.docID] = doc
	i.documentsLock.Unlock()

	// insert each of the indexed fields into the reverse index
	// TODO: current implementation allows for partial indexing. Evaluate perf impact of not doing that.
	for _, field := range doc.Fields {
		if err := i.termsDict.Insert(field, doc.docID); err != nil {
			return err
		}
	}

	return nil
}

func (i *memIndex) Query(query segment.Query) ([]doc.Document, error) {
	return i.searcher.Query(query)
}

func (i *memIndex) Filter(
	fieldName []byte,
	fieldValue []byte,
	isRegex bool,
) (segment.PostingsList, matchPredicate, error) {
	docs, err := i.termsDict.Fetch(fieldName, fieldValue, isRegex)
	return docs, nil, err
}

func (i *memIndex) Fetch(p segment.PostingsList) ([]doc.Document, error) {
	docs := make([]doc.Document, 0, p.Size())
	iter := p.Iter()

	// retrieve all the filtered document ids
	for iter.Next() {
		id := iter.Current()
		d, ok := i.fetchDocument(id)
		if !ok {
			return nil, fmt.Errorf("unknown doc-id: %d", id)
		}
		docs = append(docs, d.Document)
	}

	return docs, nil
}

func (i *memIndex) fetchDocument(id segment.DocID) (document, bool) {
	i.documentsLock.RLock()
	d, ok := i.documents[id]
	i.documentsLock.RUnlock()
	return d, ok
}

func (i *memIndex) Delete(d doc.Document) error {
	panic("not implemented")
}
func (i *memIndex) Size() uint32 {
	panic("not implemented")
}

func (i *memIndex) ID() segment.ID {
	panic("not implemented")
}

func (i *memIndex) Optimize() error {
	panic("not implemented")
}

func (i *memIndex) Iter() segment.Iter {
	return newMemIndexIter(i)
}

type memIndexIter struct {
	idx *memIndex

	current segment.DocID
	max     segment.DocID
}

func newMemIndexIter(idx *memIndex) segment.Iter {
	max := segment.DocID(idx.docIDGen.Load())
	return &memIndexIter{
		idx: idx,
		max: max,
	}
}

func (i *memIndexIter) Next() bool {
	i.current++
	return i.current <= i.max
}

func (i *memIndexIter) Current() (doc.Document, bool, segment.DocID) {
	d, ok := i.idx.fetchDocument(i.current)
	if !ok {
		return doc.Document{}, false, 0
	}
	return d.Document, d.tombstoned, i.current
}

func (i *memIndexIter) Err() error {
	if i.current > (i.max + 1) {
		return fmt.Errorf("iteration past valid index (current:%d, max:%d)", i.current, i.max)
	}
	return nil
}
