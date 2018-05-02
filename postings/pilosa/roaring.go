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

package pilosa

import (
	"errors"
	"fmt"
	"sync"

	"github.com/m3db/m3ninx/postings"

	"github.com/pilosa/pilosa/roaring"
)

var (
	errIntersectRoaringOnly  = errors.New("Intersect only supported between roaringDocId sets")
	errUnionRoaringOnly      = errors.New("Union only supported between roaringDocId sets")
	errDifferenceRoaringOnly = errors.New("Difference only supported between roaringDocId sets")
	errIteratorClosed        = errors.New("iterator has been closed")
)

// postingsList wraps a Roaring Bitmap with a mutex for thread safety.
type postingsList struct {
	sync.RWMutex
	bitmap *roaring.Bitmap
}

// NewPostingsList returns a new mutable postings list backed by a Roaring Bitmap.
func NewPostingsList() postings.MutableList {
	return &postingsList{
		bitmap: roaring.NewBitmap(),
	}
}

func (d *postingsList) Insert(i postings.ID) {
	d.Lock()
	d.bitmap.Add(uint64(i))
	d.Unlock()
}

func (d *postingsList) Intersect(other postings.List) error {
	o, ok := other.(*postingsList)
	if !ok {
		return errIntersectRoaringOnly
	}

	o.RLock()
	d.Lock()
	d.bitmap = d.bitmap.Intersect(o.bitmap)
	d.Unlock()
	o.RUnlock()
	return nil
}

func (d *postingsList) Difference(other postings.List) error {
	o, ok := other.(*postingsList)
	if !ok {
		return errDifferenceRoaringOnly
	}

	o.RLock()
	d.Lock()
	d.bitmap = d.bitmap.Difference(o.bitmap)
	d.Unlock()
	o.RUnlock()
	return nil
}

func (d *postingsList) Union(other postings.List) error {
	o, ok := other.(*postingsList)
	if !ok {
		return errUnionRoaringOnly
	}

	o.RLock()
	d.Lock()
	d.bitmap = d.bitmap.Union(o.bitmap)
	d.Unlock()
	o.RUnlock()
	return nil
}

func (d *postingsList) RemoveRange(min, max postings.ID) {
	d.Lock()
	for i := min; i <= max; i++ {
		d.bitmap.Remove(uint64(i))
	}
	d.Unlock()
}

func (d *postingsList) Reset() {
	d.Lock()
	d.bitmap = roaring.NewBitmap()
	d.Unlock()
}

func (d *postingsList) Contains(i postings.ID) bool {
	d.RLock()
	contains := d.bitmap.Contains(uint64(i))
	d.RUnlock()
	return contains
}

func (d *postingsList) IsEmpty() bool {
	d.RLock()
	empty := d.bitmap.Max() == 0
	d.RUnlock()
	return empty
}

func (d *postingsList) Max() (postings.ID, error) {
	d.RLock()
	max := d.bitmap.Max()
	d.RUnlock()
	if max == 0 {
		return 0, fmt.Errorf("empty")
	}
	return postings.ID(max), nil
}

func (d *postingsList) Len() int {
	d.RLock()
	l := d.bitmap.Count()
	d.RUnlock()
	return int(l)
}

func (d *postingsList) Iterator() postings.Iterator {
	return &roaringIterator{
		iter: d.bitmap.Iterator(),
		max:  d.bitmap.Max(),
	}
}

func (d *postingsList) Clone() postings.MutableList {
	d.RLock()
	// TODO: It's cheaper to Clone than to cache roaring bitmaps, see
	// `postings_list_bench_test.go`. Their internals don't allow for
	// pooling at the moment. We should address this when get a chance
	// (move to another implementation / address deficiencies).
	clone := d.bitmap.Clone()
	d.RUnlock()
	return &postingsList{
		bitmap: clone,
	}
}

func (d *postingsList) Equal(other postings.List) bool {
	panic("not implemented")
	/*
		if d.Len() != other.Len() {
			return false
		}

		o, ok := other.(*postingsList)
		if ok {
			return d.bitmap.Equals(o.bitmap)
		}

		iter := d.Iterator()
		otherIter := other.Iterator()

		for iter.Next() {
			if !otherIter.Next() {
				return false
			}
			if iter.Current() != otherIter.Current() {
				return false
			}
		}

		return true
	*/
}

type roaringIterator struct {
	iter    *roaring.Iterator
	max     uint64
	current postings.ID
	closed  bool
}

func (it *roaringIterator) Current() postings.ID {
	return it.current
}

func (it *roaringIterator) Next() bool {
	if it.closed {
		return false
	}
	v, done := it.iter.Next()
	if done {
		it.closed = true
		return false
	}
	if v > it.max {
		it.closed = true
		return false
	}

	it.current = postings.ID(v)
	return true
}

func (it *roaringIterator) Err() error {
	return nil
}

func (it *roaringIterator) Close() error {
	if it.closed {
		return errIteratorClosed
	}
	it.closed = true
	return nil
}
