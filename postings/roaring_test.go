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

package postings

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoaringPostingsListEmpty(t *testing.T) {
	d := NewRoaringPostingsList()
	require.True(t, d.IsEmpty())
	require.Equal(t, uint64(0), d.Size())
}

func TestRoaringPostingsListInsert(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())
	// Idempotency of inserts.
	require.NoError(t, d.Insert(1))
	require.Equal(t, uint64(1), d.Size())
	require.True(t, d.Contains(1))
}

func TestRoaringPostingsListClone(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())

	c := d.Clone()
	require.True(t, c.Contains(1))
	require.Equal(t, uint64(1), c.Size())

	// Ensure only clone is uniquely backed.
	require.NoError(t, c.Insert(2))
	require.True(t, c.Contains(2))
	require.Equal(t, uint64(2), c.Size())
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())
}

func TestRoaringPostingsListIntersect(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())

	c := d.Clone()
	require.True(t, c.Contains(1))

	require.NoError(t, d.Insert(2))
	require.NoError(t, c.Insert(3))

	require.NoError(t, d.Intersect(c))
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())
	require.True(t, c.Contains(1))
	require.True(t, c.Contains(3))
	require.Equal(t, uint64(2), c.Size())
}

func TestRoaringPostingsListDifference(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())

	c := d.Clone()
	require.True(t, c.Contains(1))

	require.NoError(t, d.Insert(2))
	require.NoError(t, d.Insert(3))
	require.NoError(t, d.Difference(c))

	require.False(t, d.Contains(1))
	require.True(t, c.Contains(1))
	require.Equal(t, uint64(2), d.Size())
	require.Equal(t, uint64(1), c.Size())
	require.True(t, d.Contains(3))
	require.True(t, d.Contains(2))
}

func TestRoaringPostingsListUnion(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())

	c := d.Clone()
	require.True(t, c.Contains(1))
	require.NoError(t, d.Insert(2))
	require.NoError(t, c.Insert(3))

	require.NoError(t, d.Union(c))
	require.True(t, d.Contains(1))
	require.True(t, d.Contains(2))
	require.True(t, d.Contains(3))
	require.Equal(t, uint64(3), d.Size())
	require.True(t, c.Contains(1))
	require.True(t, c.Contains(3))
	require.Equal(t, uint64(2), c.Size())
}

func TestRoaringPostingsListRemoveRange(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.NoError(t, d.Insert(2))
	require.NoError(t, d.Insert(7))
	require.NoError(t, d.Insert(9))

	require.NoError(t, d.RemoveRange(2, 8))
	require.Equal(t, uint64(2), d.Size())
	require.True(t, d.Contains(1))
	require.False(t, d.Contains(2))
	require.False(t, d.Contains(7))
	require.True(t, d.Contains(9))
}

func TestRoaringPostingsListReset(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.True(t, d.Contains(1))
	require.Equal(t, uint64(1), d.Size())
	d.Reset()
	require.True(t, d.IsEmpty())
	require.Equal(t, uint64(0), d.Size())
}

func TestRoaringPostingsListIter(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.NoError(t, d.Insert(2))
	require.Equal(t, uint64(2), d.Size())

	it := d.Iterator()
	found := map[ID]bool{
		1: false,
		2: false,
	}
	for it.Next() {
		found[it.Current()] = true
	}

	for id, ok := range found {
		require.True(t, ok, id)
	}
}

func TestRoaringPostingsListIterInsertAfter(t *testing.T) {
	d := NewRoaringPostingsList()
	require.NoError(t, d.Insert(1))
	require.NoError(t, d.Insert(2))
	require.Equal(t, uint64(2), d.Size())

	it := d.Iterator()
	numElems := 0
	d.Insert(3)
	require.Equal(t, uint64(3), d.Size())
	found := map[ID]bool{
		1: false,
		2: false,
	}
	for it.Next() {
		found[it.Current()] = true
		numElems++
	}

	for id, ok := range found {
		require.True(t, ok, id)
	}
	require.Equal(t, 2, numElems)
}
