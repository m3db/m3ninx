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
	"bytes"
	"fmt"
	"regexp"
	"regexp/syntax"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"

	"github.com/m3db/codesearch/index"
)

var (
	// sentinelTrigram is stored for every field so that we can perform match all lookups.
	sentinelTrigram = ""
)

// The trigram terms dictionary works by breaking down the value of a field into its constiuent
// trigrams and storing each trigram in a simple dictionary. For example, given a field
// (name: "foo", value: "fizzbuzz") and DocID `i` we
//   (1) Compute all trigrams for the given value. In this case:
//         fiz, izz, zzb, zbu, buz, uzz
//   (2) For each trigram `t` created in step 1, store the entry (value: `t`, docID: `i`) in the
//       postings list for the field name "foo".
//

// trigramsTermsDict uses trigrams in a two-level map to model a terms dictionary.
// i.e. fieldName > trigram(fieldValue) -> postingsList
type trigramTermsDictionary struct {
	opts Options

	fields struct {
		sync.RWMutex
		idsMap map[segment.DocID][]doc.Field
	}
	backingDict *simpleTermsDictionary
}

func newTrigramTermsDictionary(opts Options) termsDictionary {
	std := newSimpleTermsDictionary(opts).(*simpleTermsDictionary)
	ttd := &trigramTermsDictionary{
		opts:        opts,
		backingDict: std,
	}
	ttd.fields.idsMap = make(map[segment.DocID][]doc.Field)
	return ttd
}

func (t *trigramTermsDictionary) Insert(field doc.Field, i segment.DocID) error {
	// TODO: Benchmark performance difference between first constructing a set of unique
	// trigrams versus inserting all trigrams and relying on the backing dictionary to
	// deduplicate them.
	trigrams := computeTrigrams(field.Value)
	for _, tri := range trigrams {
		if err := t.backingDict.Insert(
			doc.Field{
				Name:      field.Name,
				Value:     tri,
				ValueType: field.ValueType,
			},
			i,
		); err != nil {
			return err
		}
	}
	t.fields.Lock()
	defer t.fields.Unlock()
	t.fields.idsMap[i] = append(t.fields.idsMap[i], field)
	return nil
}

func (t *trigramTermsDictionary) Fetch(
	fieldName []byte,
	fieldValueFilter []byte,
	opts termFetchOptions,
) (segment.PostingsList, error) {
	re, err := syntax.Parse(string(fieldValueFilter), syntax.Perl)
	if err != nil {
		return nil, err
	}
	q := index.RegexpQuery(re)
	canidates, err := t.postingQuery(fieldName, q, nil, false)
	if err != nil {
		return nil, err
	}
	defer t.opts.PostingsListPool().Put(canidates)

	// NB: The trigram index can return false postives so we verify that the returned
	// documents do in fact match the given filter below.
	var (
		regex *regexp.Regexp
		it    = canidates.Iter()
		ids   = t.opts.PostingsListPool().Get()
	)
	if opts.isRegexp {
		regex, err = regexp.Compile(string(fieldValueFilter))
		if err != nil {
			return nil, err
		}
	}

	// TODO: Investigate releasing the lock every N iterations so we don't block
	// inserts for the entire time we hold the lock.
	t.fields.RLock()
	defer t.fields.RUnlock()

	for it.Next() {
		id := it.Current()
		fields, ok := t.fields.idsMap[id]
		if !ok {
			return nil, fmt.Errorf("ID '%v' found in postings list but not ID map", id)
		}

		var matched bool
		for _, field := range fields {
			if !bytes.Equal(fieldName, field.Name) {
				continue
			}

			if opts.isRegexp {
				if regex.Match(field.Value) {
					matched = true
					break
				}
			} else {
				if bytes.Equal(fieldValueFilter, field.Value) {
					matched = true
					break
				}
			}
		}

		if matched {
			ids.Insert(id)
		}
	}

	return ids, nil
}

func (t *trigramTermsDictionary) postingQuery(
	fieldName []byte,
	q *index.Query,
	candidateIDs segment.PostingsList,
	created bool,
) (segment.PostingsList, error) {
	switch q.Op {
	case index.QNone:
		// Do nothing.

	case index.QAll:
		if candidateIDs != nil {
			return candidateIDs, nil
		}
		ids, err := t.docIDsForTrigram(fieldName, sentinelTrigram)
		if err != nil {
			return nil, err
		}
		candidateIDs = ids.Clone()

	case index.QAnd:
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

	case index.QOr:
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
	return t.backingDict.Fetch(fieldName, []byte(tri), termFetchOptions{isRegexp: false})
}

// computeTrigrams returns the trigrams composing a byte slice, including the sentinel
// trigram. The slice of trigrams returned is not guaranteed to be unique.
func computeTrigrams(value []byte) [][]byte {
	numTrigrams := len(value) - 2
	trigrams := make([][]byte, 0, numTrigrams+1)
	for i := 2; i < len(value); i++ {
		trigrams = append(trigrams, value[i-2:i+1])
	}
	// NB: Taking a byte slice of an empty string won't allocate, see
	// BenchmarkEmptyStringToByteSlice.
	trigrams = append(trigrams, []byte(sentinelTrigram))

	return trigrams
}
