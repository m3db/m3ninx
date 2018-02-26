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

package postings

import (
	"errors"
	"sync"

	"github.com/RoaringBitmap/roaring"
)

var (
	errIntersectRoaringOnly  = errors.New("Intersect only supported between roaringDocId sets")
	errUnionRoaringOnly      = errors.New("Union only supported between roaringDocId sets")
	errDifferenceRoaringOnly = errors.New("Difference only supported between roaringDocId sets")
	errIteratorClosed        = errors.New("iterator has been closed")
)

// roaringPostingsList wraps a Roaring Bitmap with a mutex for thread safety.
type roaringPostingsList struct {
	sync.RWMutex
	bitmap *roaring.Bitmap
}

// NewRoaringPostingsList returns a new mutable postings list backed by a Roaring Bitmap.
func NewRoaringPostingsList() MutableList {
	return &roaringPostingsList{
		bitmap: roaring.NewBitmap(),
	}
}

func (d *roaringPostingsList) Insert(i ID) error {
	d.Lock()
	d.bitmap.Add(uint32(i))
	d.Unlock()

	return nil
}

func (d *roaringPostingsList) Intersect(other List) error {
	o, ok := other.(*roaringPostingsList)
	if !ok {
		return errIntersectRoaringOnly
	}

	o.RLock()
	d.Lock()
	d.bitmap.And(o.bitmap)
	d.Unlock()
	o.RUnlock()
	return nil
}

func (d *roaringPostingsList) Difference(other List) error {
	o, ok := other.(*roaringPostingsList)
	if !ok {
		return errDifferenceRoaringOnly
	}

	d.Lock()
	o.RLock()
	d.bitmap.AndNot(o.bitmap)
	o.RUnlock()
	d.Unlock()
	return nil
}

func (d *roaringPostingsList) Union(other List) error {
	o, ok := other.(*roaringPostingsList)
	if !ok {
		return errUnionRoaringOnly
	}

	o.RLock()
	d.Lock()
	d.bitmap.Or(o.bitmap)
	d.Unlock()
	o.RUnlock()
	return nil
}

func (d *roaringPostingsList) RemoveRange(min, max ID) error {
	d.Lock()
	d.bitmap.RemoveRange(uint64(min), uint64(max))
	d.Unlock()
	return nil
}

func (d *roaringPostingsList) Reset() {
	d.Lock()
	d.bitmap.Clear()
	d.Unlock()
}

func (d *roaringPostingsList) Contains(i ID) bool {
	d.RLock()
	contains := d.bitmap.Contains(uint32(i))
	d.RUnlock()
	return contains
}

func (d *roaringPostingsList) IsEmpty() bool {
	d.RLock()
	empty := d.bitmap.IsEmpty()
	d.RUnlock()
	return empty
}

func (d *roaringPostingsList) Max() (ID, error) {
	d.RLock()
	if d.bitmap.IsEmpty() {
		d.RUnlock()
		return 0, ErrEmptyList
	}
	max := d.bitmap.Maximum()
	d.RUnlock()
	return ID(max), nil
}

func (d *roaringPostingsList) Min() (ID, error) {
	d.RLock()
	if d.bitmap.IsEmpty() {
		d.RUnlock()
		return 0, ErrEmptyList
	}
	min := d.bitmap.Minimum()
	d.RUnlock()
	return ID(min), nil
}

func (d *roaringPostingsList) Len() int {
	d.RLock()
	l := d.bitmap.GetCardinality()
	d.RUnlock()
	return int(l)
}

func (d *roaringPostingsList) Iterator() Iterator {
	return &roaringIterator{
		iter: d.bitmap.Iterator(),
	}
}

func (d *roaringPostingsList) Clone() MutableList {
	d.RLock()
	// TODO: It's cheaper to Clone than to cache roaring bitmaps, see
	// `postings_list_bench_test.go`. Their internals don't allow for
	// pooling at the moment. We should address this when get a chance
	// (move to another implementation / address deficiencies).
	clone := d.bitmap.Clone()
	d.RUnlock()
	return &roaringPostingsList{
		bitmap: clone,
	}
}

func (d *roaringPostingsList) Equal(other List) bool {
	if d.Len() != other.Len() {
		return false
	}

	o, ok := other.(*roaringPostingsList)
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
}

type roaringIterator struct {
	iter    roaring.IntIterable
	current ID
	closed  bool
}

func (it *roaringIterator) Current() ID {
	return it.current
}

func (it *roaringIterator) Next() bool {
	if it.closed || !it.iter.HasNext() {
		return false
	}
	it.current = ID(it.iter.Next())
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
