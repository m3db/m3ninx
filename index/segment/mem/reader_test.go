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

package mem

import (
	"regexp"
	"sync"
	"testing"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/postings"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestReaderMatchExact(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	maxID := postings.ID(55)

	name, value := []byte("apple"), []byte("red")
	postingsList := postings.NewRoaringPostingsList()
	postingsList.Insert(postings.ID(42))
	postingsList.Insert(postings.ID(50))
	postingsList.Insert(postings.ID(57))

	segment := NewMockreadableSegment(mockCtrl)
	gomock.InOrder(
		segment.EXPECT().matchExact(name, value).Return(postingsList, nil),
	)

	reader := newReader(segment, maxID, new(sync.WaitGroup))

	actual, err := reader.MatchExact(name, value)
	require.NoError(t, err)
	require.True(t, postingsList.Equal(actual))
}

func TestReaderMatchRegex(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	maxID := postings.ID(55)

	name, pattern := []byte("apple"), []byte("r.*")
	re := regexp.MustCompile(string(pattern))
	postingsList := postings.NewRoaringPostingsList()
	postingsList.Insert(postings.ID(42))
	postingsList.Insert(postings.ID(50))
	postingsList.Insert(postings.ID(57))

	segment := NewMockreadableSegment(mockCtrl)
	gomock.InOrder(
		segment.EXPECT().matchRegex(name, pattern, re).Return(postingsList, nil),
	)

	reader := newReader(segment, maxID, new(sync.WaitGroup))

	actual, err := reader.MatchRegex(name, pattern, re)
	require.NoError(t, err)
	require.True(t, postingsList.Equal(actual))
}

func TestReaderDocs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	maxID := postings.ID(50)
	docs := []doc.Document{
		doc.Document{
			Fields: []doc.Field{
				doc.Field{
					Name:  []byte("apple"),
					Value: []byte("red"),
				},
			},
		},
		doc.Document{
			Fields: []doc.Field{
				doc.Field{
					Name:  []byte("banana"),
					Value: []byte("yellow"),
				},
			},
		},
	}

	segment := NewMockreadableSegment(mockCtrl)
	gomock.InOrder(
		segment.EXPECT().getDoc(postings.ID(42)).Return(docs[0], nil),
		segment.EXPECT().getDoc(postings.ID(47)).Return(docs[1], nil),
	)

	postingsList := postings.NewRoaringPostingsList()
	postingsList.Insert(postings.ID(42))
	postingsList.Insert(postings.ID(47))
	postingsList.Insert(postings.ID(57)) // IDs past maxID should be ignored.

	reader := newReader(segment, maxID, new(sync.WaitGroup))
	defer reader.Close()

	iter, err := reader.Docs(postingsList, nil)
	require.NoError(t, err)

	actualDocs := make([]doc.Document, 0, len(docs))
	for iter.Next() {
		actualDocs = append(actualDocs, iter.Current())
	}

	require.NoError(t, iter.Err())
	require.Equal(t, docs, actualDocs)
}
