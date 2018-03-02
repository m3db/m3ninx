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

func TestBooleanQueryMust(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firstField, firstTerm := []byte("apple"), []byte("red")
	firstQuery := NewTermQuery(firstField, firstTerm)
	firstPostingsList := postings.NewRoaringPostingsList()
	firstPostingsList.Insert(postings.ID(42))
	firstPostingsList.Insert(postings.ID(50))
	firstPostingsList.Insert(postings.ID(57))

	secondField, secondTerm := []byte("banana"), []byte("yellow")
	secondQuery := NewTermQuery(secondField, secondTerm)
	secondPostingsList := postings.NewRoaringPostingsList()
	secondPostingsList.Insert(postings.ID(44))
	secondPostingsList.Insert(postings.ID(50))
	secondPostingsList.Insert(postings.ID(61))

	reader := index.NewMockReader(mockCtrl)
	gomock.InOrder(
		reader.EXPECT().MatchTerm(firstField, firstTerm).Return(firstPostingsList, nil),
		reader.EXPECT().MatchTerm(secondField, secondTerm).Return(secondPostingsList, nil),
	)

	expected := postings.NewRoaringPostingsList()
	expected.Insert(postings.ID(50))

	q, err := NewBooleanQuery([]search.Query{firstQuery, secondQuery}, nil, nil)
	require.NoError(t, err)

	actual, err := q.Execute(reader)
	require.NoError(t, err)
	require.True(t, expected.Equal(actual))
}

func TestBooleanQueryShould(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firstField, firstTerm := []byte("apple"), []byte("red")
	firstQuery := NewTermQuery(firstField, firstTerm)
	firstPostingsList := postings.NewRoaringPostingsList()
	firstPostingsList.Insert(postings.ID(42))
	firstPostingsList.Insert(postings.ID(50))
	firstPostingsList.Insert(postings.ID(57))

	secondName, secondTerm := []byte("banana"), []byte("yellow")
	secondQuery := NewTermQuery(secondName, secondTerm)
	secondPostingsList := postings.NewRoaringPostingsList()
	secondPostingsList.Insert(postings.ID(44))
	secondPostingsList.Insert(postings.ID(50))
	secondPostingsList.Insert(postings.ID(61))

	reader := index.NewMockReader(mockCtrl)
	gomock.InOrder(
		reader.EXPECT().MatchTerm(firstField, firstTerm).Return(firstPostingsList, nil),
		reader.EXPECT().MatchTerm(secondName, secondTerm).Return(secondPostingsList, nil),
	)

	expected := postings.NewRoaringPostingsList()
	expected.Insert(postings.ID(42))
	expected.Insert(postings.ID(44))
	expected.Insert(postings.ID(50))
	expected.Insert(postings.ID(57))
	expected.Insert(postings.ID(61))

	q, err := NewBooleanQuery(nil, []search.Query{firstQuery, secondQuery}, nil)
	require.NoError(t, err)

	actual, err := q.Execute(reader)
	require.NoError(t, err)
	require.True(t, expected.Equal(actual))
}

func TestBooleanQueryMustNot(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firstField, firstTerm := []byte("apple"), []byte("red")
	firstQuery := NewTermQuery(firstField, firstTerm)
	firstPostingsList := postings.NewRoaringPostingsList()
	firstPostingsList.Insert(postings.ID(42))
	firstPostingsList.Insert(postings.ID(50))
	firstPostingsList.Insert(postings.ID(57))

	secondName, secondTerm := []byte("banana"), []byte("yellow")
	secondQuery := NewTermQuery(secondName, secondTerm)
	secondPostingsList := postings.NewRoaringPostingsList()
	secondPostingsList.Insert(postings.ID(44))
	secondPostingsList.Insert(postings.ID(50))
	secondPostingsList.Insert(postings.ID(61))

	reader := index.NewMockReader(mockCtrl)
	gomock.InOrder(
		reader.EXPECT().MatchTerm(firstField, firstTerm).Return(firstPostingsList, nil),
		reader.EXPECT().MatchTerm(secondName, secondTerm).Return(secondPostingsList, nil),
	)

	expected := postings.NewRoaringPostingsList()
	expected.Insert(postings.ID(42))
	expected.Insert(postings.ID(57))

	q, err := NewBooleanQuery([]search.Query{firstQuery}, nil, []search.Query{secondQuery})
	require.NoError(t, err)

	actual, err := q.Execute(reader)
	require.NoError(t, err)
	require.True(t, expected.Equal(actual))
}

func TestBooleanQueryAll(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	firstField, firstTerm := []byte("apple"), []byte("red")
	firstQuery := NewTermQuery(firstField, firstTerm)
	firstPostingsList := postings.NewRoaringPostingsList()
	firstPostingsList.Insert(postings.ID(42))
	firstPostingsList.Insert(postings.ID(50))
	firstPostingsList.Insert(postings.ID(57))

	secondName, secondTerm := []byte("banana"), []byte("yellow")
	secondQuery := NewTermQuery(secondName, secondTerm)
	secondPostingsList := postings.NewRoaringPostingsList()
	secondPostingsList.Insert(postings.ID(44))
	secondPostingsList.Insert(postings.ID(50))
	secondPostingsList.Insert(postings.ID(57))

	thirdName, thirdValue := []byte("banana"), []byte("yellow")
	thirdQuery := NewTermQuery(thirdName, thirdValue)
	thirdPostingsList := postings.NewRoaringPostingsList()
	thirdPostingsList.Insert(postings.ID(39))
	thirdPostingsList.Insert(postings.ID(50))
	thirdPostingsList.Insert(postings.ID(61))

	reader := index.NewMockReader(mockCtrl)
	gomock.InOrder(
		reader.EXPECT().MatchTerm(firstField, firstTerm).Return(firstPostingsList, nil),
		reader.EXPECT().MatchTerm(secondName, secondTerm).Return(secondPostingsList, nil),
		reader.EXPECT().MatchTerm(thirdName, thirdValue).Return(thirdPostingsList, nil),
	)

	expected := postings.NewRoaringPostingsList()
	expected.Insert(postings.ID(57))

	q, err := NewBooleanQuery([]search.Query{firstQuery}, []search.Query{secondQuery}, []search.Query{thirdQuery})
	require.NoError(t, err)

	actual, err := q.Execute(reader)
	require.NoError(t, err)
	require.True(t, expected.Equal(actual))
}

func TestBooleanQueryOnlyMustNot(t *testing.T) {
	query := NewTermQuery([]byte("apple"), []byte("red"))

	_, err := NewBooleanQuery(nil, nil, []search.Query{query})
	require.Error(t, err)
}
