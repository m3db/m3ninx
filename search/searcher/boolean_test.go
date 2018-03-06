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

package searcher

import (
	"testing"

	"github.com/m3db/m3ninx/postings"
	"github.com/m3db/m3ninx/postings/roaring"
	"github.com/m3db/m3ninx/search"
	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
)

func TestBooleanSearcher(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Must searcher.
	mustPL1 := roaring.NewPostingsList()
	mustPL1.Insert(postings.ID(42))
	mustPL1.Insert(postings.ID(45))
	mustPL1.Insert(postings.ID(50))
	mustPL2 := roaring.NewPostingsList()
	mustPL2.Insert(postings.ID(61))
	mustPL2.Insert(postings.ID(63))
	mustPL2.Insert(postings.ID(67))
	mustSearcher := search.NewMockSearcher(mockCtrl)

	// Should searcher.
	shouldPL1 := roaring.NewPostingsList()
	shouldPL1.Insert(postings.ID(43))
	shouldPL1.Insert(postings.ID(45))
	shouldPL1.Insert(postings.ID(50))
	shouldPL2 := roaring.NewPostingsList()
	shouldPL2.Insert(postings.ID(61))
	shouldPL2.Insert(postings.ID(63))
	shouldPL2.Insert(postings.ID(67))
	shouldSearcher := search.NewMockSearcher(mockCtrl)

	// MustNot searcher.
	mustNotPL1 := roaring.NewPostingsList()
	mustNotPL1.Insert(postings.ID(39))
	mustNotPL1.Insert(postings.ID(44))
	mustNotPL1.Insert(postings.ID(50))
	mustNotPL2 := roaring.NewPostingsList()
	mustNotPL2.Insert(postings.ID(61))
	mustNotPL2.Insert(postings.ID(62))
	mustNotPL2.Insert(postings.ID(67))
	mustNotSearcher := search.NewMockSearcher(mockCtrl)

	gomock.InOrder(
		// The mock Searchers have length 2, corresponding to 2 readers.
		mustSearcher.EXPECT().Len().Return(2),
		shouldSearcher.EXPECT().Len().Return(2),
		mustNotSearcher.EXPECT().Len().Return(2),

		// Get the postings lists for the first Reader.
		mustSearcher.EXPECT().Next().Return(true),
		mustSearcher.EXPECT().Current().Return(mustPL1),
		shouldSearcher.EXPECT().Next().Return(true),
		shouldSearcher.EXPECT().Current().Return(shouldPL1),
		mustNotSearcher.EXPECT().Next().Return(true),
		mustNotSearcher.EXPECT().Current().Return(mustNotPL1),

		// Get the postings lists for the second Reader.
		mustSearcher.EXPECT().Next().Return(true),
		mustSearcher.EXPECT().Current().Return(mustPL2),
		shouldSearcher.EXPECT().Next().Return(true),
		shouldSearcher.EXPECT().Current().Return(shouldPL2),
		mustNotSearcher.EXPECT().Next().Return(true),
		mustNotSearcher.EXPECT().Current().Return(mustNotPL2),

		// Close the searchers.
		mustSearcher.EXPECT().Close().Return(nil),
		shouldSearcher.EXPECT().Close().Return(nil),
		mustNotSearcher.EXPECT().Close().Return(nil),
	)

	s, err := NewBooleanSearcher(mustSearcher, shouldSearcher, mustNotSearcher)
	require.NoError(t, err)

	// Ensure the searcher is searching over two readers.
	require.Equal(t, 2, s.Len())

	// Test the postings list from the first Reader.
	require.True(t, s.Next())

	expected := mustPL1.Clone()
	expected.Intersect(shouldPL1)
	expected.Difference(mustNotPL1)
	require.True(t, s.Current().Equal(expected))

	// Test the postings list from the second Reader.
	require.True(t, s.Next())

	expected = mustPL2.Clone()
	expected.Intersect(shouldPL2)
	expected.Difference(mustNotPL2)
	require.True(t, s.Current().Equal(expected))

	require.False(t, s.Next())
	require.NoError(t, s.Err())

	require.NoError(t, s.Close())
}
