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
	"testing"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"

	"github.com/stretchr/testify/require"
)

func newTestOptions() Options {
	return NewOptions()
}

func TestNewMemSegment(t *testing.T) {
	opts := newTestOptions()
	idx, err := New(opts)
	require.NoError(t, err)

	metricID := []byte("some-random-id")
	tags := map[string]string{
		"abc": "one",
		"def": "two",
	}

	d := doc.New(metricID, tags)
	require.NoError(t, idx.Insert(d))

	docs, err := idx.Query(segment.Query{
		Conjunction: segment.AndConjunction,
		Filters: []segment.Filter{
			segment.Filter{
				FieldName:        []byte("abc"),
				FieldValueFilter: []byte("one"),
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, 1, len(docs))
	require.Equal(t, metricID, []byte(docs[0].ID))
}
