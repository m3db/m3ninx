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

	"github.com/m3db/m3ninx/index/segment"

	"github.com/stretchr/testify/require"
)

func TestPostingsListInsertNew(t *testing.T) {
	p := newPostingsManager(NewOptions())
	id := p.Insert(123)
	bitmap, err := p.Fetch(id)
	require.NoError(t, err)
	require.True(t, bitmap.Contains(123))
}

func TestPostingsListInsertSet(t *testing.T) {
	p := newPostingsManager(NewOptions())
	set := segment.NewPostingsList()
	set.Insert(12)
	set.Insert(23)
	id, err := p.InsertSet(set)
	require.NoError(t, err)
	bitmap, err := p.Fetch(id)
	require.NoError(t, err)
	require.True(t, bitmap.Contains(12))
	require.True(t, bitmap.Contains(23))
}

func TestPostingsListFetchIllegalOffset(t *testing.T) {
	p := newPostingsManager(NewOptions())
	err := p.Update(100, 123)
	require.Error(t, err)
	_, err = p.Fetch(100)
	require.Error(t, err)
}

func TestPostingsListUpdate(t *testing.T) {
	p := newPostingsManager(NewOptions())
	id := p.Insert(123)
	err := p.Update(id, 142)
	require.NoError(t, err)

	bitmap, err := p.Fetch(id)
	require.NoError(t, err)
	require.True(t, bitmap.Contains(123))
	require.True(t, bitmap.Contains(142))
}
func TestPostingsListInsertNewMany(t *testing.T) {
	p := newPostingsManager(NewOptions())
	for i := 0; i < 100; i++ {
		id := p.Insert(segment.DocID(i))
		require.Equal(t, i, int(id))
	}

	for i := 0; i < 100; i++ {
		bitmap, err := p.Fetch(postingsManagerOffset(i))
		require.NoError(t, err)
		require.True(t, bitmap.Contains(segment.DocID(i)))
	}
}
