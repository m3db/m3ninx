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
	re "regexp"
	"sync"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment/mem/fieldsgen"
	"github.com/m3db/m3ninx/index/segment/mem/postingsgen"
	"github.com/m3db/m3ninx/postings"
)

// termsDict is an in-memory terms dictionary. It maps fields to postings lists.
type termsDict struct {
	opts Options

	fields struct {
		sync.RWMutex
		internalMap *fieldsgen.Map
	}
}

func newTermsDict(opts Options) termsDictionary {
	dict := &termsDict{
		opts: opts,
	}
	dict.fields.internalMap = fieldsgen.New(opts.InitialCapacity())
	return dict
}

func (d *termsDict) Insert(field doc.Field, id postings.ID) error {
	postingsMap := d.getOrAddName(field.Name)
	return postingsMap.Add(field.Value, id)
}

func (d *termsDict) ContainsTerm(field, term []byte) (bool, error) {
	_, found, err := d.matchTerm(field, term)
	if err != nil {
		return false, err
	}
	return found, nil
}

func (d *termsDict) MatchTerm(field, term []byte) (postings.List, error) {
	pl, found, err := d.matchTerm(field, term)
	if err != nil {
		return nil, err
	}
	if !found {
		return d.opts.PostingsListPool().Get(), nil
	}
	return pl, nil
}

func (d *termsDict) matchTerm(field, term []byte) (postings.List, bool, error) {
	d.fields.RLock()
	postingsMap, ok := d.fields.internalMap.Get(field)
	d.fields.RUnlock()
	if !ok {
		return nil, false, nil
	}
	pl, ok := postingsMap.Get(term)
	if !ok {
		return nil, false, nil
	}
	return pl, true, nil
}

func (d *termsDict) MatchRegexp(
	field, regexp []byte,
	compiled *re.Regexp,
) (postings.List, error) {
	d.fields.RLock()
	postingsMap, ok := d.fields.internalMap.Get(field)
	d.fields.RUnlock()
	if !ok {
		return d.opts.PostingsListPool().Get(), nil
	}
	pl, ok := postingsMap.GetRegex(compiled)
	if !ok {
		return d.opts.PostingsListPool().Get(), nil
	}
	return pl, nil
}

func (d *termsDict) getOrAddName(name []byte) *postingsgen.ConcurrentMap {
	// Cheap read lock to see if it already exists.
	d.fields.RLock()
	postingsMap, ok := d.fields.internalMap.Get(name)
	d.fields.RUnlock()
	if ok {
		return postingsMap
	}

	// Acquire write lock and create.
	d.fields.Lock()
	postingsMap, ok = d.fields.internalMap.Get(name)

	// Check if it's been created since we last acquired the lock.
	if ok {
		d.fields.Unlock()
		return postingsMap
	}

	postingsMap = postingsgen.NewConcurrentMap(postingsgen.ConcurrentMapOpts{
		InitialSize:      d.opts.InitialCapacity(),
		PostingsListPool: d.opts.PostingsListPool(),
	})
	d.fields.internalMap.SetUnsafe(name, postingsMap, fieldsgen.SetUnsafeOptions{
		NoCopyKey:     true,
		NoFinalizeKey: true,
	})
	d.fields.Unlock()
	return postingsMap
}
