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
	"testing"

	"github.com/m3db/m3ninx/doc"

	"github.com/stretchr/testify/suite"
)

type newTrigramTermsDictFn func() *trigramTermsDictionary

type trigramTermsDictionaryTestSuite struct {
	suite.Suite

	fn        newTrigramTermsDictFn
	termsDict *trigramTermsDictionary
}

func (t *trigramTermsDictionaryTestSuite) SetupTest() {
	t.termsDict = t.fn()
}

func (t *trigramTermsDictionaryTestSuite) TestInsert() {
	err := t.termsDict.Insert(
		doc.Field{
			Name:  []byte("foo"),
			Value: []byte("bar"),
		},
		1,
	)
	t.Require().NoError(err)

	pl, err := t.termsDict.MatchExact([]byte("foo"), []byte("bar"))
	t.Require().NoError(err)
	t.Require().NotNil(pl)
	t.Equal(uint64(1), pl.Size())
	t.True(pl.Contains(1))
}

func (t *trigramTermsDictionaryTestSuite) TestMatchRegex() {
	err := t.termsDict.Insert(doc.Field{
		Name:  []byte("foo"),
		Value: []byte("bar-1"),
	}, 1)
	t.Require().NoError(err)
	err = t.termsDict.Insert(doc.Field{
		Name:  []byte("foo"),
		Value: []byte("bar-2"),
	}, 2)
	t.Require().NoError(err)

	pattern := "bar-.*"
	re := regexp.MustCompile(pattern)
	pl, err := t.termsDict.MatchRegex([]byte("foo"), []byte(pattern), re)
	t.Require().NoError(err)
	t.Require().NotNil(pl)
	t.Equal(uint64(2), pl.Size())
	t.True(pl.Contains(1))
	t.True(pl.Contains(2))
}

func TestTrigramTermsDictionary(t *testing.T) {
	opts := NewOptions()
	suite.Run(t, &trigramTermsDictionaryTestSuite{
		fn: func() *trigramTermsDictionary {
			return newTrigramTermsDictionary(opts)
		},
	})
}
