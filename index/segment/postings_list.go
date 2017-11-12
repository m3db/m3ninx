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

package segment

import (
	"sync"

	"github.com/RoaringBitmap/roaring"
)

// roaringPostingsList wraps a roaring.Bitmap w/ a mutex for thread safety.
type roaringPostingsList struct {
	sync.RWMutex
	bitmap *roaring.Bitmap
}

// NewPostingsList returns a new PostingsList.
func NewPostingsList() PostingsList {
	return &roaringPostingsList{
		bitmap: roaring.NewBitmap(),
	}
}

func (d *roaringPostingsList) Insert(i DocID) {
	d.Lock()
	d.bitmap.Add(uint32(i))
	d.Unlock()
}

func (d *roaringPostingsList) Intersect(other ImmutablePostingsList) {
	o, ok := other.(*roaringPostingsList)
	if !ok {
		panic("Intersect only supported between roaringDocId sets")
	}

	d.Lock()
	o.RLock()
	d.bitmap.And(o.bitmap)
	o.RUnlock()
	d.Unlock()
}

func (d *roaringPostingsList) Union(other ImmutablePostingsList) {
	o, ok := other.(*roaringPostingsList)
	if !ok {
		panic("Union only supported between roaringDocId sets")
	}

	d.Lock()
	o.RLock()
	d.bitmap.Or(o.bitmap)
	o.RUnlock()
	d.Unlock()
}

func (d *roaringPostingsList) Reset() {
	d.Lock()
	d.bitmap.Clear()
	d.Unlock()
}

func (d *roaringPostingsList) Contains(i DocID) bool {
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

func (d *roaringPostingsList) Size() uint64 {
	d.RLock()
	size := d.bitmap.GetCardinality()
	d.RUnlock()
	return size
}

func (d *roaringPostingsList) Iter() PostingsIter {
	// make a copy to ensure iteration doesn't conflict with updates
	other := d.Clone().(*roaringPostingsList)
	return &roaringIter{
		i: other.bitmap.Iterator(),
	}
}

func (d *roaringPostingsList) Clone() PostingsList {
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

type roaringIter struct {
	i roaring.IntIterable
}

func (r *roaringIter) Current() DocID {
	return DocID(r.i.Next())
}

func (r *roaringIter) Next() bool {
	return r.i.HasNext()
}