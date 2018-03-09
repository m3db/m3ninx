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

package query

import (
	re "regexp"

	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/search"
	"github.com/m3db/m3ninx/search/searcher"
)

// regexpQuery finds documents which match the given regular expression.
type regexpQuery struct {
	field    []byte
	regexp   []byte
	compiled *re.Regexp
}

// NewRegexpQuery constructs a new query for the given regular expression.
func NewRegexpQuery(field, regexp []byte) (search.Query, error) {
	compiled, err := re.Compile(string(regexp))
	if err != nil {
		return nil, err
	}

	return &regexpQuery{
		field:    field,
		regexp:   regexp,
		compiled: compiled,
	}, nil
}

func (q *regexpQuery) Searcher(s index.Snapshot) (search.Searcher, error) {
	rs, err := s.Readers()
	if err != nil {
		return nil, err
	}
	return searcher.NewRegexpSearcher(rs, q.field, q.regexp, q.compiled), nil
}
