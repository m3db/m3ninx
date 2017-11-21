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

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/x/trigram"
)

var (
	errQueryMatchedEverything = errors.New("query matched all documents, provide more specific filters")
)

// consider the following example, given a field (name: fieldName, value: fieldValue), and DocID `i`
// (1) we compute all trigrams for the given value, in this case:
//	 	eld,iel,fie,lue,Val,dVa,alu,ldV
// (2) for each trigram from (1), we store a reference to the `i` in the posting list.

// trigramsTermsDict uses trigrams with two-level map to model a terms dictionary.
// i.e. fieldName > trigram(fieldValue) -> postingsList
type trigramTermsDictionary struct {
	opts Options

	// TODO(prateek): measure perf impact of using 4-gram instead of 3-gram
	backingDict *simpleTermsDictionary
}

func newTrigramTermsDictionary(opts Options) termsDictionary {
	td := newSimpleTermsDictionary(opts).(*simpleTermsDictionary)
	return &trigramTermsDictionary{
		backingDict: td,
	}
}

func (t *trigramTermsDictionary) Insert(field doc.Field, i segment.DocID) error {
	trigrams := computeTrigrams(field.Value)
	for _, tri := range trigrams {
		// TODO: change simpleTermsDict signature to not require a field for insertion
		if err := t.backingDict.Insert(doc.Field{
			Name:      field.Name,
			Value:     tri,
			ValueType: field.ValueType,
		}, i); err != nil {
			return err
		}
	}
	return nil
}

func (t *trigramTermsDictionary) Fetch(
	fieldName []byte,
	fieldValueFilter []byte,
	isRegex bool,
) (segment.PostingsList, error) {
	re, err := syntax.Parse(string(fieldValueFilter), syntax.Perl)
	if err != nil {
		return nil, err
	}
	q := trigram.RegexpQuery(re)
	// TODO(prateek): need to to do something special for `q.Op == trigram.QAll`, i.e. when generated regexp isn't restrictive enough
	// TODO(prateek): handle case where filter has been negated
	return t.postingQuery(fieldName, q, nil, false)
}

func (t *trigramTermsDictionary) postingQuery(
	fieldName []byte,
	q *trigram.Query,
	candidateIDs segment.PostingsList,
	created bool,
) (segment.PostingsList, error) {
	switch q.Op {
	case trigram.QNone:
		// do nothing
	case trigram.QAll:
		return nil, errQueryMatchedEverything

	case trigram.QAnd:
		for _, tri := range q.Trigram {
			ids, err := t.docIDsForTrigram(fieldName, tri)
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
			ids, err := t.postingQuery(fieldName, sub, candidateIDs, created)
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
			ids, err := t.docIDsForTrigram(fieldName, tri)
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
			ids, err := t.postingQuery(fieldName, sub, candidateIDs, created)
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

func (t *trigramTermsDictionary) docIDsForTrigram(
	fieldName []byte,
	tri string,
) (segment.ImmutablePostingsList, error) {
	return t.backingDict.Fetch(fieldName, []byte(tri), false)
}

// NB(prateek): benchmarked this implementation against a hand rolled one,
// this one was 2x faster and had 3x lesser code.
func computeTrigrams(value []byte) [][]byte {
	numTrigrams := len(value) - 2
	if numTrigrams <= 0 {
		return nil
	}

	trigrams := make([][]byte, 0, numTrigrams)
	for i := 2; i < len(value); i++ {
		trigrams = append(trigrams, value[i-2:i+1])
	}

	return trigrams
}
