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

	"github.com/golang/mock/gomock"
	"github.com/m3db/m3ninx/index"
	"github.com/m3db/m3ninx/search"

	"github.com/stretchr/testify/require"
)

func TestBooleanQuery(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firstMockReader := index.NewMockReader(mockCtrl)
	secondMockReader := index.NewMockReader(mockCtrl)
	thirdMockReader := index.NewMockReader(mockCtrl)
	fourthMockReader := index.NewMockReader(mockCtrl)

	tests := []struct {
		name                  string
		must, should, mustNot []search.Query
		readers               index.Readers
		expectErr             bool
	}{
		{
			name:    "no queries provided",
			readers: index.Readers{firstMockReader},
		},
		{
			name: "must, should, and mustNot queries provided",
			must: []search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
			},
			should: []search.Query{
				NewTermQuery([]byte("color"), []byte("red")),
			},
			mustNot: []search.Query{
				NewTermQuery([]byte("variety"), []byte("red delicious")),
			},
			readers: index.Readers{secondMockReader},
		},
		{
			name: "only a must query provided",
			should: []search.Query{
				NewTermQuery([]byte("color"), []byte("red")),
			},
			readers: index.Readers{thirdMockReader},
		},
		{
			name: "only a should query provided",
			should: []search.Query{
				NewTermQuery([]byte("color"), []byte("red")),
			},
			readers: index.Readers{fourthMockReader},
		},
		{
			name: "a mustNot query without an accompanying must or should query is invalid",
			mustNot: []search.Query{
				NewTermQuery([]byte("variety"), []byte("red delicious")),
			},
			expectErr: true,
		},
	}

	gomock.InOrder(
		// When no queries are provided the readers are closed because an empty searcher doesn't
		// take a reference to them.
		firstMockReader.EXPECT().Close().Return(nil),

		// With multiple queries we need to pass a clone to each Searcher.
		secondMockReader.EXPECT().Clone().Return(nil, nil),
		secondMockReader.EXPECT().Clone().Return(nil, nil),
		secondMockReader.EXPECT().Clone().Return(nil, nil),
		secondMockReader.EXPECT().Close().Return(nil),
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q, err := NewBooleanQuery(test.must, test.should, test.mustNot)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			_, err = q.Searcher(test.readers)
			require.NoError(t, err)
		})
	}
}
