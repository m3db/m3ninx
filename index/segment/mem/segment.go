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
	re "regexp"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index"
	sgmt "github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/index/segment/mem/idsgen"
	"github.com/m3db/m3ninx/index/util"
	"github.com/m3db/m3ninx/postings"
)

var (
	errDuplicateID       = errors.New("a batch cannot contain duplicate IDs")
	errUnknownPostingsID = errors.New("unknown postings ID specified")
)

// nolint: maligned
type segment struct {
	offset    int
	newUUIDFn util.NewUUIDFn

	state struct {
		sync.RWMutex
		closed bool
	}

	// Mapping of postings ID to document.
	docs struct {
		sync.RWMutex
		data []doc.Document
	}

	// Mapping of term to postings list.
	termsDict termsDictionary

	writer struct {
		sync.Mutex
		idSet  *idsgen.Map
		nextID postings.ID
	}
	readerID postings.AtomicID
}

// NewSegment returns a new in-memory mutable segment. It will start assigning
// postings IDs at the provided offset.
func NewSegment(offset postings.ID, opts Options) (sgmt.MutableSegment, error) {
	s := &segment{
		offset:    int(offset),
		newUUIDFn: opts.NewUUIDFn(),
		termsDict: newTermsDict(opts),
		readerID:  postings.NewAtomicID(offset),
	}

	s.docs.data = make([]doc.Document, opts.InitialCapacity())

	s.writer.idSet = idsgen.New(256)
	s.writer.nextID = offset
	return s, nil
}

func (s *segment) Insert(d doc.Document) ([]byte, error) {
	s.state.RLock()
	defer s.state.RUnlock()
	if s.state.closed {
		return nil, sgmt.ErrClosed
	}

	{
		s.writer.Lock()

		b := index.NewBatch([]doc.Document{d})
		if err := s.prepareDocsWithLocks(b); err != nil {
			return nil, err
		}

		// Update the document in case we generated a UUID for it.
		d = b.Docs[0]

		s.insertDocWithLocks(d)
		s.readerID.Inc()

		s.writer.Unlock()
	}

	return d.ID, nil
}

func (s *segment) InsertBatch(b index.Batch) error {
	s.state.RLock()
	defer s.state.RUnlock()
	if s.state.closed {
		return sgmt.ErrClosed
	}

	var err error
	{
		s.writer.Lock()

		err = s.prepareDocsWithLocks(b)
		if err != nil && !index.IsBatchPartialError(err) {
			return err
		}

		for _, d := range b.Docs {
			s.insertDocWithLocks(d)
		}
		s.readerID.Add(uint32(len(b.Docs)))

		s.writer.Unlock()
	}

	return err
}

// prepareDocsWithLocks ensures the given documents can be inserted into the index. It
// must be called with the state and writer locks.
func (s *segment) prepareDocsWithLocks(b index.Batch) error {
	s.writer.idSet.Reset()

	batchErr := index.NewBatchPartialError()
	for i := 0; i < len(b.Docs); i++ {
		d := b.Docs[i]
		if err := d.Validate(); err != nil {
			if !b.AllowPartialUpdates {
				return err
			}
			batchErr.Add(err, i)
			continue
		}

		if d.HasID() {
			if s.termsDict.ContainsTerm(doc.IDReservedFieldName, d.ID) {
				// The segment already contains this document so we can remove it from those
				// we need to index.
				b.Docs[i], b.Docs[len(b.Docs)] = b.Docs[len(b.Docs)], b.Docs[i]
				b.Docs = b.Docs[:len(b.Docs)-1]

				// Decrement the loop variable since we just removed this document.
				i--
				continue
			}

			if _, ok := s.writer.idSet.Get(d.ID); ok {
				if !b.AllowPartialUpdates {
					return errDuplicateID
				}
				batchErr.Add(errDuplicateID, i)
				continue
			}
		} else {
			id, err := s.newUUIDFn()
			if err != nil {
				if !b.AllowPartialUpdates {
					return err
				}
				batchErr.Add(err, i)
				continue
			}

			d.ID = id

			// Update the document in the batch since we added an ID to it.
			b.Docs[i] = d
		}

		s.writer.idSet.SetUnsafe(d.ID, struct{}{}, idsgen.SetUnsafeOptions{
			NoCopyKey:     true,
			NoFinalizeKey: true,
		})
	}

	if batchErr.IsEmpty() {
		return nil
	}
	return batchErr
}

