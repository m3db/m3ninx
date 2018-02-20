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
	"regexp"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/postings"
)

const (
	regexpMatchFactor = 0.01
)

// simpleTermsDictionary uses two-level map to model a terms dictionary.
// i.e. fieldName -> fieldValue -> postingsList
type simpleTermsDictionary struct {
	opts   Options
	fields struct {
		sync.RWMutex
		nameToValuesMap map[string]*fieldValuesMap
		// TODO: as noted in https://github.com/m3db/m3ninx/issues/11, evalute impact of using
		// a custom hash map where we can avoid using string keys, both to save allocs and
		// help perf.
	}
}

func newSimpleTermsDictionary(opts Options) *simpleTermsDictionary {
	dict := &simpleTermsDictionary{
		opts: opts,
	}
	dict.fields.nameToValuesMap = make(map[string]*fieldValuesMap, opts.InitialCapacity())
	return dict
}

func (t *simpleTermsDictionary) Insert(field doc.Field, id postings.ID) error {
	name := string(field.Name)
	valsMap := t.getOrAddFieldName(name)
	val := string(field.Value)
	return valsMap.addID(id, val)
}

func (t *simpleTermsDictionary) MatchExact(field, value []byte) (postings.List, error) {
	t.fields.RLock()
	valsMap, ok := t.fields.nameToValuesMap[string(field)]
	t.fields.RUnlock()
	if !ok {
		// It is not an error to not have any matching values.
		return nil, nil
	}
	return valsMap.matchExact(value), nil
}

func (t *simpleTermsDictionary) MatchRegex(
	field, pattern []byte,
	re *regexp.Regexp,
) (postings.List, error) {
	t.fields.RLock()
	valsMap, ok := t.fields.nameToValuesMap[string(field)]
	t.fields.RUnlock()
	if !ok {
		// It is not an error to not have any matching values.
		return nil, nil
	}

	pls := valsMap.matchRegex(re)
	union := t.opts.PostingsListPool().Get()
	for _, pl := range pls {
		union.Union(pl)
	}
	return union, nil
}

func (t *simpleTermsDictionary) getOrAddFieldName(fieldName string) *fieldValuesMap {
	// Cheap read lock to see if it already exists.
	t.fields.RLock()
	fieldValues, ok := t.fields.nameToValuesMap[fieldName]
	t.fields.RUnlock()
	if ok {
		return fieldValues
	}

	// Acquire write lock and create.
	t.fields.Lock()
	fieldValues, ok = t.fields.nameToValuesMap[fieldName]

	// Check if it's been created since we last acquired the lock.
	if ok {
		t.fields.Unlock()
		return fieldValues
	}

	fieldValues = newFieldValuesMap(t.opts)
	t.fields.nameToValuesMap[fieldName] = fieldValues
	t.fields.Unlock()
	return fieldValues
}

type fieldValuesMap struct {
	sync.RWMutex

	opts Options

	// fieldValue -> postingsList
	values map[string]postings.MutableList
	// TODO: as noted in https://github.com/m3db/m3ninx/issues/11, evalute impact of using
	// a custom hash map where we can avoid using string keys, both to save allocs and
	// help perf.
}

func newFieldValuesMap(opts Options) *fieldValuesMap {
	return &fieldValuesMap{
		opts:   opts,
		values: make(map[string]postings.MutableList),
	}
}

func (f *fieldValuesMap) addID(id postings.ID, value string) error {
	// Try read lock to see if we already have a postings list for the given value.
	f.RLock()
	pl, ok := f.values[value]
	f.RUnlock()

	// We have a postings list, insert the ID and move on.
	if ok {
		return pl.Insert(id)
	}

	// A corresponding postings list doesn't exist, time to acquire write lock.
	f.Lock()
	pl, ok = f.values[value]

	// Check if the corresponding postings list has been created since we released lock.
	if ok {
		f.Unlock()
		return pl.Insert(id)
	}

	// Create a new posting list for the term, and insert into fieldValues.
	pl = f.opts.PostingsListPool().Get()
	f.values[value] = pl
	f.Unlock()
	return pl.Insert(id)
}

func (f *fieldValuesMap) matchExact(value []byte) postings.List {
	f.RLock()
	pl, ok := f.values[string(value)]
	f.RUnlock()
	if !ok {
		return nil
	}
	return pl
}

// TODO: consider returning an iterator here, this would require some kind of ordering semantics
// on the underlying map tho.
func (f *fieldValuesMap) matchRegex(re *regexp.Regexp) []postings.List {
	f.RLock()
	initLen := int(regexpMatchFactor * float64(len(f.values)))
	pls := make([]postings.List, 0, initLen)
	for val, pl := range f.values {
		// TODO: evaluate lock contention caused by holding on to the read lock while evaluating
		// this predicate.
		// TODO: evaluate if perform a prefix match would speed up the common case.
		if re.MatchString(val) {
			pls = append(pls, pl)
		}
	}
	f.RUnlock()

	return pls
}
