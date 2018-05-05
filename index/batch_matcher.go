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

package index

import (
	"fmt"

	"github.com/m3db/m3ninx/doc"

	"github.com/golang/mock/gomock"
)

type BatchMatcher interface {
	gomock.Matcher
}

func NewBatchMatcher(b Batch) BatchMatcher {
	return batchMatcher{b}
}

type batchMatcher struct {
	b Batch
}

func (bm batchMatcher) Matches(x interface{}) bool {
	otherBatch, ok := x.(Batch)
	if !ok {
		return false
	}

	if bm.b.AllowPartialUpdates != otherBatch.AllowPartialUpdates {
		return false
	}

	if len(bm.b.Docs) != len(otherBatch.Docs) {
		return false
	}

	for i := range bm.b.Docs {
		exp := doc.NewDocumentMatcher(bm.b.Docs[i])
		obs := otherBatch.Docs[i]
		if !exp.Matches(obs) {
			return false
		}
	}
	return true
}

func (bm batchMatcher) String() string {
	return fmt.Sprintf("BatchMatcher: %v", bm.b)
}
