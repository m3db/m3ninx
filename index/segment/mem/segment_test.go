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
	sgmt "github.com/m3db/m3ninx/index/segment"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type newMutableSegmentFn func() sgmt.MutableSegment

type segmentTestSuite struct {
	suite.Suite

	fn      newMutableSegmentFn
	segment sgmt.MutableSegment
}

func (t *segmentTestSuite) SetupTest() {
	t.segment = t.fn()
}

func (t *segmentTestSuite) TestInsert() {
	err := t.segment.Insert(doc.Document{
		Fields: []doc.Field{
			doc.Field{Name: []byte("abc"), Value: []byte("efg")},
		},
	})
	t.NoError(err)
}

// TODO: Expand segment test suite.

func TestSimpleSegment(t *testing.T) {
	fn := func() sgmt.MutableSegment {
		opts := NewOptions()
		s, err := NewSegment(0, opts)
		require.NoError(t, err)
		return s
	}
	suite.Run(t, &segmentTestSuite{fn: fn})
}
