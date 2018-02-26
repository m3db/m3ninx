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

package mem

import (
	"errors"
	"regexp"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/postings"
)

var (
	errSegmentReaderClosed = errors.New("segment reader is closed")
)

type reader struct {
	sync.RWMutex

	segment readableSegment
	maxID   postings.ID
	wg      *sync.WaitGroup

	closed bool
}

func newReader(s readableSegment, maxID postings.ID, wg *sync.WaitGroup) index.Reader {
	wg.Add(1)
	return &reader{
		segment: s,
		maxID:   maxID,
		wg:      wg,
	}
}

func (r *reader) MatchExact(name, value []byte) (postings.List, error) {
	r.RLock()
	if r.closed {
		r.RUnlock()
		return nil, errSegmentReaderClosed
	}

	// A reader can return IDs in the posting list which are greater than its maximum
	// permitted ID. The reader only guarantees that when fetching the documents associated
	// with a postings list through a call to Docs will IDs greater than the maximum be
	// filtered out.
	pl, err := r.segment.matchExact(name, value)
	r.RUnlock()
	return pl, err
}

func (r *reader) MatchRegex(name, pattern []byte, re *regexp.Regexp) (postings.List, error) {
	r.RLock()
	if r.closed {
		r.RUnlock()
		return nil, errSegmentReaderClosed
	}

	// A reader can return IDs in the posting list which are greater than its maximum
	// permitted ID. The reader only guarantees that when fetching the documents associated
	// with a postings list through a call to Docs will IDs greater than the maximum be
	// filtered out.
	pl, err := r.segment.matchRegex(name, pattern, re)
	r.RUnlock()
	return pl, err
}

func (r *reader) Docs(pl postings.List, names [][]byte) (doc.Iterator, error) {
	// TODO: Add filter for names.
	if len(names) != 0 {
		panic("names filter is unimplemented")
	}

	if pl.IsEmpty() {
		return emptyIter, nil
	}

	max, err := pl.Max()
	if err != nil {
		return nil, err
	}

	// Remove any IDs from the postings list which are greater than the maximum ID
	// permitted by the reader.
	if max > r.maxID {
		mpl, ok := pl.(postings.MutableList)
		if !ok {
			mpl = pl.Clone()
		}
		mpl.RemoveRange(r.maxID, postings.MaxID)
		pl = mpl
	}

	return newIterator(r.segment, pl.Iterator(), r.wg), nil
}

func (r *reader) Close() error {
	r.Lock()
	if r.closed {
		r.Unlock()
		return errSegmentReaderClosed
	}
	r.closed = true
	r.Unlock()

	r.wg.Done()
	return nil
}
