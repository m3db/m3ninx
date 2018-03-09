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

func TestDisjunctionQuery(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firstMockSnapshot := index.NewMockSnapshot(mockCtrl)
	secondMockSnapshot := index.NewMockSnapshot(mockCtrl)
	thirdMockSnapshot := index.NewMockSnapshot(mockCtrl)

	tests := []struct {
		name     string
		queries  []search.Query
		snapshot index.Snapshot
	}{
		{
			name:     "no queries provided",
			snapshot: firstMockSnapshot,
		},
		{
			name: "a single query provided",
			queries: []search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
			},
			snapshot: secondMockSnapshot,
		},
		{
			name: "multiple queries provided",
			queries: []search.Query{
				NewTermQuery([]byte("fruit"), []byte("apple")),
				NewTermQuery([]byte("vegetable"), []byte("carrot")),
			},
			snapshot: thirdMockSnapshot,
		},
	}

	gomock.InOrder(
		firstMockSnapshot.EXPECT().Size().Return(4, nil),

		secondMockSnapshot.EXPECT().Readers().Return(nil, nil),

		thirdMockSnapshot.EXPECT().Readers().Return(nil, nil),
		thirdMockSnapshot.EXPECT().Readers().Return(nil, nil),
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := newDisjuctionQuery(test.queries)
			_, err := q.Searcher(test.snapshot)
			require.NoError(t, err)
		})
	}
}
