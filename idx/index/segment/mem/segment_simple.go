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
	"errors"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"

	"github.com/uber-go/atomic"
)

var (
	errUnknownDocID = errors.New("unknown DocID specified")
)

// TODO(prateek): investigate impact of native heap
type simpleSegment struct {
	opts     Options
	id       doc.ID
	docIDGen *atomic.Uint32

	// internal docID -> document
	docs struct {
		sync.RWMutex
		values []document
	}

	// field (Name+Value) -> postingsManagerOffset
	termsDict *simpleTermsDictionary

	// TODO(prateek): add a delete documents bitmap to optimise fetch
}

type document struct {
	doc.Document
	docID      segment.DocID
	tombstoned bool
}

// New returns a new in-memory index.
func New(id doc.ID, opts Options) (segment.MutableSegment, error) {
	seg := &simpleSegment{
		opts:      opts,
		id:        id,
		docIDGen:  atomic.NewUint32(0),
		termsDict: newSimpleTermsDictionary(opts),
	}
	seg.docs.values = make([]document, opts.InitialCapacity())
	return seg, nil
}

// TODO(prateek): consider copy semantics for input data, esp if we can store it on the native heap
func (i *simpleSegment) Insert(d doc.Document) error {
	newDoc := document{
		Document: d,
		docID:    segment.DocID(i.docIDGen.Inc()),
	}
	i.insertDocument(newDoc)
	return i.insertTerms(newDoc)
}

func (i *simpleSegment) insertDocument(doc document) {
	docID := int64(doc.docID)
	// can early terminate if we have sufficient capacity
	i.docs.RLock()
	size := len(i.docs.values)
	if int64(size) > docID {
		// NB(prateek): only need a Read-lock here despite an insert operation because
		// we're guranteed to never have conflicts with docID (it's monotonoic increasing),
		// and have checked `i.docs.values` is large enough.
		i.docs.values[doc.docID] = doc
		i.docs.RUnlock()
		return
	}
	i.docs.RUnlock()

	// need to expand capacity
	i.docs.Lock()
	size = len(i.docs.values)
	// expanded since we released the lock
	if int64(size) > docID {
		i.docs.values[doc.docID] = doc
		i.docs.Unlock()
		return
	}

	docs := make([]document, 2*(size+1))
	copy(docs, i.docs.values)
	i.docs.values = docs
	i.docs.values[doc.docID] = doc
	i.docs.Unlock()
}

func (i *simpleSegment) insertTerms(doc document) error {
	// insert each of the indexed fields into the reverse index
	// TODO: current implementation allows for partial indexing. Evaluate perf impact of not doing that.
	for _, field := range doc.Fields {
		if err := i.termsDict.Insert(field, doc.docID); err != nil {
			return err
		}
	}

	return nil
}

func (i *simpleSegment) Delete(d doc.ID) error {
	panic("not implemented")
}

func (i *simpleSegment) MatchTerm(field, value string) (segment.PostingsList, error) {
	panic("not implemented")
}

func (i *simpleSegment) MatchRegex(field, pattern string) (segment.PostingsList, error) {
	panic("not implemented")
}

func (i *simpleSegment) Docs(pl segment.PostingsList, fields []string) (doc.Iterator, error) {
	panic("not implemented")
}
