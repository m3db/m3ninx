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
	"github.com/m3db/m3ninx/index/segment"
)

// simpleTermsDictionary uses two-level map to model a terms dictionary.
// i.e. fieldName -> fieldValue -> postingsManagerOffset
type simpleTermsDictionary struct {
	fieldNamesLock sync.RWMutex
	fieldNames     map[string]*fieldValuesMap

	opts            Options
	postingsManager postingsManager
}

// newPostingsManagerFn returns a new PostingsList.
type newPostingsManagerFn func(Options) postingsManager

func newSimpleTermsDictionary(opts Options, fn newPostingsManagerFn) termsDictionary {
	return &simpleTermsDictionary{
		fieldNames:      make(map[string]*fieldValuesMap, opts.InitialCapacity()),
		opts:            opts,
		postingsManager: fn(opts),
	}
}

func (t *simpleTermsDictionary) Insert(field doc.Field, i segment.DocID) error {
	fieldName := string(field.Name)
	fieldValues := t.getOrAddFieldName(fieldName)
	fieldValue := string(field.Value)
	return fieldValues.addDocIDForValue(fieldValue, i, t.postingsManager)
}

func (t *simpleTermsDictionary) Fetch(
	fieldName []byte,
	fieldValueFilter []byte,
	isRegexp bool,
) (segment.PostingsList, error) {
	// check if we know about the field name
	t.fieldNamesLock.RLock()
	fieldValues, ok := t.fieldNames[string(fieldName)]
	t.fieldNamesLock.RUnlock()
	if !ok {
		// not an error to not have any matching values
		return nil, nil
	}

	// get postingsManagerOffset(s) for the given value.
	offsets, err := fieldValues.fetchOffsets(fieldValueFilter, isRegexp)
	if err != nil {
		return nil, err
	}

	// union all the docsIDSets
	set := segment.NewPostingsList()
	for _, pid := range offsets {
		// check if we have docIDs set for the given posting list.
		ids, err := t.postingsManager.Fetch(pid)
		if err != nil {
			return nil, err
		}
		if ids == nil {
			continue
		}
		set.Union(ids)
	}

	return set, nil
}

func (t *simpleTermsDictionary) getOrAddFieldName(fieldName string) *fieldValuesMap {
	// cheap read lock to see if it already exists
	t.fieldNamesLock.RLock()
	fieldValues, ok := t.fieldNames[fieldName]
	t.fieldNamesLock.RUnlock()
	if ok {
		return fieldValues
	}

	// acquire write lock and create
	t.fieldNamesLock.Lock()
	fieldValues, ok = t.fieldNames[fieldName]

	// check if it's been created since we last acquired the lock
	if ok {
		t.fieldNamesLock.Unlock()
		return fieldValues
	}

	fieldValues = newFieldValuesMap()
	t.fieldNames[fieldName] = fieldValues
	t.fieldNamesLock.Unlock()
	return fieldValues
}

type fieldValuesMap struct {
	sync.RWMutex
	// fieldValue -> postingsManagerOffset
	values map[string]postingsManagerOffset
}

func newFieldValuesMap() *fieldValuesMap {
	return &fieldValuesMap{
		values: make(map[string]postingsManagerOffset),
	}
}

func (f *fieldValuesMap) addDocIDForValue(value string, i segment.DocID, pl postingsManager) error {
	// try read lock to see if we already have a postingsManagerOffset for the given value.
	f.RLock()
	pid, ok := f.values[value]
	f.RUnlock()

	// we have a postingsManagerOffset, mark the docID and move on.
	if ok {
		return pl.Update(pid, i)
	}

	// postingsManagerOffset doesn't exist, time to acquire write lock
	f.Lock()
	pid, ok = f.values[value]

	// check if it's been created since we released lock
	if ok {
		f.Unlock()
		return pl.Update(pid, i)
	}

	// create new posting id for the term, and insert into fieldValues
	offset := pl.Insert(i)
	f.values[value] = offset
	f.Unlock()
	return nil
}

func (f *fieldValuesMap) fetchOffsets(valueFilter []byte, regexp bool) ([]postingsManagerOffset, error) {
	// special case when we're looking for an exact match
	if !regexp {
		return f.fetchExact(valueFilter), nil
	}

	// otherwise, we have to iterate over all known values
	pred, err := newRegexPredicate(valueFilter)
	if err != nil {
		return nil, err
	}
	var offsets []postingsManagerOffset
	f.RLock()
	for value, offset := range f.values {
		if pred(value) {
			offsets = append(offsets, offset)
		}
	}
	f.RUnlock()

	return offsets, nil
}

func (f *fieldValuesMap) fetchExact(valueFilter []byte) []postingsManagerOffset {
	f.RLock()
	pid, ok := f.values[string(valueFilter)]
	f.RUnlock()
	if !ok {
		return nil
	}
	return []postingsManagerOffset{pid}
}

func newRegexPredicate(valueFilter []byte) (valuePredicate, error) {
	filter := string(valueFilter)
	re, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	return re.MatchString, nil
}

type valuePredicate func(v string) bool

func mustNotBeOutOfRange(err error) {
	if err == ErrOutOfRange {
		panic(err)
	}
}
