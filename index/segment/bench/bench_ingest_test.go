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

package bench

import (
	"math/rand"
	"testing"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment/mem"
)

func BenchmarkSimpleIndexInsert(b *testing.B) {
	opts := mem.NewOptions()
	simpleSegment, _ := mem.New(opts)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := generatedDocuments[i%numGeneratedDocs]
		simpleSegment.Insert(d)
	}
}

func BenchmarkTrigramIndexInsert(b *testing.B) {
	opts := mem.NewOptions()
	triSegment := mem.NewTrigram(opts)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := generatedDocuments[i%numGeneratedDocs]
		triSegment.Insert(d)
	}
}

var generatedDocuments []doc.Document
var numGeneratedDocs = 10000

func init() {
	// generate documents
	gs := GeneratorSpec{
		Rng:                      rand.New(rand.NewSource(1234567)),
		MinNumFieldsPerDoc:       3,
		MaxNumFieldsPerDoc:       10,
		MinNumWordsPerFieldValue: 1,
		MaxNumWordsPerFieldValue: 3,
	}
	generatedDocuments = gs.Generate(numGeneratedDocs)
}
