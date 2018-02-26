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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestExactQuery(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	name, value := []byte("apple"), []byte("red")

	postingsList := postings.NewRoaringPostingsList()
	postingsList.Insert(postings.ID(42))
	postingsList.Insert(postings.ID(50))
	postingsList.Insert(postings.ID(57))

	reader := index.NewMockReader(mockCtrl)
	gomock.InOrder(
		reader.EXPECT().MatchExact(name, value).Return(postingsList, nil),
	)

	q := NewExactQuery(name, value)
	actual, err := q.Execute(reader)
	require.NoError(t, err)
	require.True(t, postingsList.Equal(actual))
}
