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
	"regexp"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index"
	sgmt "github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/postings"
)

var (
	errSegmentClosed     = errors.New("segment is closed")
	errSegmentSealed     = errors.New("segment has been sealed")
	errUnknownPostingsID = errors.New("unknown postings ID specified")
)

// TODO(prateek): investigate impact of native heap
type segment struct {
	sync.RWMutex

	opts Options

	idGenerator postings.IDGenerator
	wg          sync.WaitGroup
	sealed      bool
	closed      bool

	// internal postings ID -> document
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
	postingsID postings.ID
}

// NewSegment returns a new in-memory mutable segment.
func NewSegment(g postings.IDGenerator, opts Options) (sgmt.MutableSegment, error) {
	s := &segment{
		opts:        opts,
		idGenerator: g,
		termsDict:   newSimpleTermsDictionary(opts),
	}
	s.docs.values = make([]document, opts.InitialCapacity())
	return s, nil
}

// TODO(prateek): consider copy semantics for input data, esp if we can store it on the native heap
func (s *segment) Insert(d doc.Document) error {
	s.RLock()
	if s.closed {
		s.RUnlock()
		return errSegmentClosed
	}

	if s.sealed {
		s.RUnlock()
		return errSegmentSealed
	}

	newDoc := document{
		Document:   d,
		postingsID: postings.ID(s.idGenerator.Next()),
	}
	s.insertDoc(newDoc)
	err := s.insertTerms(newDoc)

	s.RUnlock()
	return err
}

func (s *segment) Reader() (index.Reader, error) {
	s.RLock()
	if s.closed {
		s.RUnlock()
		return nil, errSegmentClosed
	}

	s.wg.Add(1)

	r := &reader{
		segment: s,
		maxID:   postings.ID(s.idGenerator.Current()),
	}
	return r, nil
}

func (s *segment) Seal() error {
	s.Lock()
	if s.sealed {
		s.Unlock()
		return errSegmentSealed
	}

	s.sealed = true
	s.Unlock()
	return nil
}

func (s *segment) Close() error {
	s.Lock()
	if s.closed {
		s.Unlock()
		return errSegmentClosed
	}

	s.sealed = true
	s.closed = true
	s.Unlock()

	// Wait for all readers to be closed.
	s.wg.Wait()
	return nil
}

func (s *segment) insertDoc(doc document) {
	postingsID := int64(doc.postingsID)

	// Can terminate early if we have sufficient capacity.
	s.docs.RLock()
	size := len(s.docs.values)
	if int64(size) > postingsID {
		// NB(prateek): only need a Read-lock here despite an insert operation because
		// we're guranteed to never have conflicts with docID (it's monotonoic increasing),
		// and have checked `i.docs.values` is large enough.
		s.docs.values[doc.postingsID] = doc
		s.docs.RUnlock()
		return
	}
	s.docs.RUnlock()

	// Need to expand capacity.
	s.docs.Lock()
	size = len(s.docs.values)
	// The size has been expanded since we released the lock.
	if int64(size) > postingsID {
		s.docs.values[doc.postingsID] = doc
		s.docs.Unlock()
		return
	}

	docs := make([]document, 2*(size+1))
	copy(docs, s.docs.values)
	s.docs.values = docs
	s.docs.values[doc.postingsID] = doc
	s.docs.Unlock()
}

func (s *segment) insertTerms(doc document) error {
	// insert each of the indexed fields into the reverse index
	// TODO: current implementation allows for partial indexing. Evaluate perf impact of not doing that.
	for _, field := range doc.Fields {
		if err := s.termsDict.Insert(field, doc.postingsID); err != nil {
			return err
		}
	}

	return nil
}

func (s *segment) matchExact(field, value []byte) (postings.List, error) {
	return s.termsDict.MatchExact(field, value)
}

func (s *segment) matchRegex(field, pattern []byte, re *regexp.Regexp) (postings.List, error) {
	if re == nil {
		var err error
		re, err = regexp.Compile(string(pattern))
		if err != nil {
			return nil, err
		}
	}
	return s.termsDict.MatchRegex(field, pattern, re)
}

func (s *segment) getDocs(pl postings.List, fields [][]byte) (doc.Iterator, error) {
	panic("not implemented")
}
