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

package index

import (
	"crypto/md5"
	"sync"
	"time"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/index/segment/mem"
	"github.com/uber-go/tally"

	"go.uber.org/atomic"
)

// TODO(prateek): move to options
const initialSize = 1024 * 1024

type compositeIndex struct {
	opts   Options
	hashFn doc.HashFn

	// NB(prateek): cheap "don't write" cache internal to the process
	// hash(doc.ID)  -> internal docID
	// TODO(prateek): don't write cache needs the following fixes:
	//   - this needs to be made into an actual 2 level hash-map. We
	//     currently treat the hash of the document as a unique
	//     identifier. It's not.
	//   - This can't be allowed to grow un-bounded. Should change it
	//     to a LRU/Arc instead.
	seenIdsLock sync.RWMutex
	seenIdsMap  map[doc.Hash]struct{}

	active  *atomic.Bool
	metrics compositeIndexMetrics

	memSegment mem.Segment
}

// New returns a new Index.
func New(opts Options) (Index, error) {
	memSeg, err := mem.New(opts.MemSegmentOptions())
	if err != nil {
		return nil, err
	}

	hashFn := func(i doc.ID) doc.Hash {
		// TODO: evaluate impact of other hash functions on perf/correctness
		return md5.Sum([]byte(i))
	}
	idx := &compositeIndex{
		opts:   opts,
		hashFn: hashFn,

		seenIdsMap: make(map[doc.Hash]struct{}, initialSize),
		active:     atomic.NewBool(true),
		metrics:    newMetrics(opts.InstrumentOptions().MetricsScope()),
		memSegment: memSeg,
	}

	go idx.reportMetrics()
	return idx, nil
}

func (i *compositeIndex) reportMetrics() {
	t := time.NewTicker(i.opts.InstrumentOptions().ReportInterval())
	defer t.Stop()
	for i.active.Load() {
		<-t.C
		i.metrics.cardinality.Update(float64((i.Size())))
	}
}

func (i *compositeIndex) Insert(d doc.Document) error {
	// early terminate if id is already present
	docHash := i.hashFn(d.ID)
	i.seenIdsLock.RLock()
	_, ok := i.seenIdsMap[docHash]
	i.seenIdsLock.RUnlock()
	if ok {
		i.metrics.numAlreadySeen.Inc(1)
		return ErrDocAlreadyInserted
	}

	// insert into don't write cache
	i.seenIdsLock.Lock()
	// check if it's been inserted since we released lock
	_, ok = i.seenIdsMap[docHash]
	if ok {
		i.seenIdsLock.Unlock()
		i.metrics.numAlreadySeen.Inc(1)
		return ErrDocAlreadyInserted
	}
	i.seenIdsMap[docHash] = struct{}{}
	i.seenIdsLock.Unlock()

	return i.memSegment.Insert(d)
}

func (i *compositeIndex) Query(q segment.Query) ([]doc.Document, error) {
	return i.memSegment.Query(q)
}

func (i *compositeIndex) Size() uint32 {
	return i.memSegment.Size()
}

func (i *compositeIndex) Open() error {
	panic("not implemented")
}

func (i *compositeIndex) Close() error {
	panic("not implemented")
}

func (i *compositeIndex) Update(d doc.Document) error {
	panic("not implemented")
}

func (i *compositeIndex) Delete(d doc.Document) error {
	panic("not implemented")
}

func (i *compositeIndex) Freeze() error {
	panic("not implemented")
}

func (i *compositeIndex) Iter() segment.Iter {
	panic("not implemented")
}

func (i *compositeIndex) ID() segment.ID {
	panic("not implemented")
}

func (i *compositeIndex) Optimize() error {
	panic("not implemented")
}

type compositeIndexMetrics struct {
	cardinality    tally.Gauge
	numAlreadySeen tally.Counter
}

func newMetrics(s tally.Scope) compositeIndexMetrics {
	scope := s.SubScope("composite")
	return compositeIndexMetrics{
		cardinality:    scope.Gauge("cardinality"),
		numAlreadySeen: scope.Counter("num-seen"),
	}
}
