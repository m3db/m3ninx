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
	"math"
	"sync"

	"github.com/m3db/m3ninx/index/segment"

	"github.com/uber-go/atomic"
)

var (
	// ErrOutOfRange is returned when the provided offset does not exist.
	ErrOutOfRange = errors.New("provided offset does not exist")
)

type postingsMgr struct {
	sync.RWMutex
	postings []segment.PostingsList

	opts    Options
	pool    segment.PostingsListPool
	counter *atomic.Int32
}

func newPostingsManager(opts Options) postingsManager {
	return &postingsMgr{
		postings: make([]segment.PostingsList, opts.InitialCapacity()),
		opts:     opts,
		pool:     opts.PostingsListPool(),
		counter:  atomic.NewInt32(-1),
	}
}

func (p *postingsMgr) Insert(i segment.DocID) postingsManagerOffset {
	t := postingsManagerOffset(p.counter.Inc())
	p.ensureSufficientCapacity(t)

	p.Lock()
	plist := p.postings[t]
	if plist == nil {
		plist = p.pool.Get()
		p.postings[t] = plist
	}
	plist.Insert(i)
	p.Unlock()
	return t
}

func (p *postingsMgr) ensureSufficientCapacity(t postingsManagerOffset) {
	p.RLock()
	size := len(p.postings)
	p.RUnlock()

	// early terminate if we have sufficient capacity
	if int(t) < size {
		return
	}

	// expand capacity
	p.Lock()

	// early terminate if we've expanded capacity since releasing lock last
	if size = len(p.postings); int(t) < size {
		p.Unlock()
		return
	}

	// amortized O(1) expansion
	factor := math.Ceil(float64(t) / float64(len(p.postings)))
	if factor < 2 {
		factor = 2
	}
	newPostings := make([]segment.PostingsList, 1+int(factor)*len(p.postings))
	copy(newPostings, p.postings)
	p.postings = newPostings

	p.Unlock()
}

func (p *postingsMgr) Update(t postingsManagerOffset, i segment.DocID) error {
	p.RLock()
	// range check
	if int(t) >= len(p.postings) {
		return ErrOutOfRange
	}

	ids := p.postings[t]
	if ids == nil {
		return ErrOutOfRange
	}
	p.RUnlock()

	ids.Insert(i)
	return nil
}

func (p *postingsMgr) Fetch(t postingsManagerOffset) (segment.ImmutablePostingsList, error) {
	p.RLock()
	if int(t) >= len(p.postings) || p.postings[t] == nil {
		p.RUnlock()
		return nil, ErrOutOfRange
	}

	ids := p.postings[t]
	p.RUnlock()
	return ids, nil
}
