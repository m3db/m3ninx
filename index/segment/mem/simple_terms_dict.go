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

func newSimpleTermsDictionary(opts Options) termsDictionary {
	dict := &simpleTermsDictionary{
		opts: opts,
	}
	dict.fields.nameToValuesMap = make(map[string]*fieldValuesMap, opts.InitialCapacity())
	return dict
}

func (t *simpleTermsDictionary) Insert(field doc.Field, i segment.DocID) error {
	fieldName := string(field.Name)
	fieldValues := t.getOrAddFieldName(fieldName)
	fieldValue := string(field.Value)
	return fieldValues.addDocIDForValue(fieldValue, i)
}

func (t *simpleTermsDictionary) Fetch(
	fieldName []byte,
	fieldValueFilter []byte,
	isRegexp bool,
) (segment.PostingsList, error) {
	// check if we know about the field name
	t.fields.RLock()
	fieldValues, ok := t.fields.nameToValuesMap[string(fieldName)]
	t.fields.RUnlock()
	if !ok {
		// not an error to not have any matching values
		return nil, nil
	}

	// get postingList(s) for the given value.
	lists, err := fieldValues.fetchLists(fieldValueFilter, isRegexp)
	if err != nil {
		return nil, err
	}

	// union all the postingsList(s)
	unionedList := t.opts.PostingsListPool().Get()
	for _, ids := range lists {
		unionedList.Union(ids)
	}

	return unionedList, nil
}

func (t *simpleTermsDictionary) getOrAddFieldName(fieldName string) *fieldValuesMap {
	// cheap read lock to see if it already exists
	t.fields.RLock()
	fieldValues, ok := t.fields.nameToValuesMap[fieldName]
	t.fields.RUnlock()
	if ok {
		return fieldValues
	}

	// acquire write lock and create
	t.fields.Lock()
	fieldValues, ok = t.fields.nameToValuesMap[fieldName]

	// check if it's been created since we last acquired the lock
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
	values map[string]segment.PostingsList
	// TODO: as noted in https://github.com/m3db/m3ninx/issues/11, evalute impact of using
	// a custom hash map where we can avoid using string keys, both to save allocs and
	// help perf.
}

func newFieldValuesMap(opts Options) *fieldValuesMap {
	return &fieldValuesMap{
		opts:   opts,
		values: make(map[string]segment.PostingsList),
	}
}

func (f *fieldValuesMap) addDocIDForValue(value string, i segment.DocID) error {
	// try read lock to see if we already have a postingsList for the given value.
	f.RLock()
	pid, ok := f.values[value]
	f.RUnlock()

	// we have a postingsList, mark the docID and move on.
	if ok {
		pid.Insert(i)
		return nil
	}

	// postingsList doesn't exist, time to acquire write lock
	f.Lock()
	pid, ok = f.values[value]

	// check if it's been created since we released lock
	if ok {
		f.Unlock()
		pid.Insert(i)
		return nil
	}

	// create new posting id for the term, and insert into fieldValues
	pid = f.opts.PostingsListPool().Get()
	f.values[value] = pid
	f.Unlock()
	pid.Insert(i)
	return nil
}

func (f *fieldValuesMap) fetchLists(valueFilter []byte, regexp bool) ([]segment.PostingsList, error) {
	// special case when we're looking for an exact match
	if !regexp {
		return f.fetchExact(valueFilter), nil
	}

	// otherwise, we have to iterate over all known values
	pred, err := newRegexPredicate(valueFilter)
	if err != nil {
		return nil, err
	}
	var postingsLists []segment.PostingsList
	f.RLock()
	for value, list := range f.values {
		if pred(value) {
			postingsLists = append(postingsLists, list)
		}
	}
	f.RUnlock()

	return postingsLists, nil
}

func (f *fieldValuesMap) fetchExact(valueFilter []byte) []segment.PostingsList {
	f.RLock()
	pid, ok := f.values[string(valueFilter)]
	f.RUnlock()
	if !ok {
		return nil
	}
	return []segment.PostingsList{pid}
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
