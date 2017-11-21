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

func TestConcatenateBytes(t *testing.T) {
	type testCase struct {
		inputA []byte
		inputB byte
		inputC []byte
		output []byte
	}

	testCases := []testCase{
		testCase{
			inputA: []byte("abc"),
			inputB: '=',
			inputC: []byte("defg"),
			output: []byte("abc=defg"),
		},
		testCase{
			inputA: []byte("abc"),
			inputB: '=',
			inputC: []byte(""),
			output: []byte("abc="),
		},
		testCase{
			inputA: []byte(""),
			inputB: '=',
			inputC: []byte(""),
			output: []byte("="),
		},
		testCase{
			inputA: []byte(""),
			inputB: '=',
			inputC: []byte("abc"),
			output: []byte("=abc"),
		},
	}

	for _, tc := range testCases {
		o := concatenateBytes(tc.inputA, tc.inputB, tc.inputC)
		require.Equal(t, tc.output, o)
	}
}

func newFSTIndex(t *testing.T, docs []doc.Document) segment.Readable {
	opts := NewOptions()
	simpleSegment, err := New(opts)
	require.NoError(t, err)
	for _, d := range docs {
		require.NoError(t, simpleSegment.Insert(d), "unable to insert document", d)
	}

	fstIndex, err := NewFST(1, []segment.Segment{simpleSegment}, opts)
	require.NoError(t, err)
	return fstIndex
}

func TestFSTSimpleQuery(t *testing.T) {
	fstIndex := newFSTIndex(t, []doc.Document{
		doc.Document{
			ID: []byte("abc=efg"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("abc"), Value: doc.Value("efg")},
			},
		},
	})

	docs, err := fstIndex.Query(segment.Query{
		Conjunction: segment.AndConjunction,
		Filters: []segment.Filter{
			segment.Filter{
				FieldName:        []byte("abc"),
				FieldValueFilter: []byte("efg"),
				Negate:           false,
				Regexp:           false,
			},
		},
	})
	require.NoError(t, err)
	require.True(t, docs != nil)
	require.Equal(t, 1, len(docs))
	require.Equal(t, doc.ID("abc=efg"), docs[0].ID)
}

func TestFSTQueryNegate(t *testing.T) {
	fstIndex := newFSTIndex(t, []doc.Document{
		doc.Document{
			ID: []byte("abc=efg"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("abc"), Value: doc.Value("efg")},
			},
		},
		doc.Document{
			ID: []byte("abc=efgh"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("abc"), Value: doc.Value("efgh")},
			},
		},
	})

	docs, err := fstIndex.Query(segment.Query{
		Conjunction: segment.AndConjunction,
		Filters: []segment.Filter{
			segment.Filter{
				FieldName:        []byte("abc"),
				FieldValueFilter: []byte("efg"),
				Negate:           true,
				Regexp:           false,
			},
			segment.Filter{
				FieldName:        []byte("abc"),
				FieldValueFilter: []byte("efgh"),
				Negate:           false,
				Regexp:           false,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(docs))
	require.Equal(t, doc.ID("abc=efgh"), docs[0].ID)
}

func TestFSTQueryRegex(t *testing.T) {
	fstIndex := newFSTIndex(t, []doc.Document{
		doc.Document{
			ID: []byte("abc=efg"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("abc"), Value: doc.Value("efg")},
			},
		},
		doc.Document{
			ID: []byte("abc=efgh"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("abc"), Value: doc.Value("efgh")},
			},
		},
	})

	docs, err := fstIndex.Query(segment.Query{
		Conjunction: segment.AndConjunction,
		Filters: []segment.Filter{
			segment.Filter{
				FieldName:        []byte("abc"),
				FieldValueFilter: []byte("efg.+"),
				Negate:           false,
				Regexp:           true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(docs))
	require.Equal(t, doc.ID("abc=efgh"), docs[0].ID)
}

func TestFSTQueryRegexNegate(t *testing.T) {
	fstIndex := newFSTIndex(t, []doc.Document{
		doc.Document{
			ID: []byte("abc=efg|foo=bar"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("foo"), Value: doc.Value("bar")},
				doc.Field{Name: []byte("abc"), Value: doc.Value("efg")},
			},
		},
		doc.Document{
			ID: []byte("abc=efgh|foo=baz"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("foo"), Value: doc.Value("bar")},
				doc.Field{Name: []byte("abc"), Value: doc.Value("efgh")},
			},
		},
		doc.Document{
			ID: []byte("foo=bar|baz=efgh"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("foo"), Value: doc.Value("bar")},
				doc.Field{Name: []byte("baz"), Value: doc.Value("efgh")},
			},
		},
	})

	docs, err := fstIndex.Query(segment.Query{
		Conjunction: segment.AndConjunction,
		Filters: []segment.Filter{
			segment.Filter{
				FieldName:        []byte("foo"),
				FieldValueFilter: []byte("bar"),
				Negate:           false,
				Regexp:           false,
			},
			segment.Filter{
				FieldName:        []byte("abc"),
				FieldValueFilter: []byte(".fg.+"),
				Negate:           true,
				Regexp:           true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(docs))
	requiredIds := []doc.ID{
		doc.ID("abc=efg|foo=bar"),
		doc.ID("foo=bar|baz=efgh"),
	}
	for _, req := range requiredIds {
		found := false
		for _, ret := range docs {
			if string(req) == string(ret.ID) {
				found = true
			}
		}
		require.True(t, found, string(req))
	}
}

func TestFSTQueryExactNegate(t *testing.T) {
	fstIndex := newFSTIndex(t, []doc.Document{
		doc.Document{
			ID: []byte("abc=efg|foo=bar"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("foo"), Value: doc.Value("bar")},
				doc.Field{Name: []byte("abc"), Value: doc.Value("efg")},
			},
		},
		doc.Document{
			ID: []byte("abc=efgh|foo=baz"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("foo"), Value: doc.Value("bar")},
				doc.Field{Name: []byte("abc"), Value: doc.Value("efgh")},
			},
		},
		doc.Document{
			ID: []byte("foo=bar|baz=efgh"),
			Fields: []doc.Field{
				doc.Field{Name: []byte("foo"), Value: doc.Value("bar")},
				doc.Field{Name: []byte("baz"), Value: doc.Value("efgh")},
			},
		},
	})

	docs, err := fstIndex.Query(segment.Query{
		Conjunction: segment.AndConjunction,
		Filters: []segment.Filter{
			segment.Filter{
				FieldName:        []byte("foo"),
				FieldValueFilter: []byte("bar"),
				Negate:           false,
				Regexp:           false,
			},
			segment.Filter{
				FieldName:        []byte("abc"),
				FieldValueFilter: []byte("efgh"),
				Negate:           true,
				Regexp:           false,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(docs))
	requiredIds := []doc.ID{
		doc.ID("abc=efg|foo=bar"),
		doc.ID("foo=bar|baz=efgh"),
	}
	for _, req := range requiredIds {
		found := false
		for _, ret := range docs {
			if string(req) == string(ret.ID) {
				found = true
			}
		}
		require.True(t, found, string(req))
	}
}
