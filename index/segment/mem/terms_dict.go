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
	"github.com/m3db/m3ninx/postings"
)

// termsDict is an in-memory terms dictionary. It maps fields to postings lists.
type termsDict struct {
	opts Options

	fields struct {
		sync.RWMutex

		mapping map[string]*postingsMap
	}
}

func newTermsDict(opts Options) termsDictionary {
	dict := &termsDict{
		opts: opts,
	}
	dict.fields.mapping = make(map[string]*postingsMap, opts.InitialCapacity())
	return dict
}

func (d *termsDict) Insert(field doc.Field, id postings.ID) error {
	name := string(field.Name)
	postingsMap := d.getOrAddName(name)
	return postingsMap.addID(field.Value, id)
}

func (d *termsDict) MatchTerm(field, term []byte) (postings.List, error) {
	d.fields.RLock()
	postingsMap, ok := d.fields.mapping[string(field)]
	d.fields.RUnlock()
	if !ok {
		// It is not an error to not have any matching values.
		return d.opts.PostingsListPool().Get(), nil
	}
	return postingsMap.get(term), nil
}

func (d *termsDict) MatchRegexp(
	field, regexp []byte,
	compiled *re.Regexp,
) (postings.List, error) {
	d.fields.RLock()
	postingsMap, ok := d.fields.mapping[string(field)]
	d.fields.RUnlock()
	if !ok {
		// It is not an error to not have any matching values.
		return d.opts.PostingsListPool().Get(), nil
	}

	pls := postingsMap.getRegex(compiled)
	union := d.opts.PostingsListPool().Get()
	for _, pl := range pls {
		union.Union(pl)
	}
	return union, nil
}

func (d *termsDict) getOrAddName(name string) *postingsMap {
	// Cheap read lock to see if it already exists.
	d.fields.RLock()
	postingsMap, ok := d.fields.mapping[name]
	d.fields.RUnlock()
	if ok {
		return postingsMap
	}

	// Acquire write lock and create.
	d.fields.Lock()
	postingsMap, ok = d.fields.mapping[name]

	// Check if it's been created since we last acquired the lock.
	if ok {
		d.fields.Unlock()
		return postingsMap
	}

	postingsMap = newPostingsMap(d.opts)
	d.fields.mapping[name] = postingsMap
	d.fields.Unlock()
	return postingsMap
}