// insertDocWithLocks inserts a document into the index. It must be called with the
// state and writer locks.
func (s *segment) insertDocWithLocks(d doc.Document) {
	nextID := s.writer.nextID
	s.indexDocWithStateLock(nextID, d)
	s.storeDocWithStateLock(nextID, d)
	s.writer.nextID++
}

// indexDocWithStateLock indexes the fields of a document in the segment's terms
// dictionary. It must be called with the segment's state lock.
func (s *segment) indexDocWithStateLock(id postings.ID, d doc.Document) error {
	for _, f := range d.Fields {
		s.termsDict.Insert(f, id)
	}
	s.termsDict.Insert(doc.Field{
		Name:  doc.IDReservedFieldName,
		Value: d.ID,
	}, id)
	return nil
}

// storeDocWithStateLock stores a documents into the segment's mapping of postings
// IDs to documents. It must be called with the segment's state lock.
func (s *segment) storeDocWithStateLock(id postings.ID, d doc.Document) {
	idx := int(id) - s.offset

	// Can return early if we have sufficient capacity.
	{
		s.docs.RLock()
		size := len(s.docs.data)
		if size > idx {
			// NB(prateek): We only need a Read-lock here despite an insert operation because
			// we're guaranteed to never have conflicts with docID (it's monotonically increasing),
			// and have checked `i.docs.data` is large enough.
			s.docs.data[idx] = d
			s.docs.RUnlock()
			return
		}
		s.docs.RUnlock()
	}

	// Otherwise we need to expand capacity.
	{
		s.docs.Lock()
		size := len(s.docs.data)

		// The slice has already been expanded since we released the lock.
		if size > idx {
			s.docs.data[idx] = d
			s.docs.Unlock()
			return
		}

		data := make([]doc.Document, 2*(size+1))
		copy(data, s.docs.data)
		s.docs.data = data
		s.docs.data[idx] = d
		s.docs.Unlock()
	}
}

func (s *segment) Reader() (index.Reader, error) {
	s.state.RLock()
	defer s.state.RUnlock()
	if s.state.closed {
		return nil, sgmt.ErrClosed
	}

	limit := s.readerID.Load()
	r := newReader(s, limit)
	return r, nil
}

func (s *segment) matchTerm(field, term []byte) (postings.List, error) {
	s.state.RLock()
	defer s.state.RUnlock()
	if s.state.closed {
		return nil, sgmt.ErrClosed
	}

	return s.termsDict.MatchTerm(field, term), nil
}

func (s *segment) matchRegexp(name, regexp []byte, compiled *re.Regexp) (postings.List, error) {
	s.state.RLock()
	defer s.state.RUnlock()
	if s.state.closed {
		return nil, sgmt.ErrClosed
	}

	if compiled == nil {
		var err error
		compiled, err = re.Compile(string(regexp))
		if err != nil {
			return nil, err
		}
	}
	return s.termsDict.MatchRegexp(name, regexp, compiled), nil
}

func (s *segment) getDoc(id postings.ID) (doc.Document, error) {
	s.state.RLock()
	defer s.state.RUnlock()
	if s.state.closed {
		return doc.Document{}, sgmt.ErrClosed
	}

	idx := int(id) - s.offset

	s.docs.RLock()
	if idx >= len(s.docs.data) {
		s.docs.RUnlock()
		return doc.Document{}, errUnknownPostingsID
	}
	d := s.docs.data[idx]
	s.docs.RUnlock()

	return d, nil
}

func (s *segment) Close() error {
	s.state.Lock()
	defer s.state.Unlock()
	if s.state.closed {
		return sgmt.ErrClosed
	}

	s.state.closed = true
	return nil
}
