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
	"github.com/m3db/m3ninx/postings"
	"github.com/m3db/m3ninx/search"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestConjunctionQuery(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firstName, firstValue := []byte("apple"), []byte("red")
	firstQuery := NewExactQuery(firstName, firstValue)
	firstPostingsList := postings.NewRoaringPostingsList()
	firstPostingsList.Insert(postings.ID(42))
	firstPostingsList.Insert(postings.ID(50))
	firstPostingsList.Insert(postings.ID(57))

	secondName, secondValue := []byte("banana"), []byte("yellow")
	secondQuery := NewExactQuery(secondName, secondValue)
	secondPostingsList := postings.NewRoaringPostingsList()
	secondPostingsList.Insert(postings.ID(44))
	secondPostingsList.Insert(postings.ID(50))
	secondPostingsList.Insert(postings.ID(61))

	reader := index.NewMockReader(mockCtrl)
	gomock.InOrder(
		reader.EXPECT().MatchExact(firstName, firstValue).Return(firstPostingsList, nil),
		reader.EXPECT().MatchExact(secondName, secondValue).Return(secondPostingsList, nil),
	)

	expected := postings.NewRoaringPostingsList()
	expected.Insert(postings.ID(50))

	q := newConjuctionQuery([]search.Query{firstQuery, secondQuery})
	actual, err := q.Execute(reader)
	require.NoError(t, err)
	require.True(t, expected.Equal(actual))
}

func TestConjunctionQueryNoQueries(t *testing.T) {
	q := newConjuctionQuery(nil)
	results, err := q.Execute(nil)
	require.NoError(t, err)
	require.Equal(t, 0, results.Len())
}
