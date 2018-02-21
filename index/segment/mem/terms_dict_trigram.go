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
	"fmt"
	"regexp"
	"regexp/syntax"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/postings"

	cindex "github.com/m3db/codesearch/index"
)

var (
	errUnsupportedQuery = errors.New("query is not supported by the trigram terms dictionary")
)

// trigramsTermsDict stores trigrams of terms to support regular expression queries
// more efficiently. It does not support all queries and returns an errUnsupportedQuery
// for queries which it does not support. It can also return false positives and it is the
// caller's responsibility to check the documents to remove false positives.
//
// The trigram terms dictionary works by breaking down the value of a field into
// its constiuent trigrams and storing each trigram in a simple dictionary. For
// example, given a field (name: "foo", value: "fizzbuzz") and DocID `i` we
//   (1) Compute all trigrams for the given value. In this case:
//         fiz, izz, zzb, zbu, buz, uzz
//   (2) For each trigram `t` created in step 1, store the entry
//       (value: `t`, docID: `i`) in the postings list for the field name "foo".
//
type trigramTermsDict struct {
	opts Options

	backingDict termsDict
}

func newTrigramTermsDict(opts Options) termsDict {
	return &trigramTermsDict{
		opts:        opts,
		backingDict: newSimpleTermsDict(opts),
	}
}

func (t *trigramTermsDict) Insert(field doc.Field, id postings.ID) error {
	// TODO: Benchmark performance difference between first constructing a set of unique
	// trigrams versus inserting all trigrams and relying on the backing dictionary to
	// deduplicate them.
	trigrams := computeTrigrams(field.Value)
	for _, tri := range trigrams {
		if err := t.backingDict.Insert(
			doc.Field{
				Name:  field.Name,
				Value: tri,
			},
			id,
		); err != nil {
			return fmt.Errorf("unable to insert trigram %s into backing dictionary: %v", tri, err)
		}
	}
	return nil
}

func (t *trigramTermsDict) MatchExact(name, value []byte) (postings.List, error) {
	return t.matchRegex(name, value)
}

func (t *trigramTermsDict) MatchRegex(
	name, pattern []byte,
	re *regexp.Regexp,
) (postings.List, error) {
	return t.matchRegex(name, pattern)
}

func (t *trigramTermsDict) matchRegex(name, pattern []byte) (postings.List, error) {
	// TODO: Consider updating syntax.Parse to accepts a byte string so we can avoid the
	// conversion to a string here.
	patternStr := string(pattern)
	re, err := syntax.Parse(patternStr, syntax.Perl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse regular expression %s: %v", patternStr, err)
	}
	q := cindex.RegexpQuery(re)
	pl, err := t.matchQuery(name, q, nil, false)
	if err != nil {
		return nil, fmt.Errorf("unable to get postings list matching query: %v", err)
	}
	return pl, nil
}

func (t *trigramTermsDict) matchQuery(
	name []byte,
	q *cindex.Query,
	candidatePL postings.MutableList,
	created bool,
) (postings.MutableList, error) {
	switch q.Op {
	case cindex.QNone:
		// Do nothing.

	case cindex.QAll:
		if candidatePL != nil {
			return candidatePL, nil
		}
		// Match all queries are not supported by the trigram terms dictionary.
		return nil, errUnsupportedQuery

	case cindex.QAnd:
		for _, tri := range q.Trigram {
			pl, err := t.matchTrigram(name, tri)
			if err != nil {
				return nil, err
			}
			if pl == nil {
				return t.opts.PostingsListPool().Get(), nil
			}
			if !created {
				candidatePL = pl.Clone()
				created = true
			} else {
				candidatePL.Intersect(pl)
			}
			if candidatePL.IsEmpty() {
				return candidatePL, nil
			}
		}

		for _, sub := range q.Sub {
			pl, err := t.matchQuery(name, sub, candidatePL, created)
			if err != nil {
				return nil, err
			}
			if pl == nil {
				return t.opts.PostingsListPool().Get(), nil
			}
			if !created {
				candidatePL = pl
				created = true
			} else {
				candidatePL.Intersect(pl)
			}
			if candidatePL.IsEmpty() {
				return candidatePL, nil
			}
		}

	case cindex.QOr:
		for _, tri := range q.Trigram {
			pl, err := t.matchTrigram(name, tri)
			if err != nil {
				return nil, err
			}
			if pl == nil {
				continue
			}
			if !created {
				candidatePL = pl.Clone()
				created = true
			} else {
				candidatePL.Union(pl)
			}
		}

		for _, sub := range q.Sub {
			pl, err := t.matchQuery(name, sub, candidatePL, created)
			if err != nil {
				return nil, err
			}
			if pl == nil {
				return t.opts.PostingsListPool().Get(), nil
			}
			if !created {
				candidatePL = pl
				created = true
			} else {
				candidatePL.Union(pl)
			}
		}
	}

	return candidatePL, nil
}

func (t *trigramTermsDict) matchTrigram(name []byte, tri string) (postings.List, error) {
	// TODO(jeromefroe): Consider adding a FetchString method to the simpleDictionary
	// to avoid the string conversion here.
	return t.backingDict.MatchExact(name, []byte(tri))
}

// computeTrigrams returns the trigrams composing a byte slice. The slice of trigrams
// returned is not guaranteed to be unique.
func computeTrigrams(value []byte) [][]byte {
	numTrigrams := len(value) - 2
	if numTrigrams < 1 {
		return nil
	}

	trigrams := make([][]byte, 0, numTrigrams)
	for i := 2; i < len(value); i++ {
		trigrams = append(trigrams, value[i-2:i+1])
	}
	return trigrams
}
