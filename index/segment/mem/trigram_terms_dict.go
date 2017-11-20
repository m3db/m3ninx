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
	"regexp/syntax"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/x/trigram"
)

var (
	errQueryMatchedEverything = errors.New("query matched all documents, provide more specific filters")
)

// consider the following example, given a field (name: fieldName, value: fieldValue), and DocID `i`
// (1) we first compute the concatenated term: "fieldName=fieldValue",
// (2) then, we compute all trigrams for the given term, in this case:
//	 	eld,iel,fie,lue,dNa,Val,dVa,ame,ldN,Nam,e=f,me=,=fi,alu,ldV
// (3) for each trigram from (2), we store a reference to the `i` in the posting list.

// trigramsTermsDict uses trigrams to model a terms dictionary.
// map[trigram] => postingOffset
type trigramTermsDictionary struct {
	opts Options

	trigramsLock sync.RWMutex
	trigrams     map[string]postingsManagerOffset
	// TODO(prateek): measure impact of 2-level map instead of concatenating fieldname+fieldvalue.
	// TODO(prateek): measure perf impact of using 4-gram instead of 3-gram

	postingsManager postingsManager
}

func newTrigramTermsDictionary(opts Options, fn newPostingsManagerFn) termsDictionary {
	return &trigramTermsDictionary{
		trigrams:        make(map[string]postingsManagerOffset, opts.InitialCapacity()),
		postingsManager: fn(opts),
	}
}

func (t *trigramTermsDictionary) Insert(field doc.Field, i segment.DocID) error {
	trigrams := computeTrigrams(field)
	for _, tri := range trigrams {
		if err := t.insertTrigram(tri, i); err != nil {
			return err
		}
	}
	return nil
}

func (t *trigramTermsDictionary) insertTrigram(trigram string, i segment.DocID) error {
	// see if we already have an offset for this trigram
	t.trigramsLock.RLock()
	off, ok := t.trigrams[trigram]
	t.trigramsLock.RUnlock()

	// found it, so update posting list and terminate
	if ok {
		return t.postingsManager.Update(off, i)
	}

	// need to create offset for this trigram
	t.trigramsLock.Lock()

	// see if it's been inserted since we released lock
	off, ok = t.trigrams[trigram]
	if ok {
		t.trigramsLock.Unlock()
		return t.postingsManager.Update(off, i)
	}

	off = t.postingsManager.Insert(i)
	t.trigrams[trigram] = off
	t.trigramsLock.Unlock()
	return nil
}

func (t *trigramTermsDictionary) Fetch(
	fieldName []byte,
	fieldValueFilter []byte,
	isRegex bool,
) (segment.PostingsList, error) {
	// TODO(prateek): measure cost of compiling regular expression multiple times.
	// potentially could avoid doing that if we had a two-level map for the term dict.
	filter := string(fieldName) + "=" + string(fieldValueFilter)
	re, err := syntax.Parse(filter, syntax.Perl)
	if err != nil {
		return nil, err
	}
	q := trigram.RegexpQuery(re)
	// TODO(prateek): handle case where filter has been negated
	return t.postingQuery(q, nil, false)
}

func (t *trigramTermsDictionary) postingQuery(
	q *trigram.Query,
	candidateIDs segment.PostingsList,
	created bool,
) (segment.PostingsList, error) {
	switch q.Op {
	case trigram.QNone:
		// do nothing
	case trigram.QAll:
		// panic("should never happen")
		return nil, errQueryMatchedEverything

	case trigram.QAnd:
		for _, tri := range q.Trigram {
			ids, err := t.docIDsForTrigram(tri)
			if err != nil {
				return nil, err
			}
			if ids == nil {
				return t.opts.PostingsListPool().Get(), nil
			}
			if !created {
				candidateIDs = ids.Clone()
				created = true
			} else {
				candidateIDs.Intersect(ids)
			}
			if candidateIDs.IsEmpty() {
				return candidateIDs, nil
			}
		}
		for _, sub := range q.Sub {
			ids, err := t.postingQuery(sub, candidateIDs, created)
			if err != nil {
				return nil, err
			}
			if ids == nil {
				return t.opts.PostingsListPool().Get(), nil
			}
			if !created {
				candidateIDs = ids
				created = true
			} else {
				candidateIDs.Intersect(ids)
			}
			if candidateIDs.IsEmpty() {
				return candidateIDs, nil
			}
		}
	case trigram.QOr:
		for _, tri := range q.Trigram {
			ids, err := t.docIDsForTrigram(tri)
			if err != nil {
				return nil, err
			}
			if ids == nil {
				continue
			}
			if !created {
				candidateIDs = ids.Clone()
				created = true
			} else {
				candidateIDs.Union(ids)
			}
		}
		for _, sub := range q.Sub {
			ids, err := t.postingQuery(sub, candidateIDs, created)
			if err != nil {
				return nil, err
			}
			if ids == nil {
				return t.opts.PostingsListPool().Get(), nil
			}
			if !created {
				candidateIDs = ids
				created = true
			} else {
				candidateIDs.Union(ids)
			}
		}
	}
	return candidateIDs, nil
}

func (t *trigramTermsDictionary) docIDsForTrigram(tri string) (segment.ImmutablePostingsList, error) {
	t.trigramsLock.RLock()
	off, ok := t.trigrams[tri]
	t.trigramsLock.RUnlock()
	if !ok {
		return nil, nil
	}
	return t.postingsManager.Fetch(off)
}

// NB(prateek): benchmarked this implementation against a hand rolled one,
// this one was 2x faster and had 3x lesser code.
func computeTrigrams(f doc.Field) []string {
	concat := string(f.Name) + "=" + string(f.Value)
	numTrigrams := len(concat) - 2
	trigrams := make([]string, 0, numTrigrams)
	for i := 2; i < len(concat); i++ {
		trigrams = append(trigrams, concat[i-2:i+1])
	}
	return trigrams
}
