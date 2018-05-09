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
	"testing"

	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/search"

	"github.com/stretchr/testify/require"
)

func TestDisjunctionQuery(t *testing.T) {
	tests := []struct {
		name    string
		queries []search.Query
	}{
		{
			name: "no queries provided",
		},
		{
			name: "a single query provided",
			queries: []search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
			},
		},
		{
			name: "multiple queries provided",
			queries: []search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
				NewTermQuery([]byte("vegetable"), []byte("carrot")),
			},
		},
	}

	rs := index.Readers{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := NewDisjuctionQuery(test.queries)
			_, err := q.Searcher(rs)
			require.NoError(t, err)
		})
	}
}

func TestDisjunctionQueryEqual(t *testing.T) {
	tests := []struct {
		name        string
		left, right search.Query
		expected    bool
	}{
		{
			name:     "empty queries",
			left:     NewDisjuctionQuery(nil),
			right:    NewDisjuctionQuery(nil),
			expected: true,
		},
		{
			name: "equal queries",
			left: NewDisjuctionQuery([]search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
				NewTermQuery([]byte("fruit"), []byte("banana")),
			}),
			right: NewDisjuctionQuery([]search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
				NewTermQuery([]byte("fruit"), []byte("banana")),
			}),
			expected: true,
		},
		{
			name: "different order",
			left: NewDisjuctionQuery([]search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
				NewTermQuery([]byte("fruit"), []byte("banana")),
			}),
			right: NewDisjuctionQuery([]search.Query{
				NewTermQuery([]byte("fruit"), []byte("banana")),
				NewTermQuery([]byte("fruit"), []byte("apple")),
			}),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.left.Equal(test.right))
		})
	}
}
