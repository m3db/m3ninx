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

package bench_test

import (
	"testing"

	"github.com/m3db/m3ninx/index/segment/bench"
	"github.com/m3db/m3ninx/index/segment/mem"
)

const numGeneratedDocs = 10000

func BenchmarkSimpleIndexInsert(b *testing.B) {
	// generate documents
	gs, err := bench.NewGeneratorSpec(1234567)
	if err != nil {
		b.Errorf("%v", err)
	}
	generatedDocuments := gs.GenerateN(numGeneratedDocs)
	opts := mem.NewOptions()
	simpleSegment, err := mem.New(opts)
	if err != nil {
		b.Errorf("%v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := generatedDocuments[i%numGeneratedDocs]
		simpleSegment.Insert(d)
	}
}

func BenchmarkTrigramIndexInsert(b *testing.B) {
	// generate documents
	gs, err := bench.NewGeneratorSpec(1234567)
	if err != nil {
		b.Errorf("%v", err)
	}
	generatedDocuments := gs.GenerateN(numGeneratedDocs)
	opts := mem.NewOptions()
	triSegment := mem.NewTrigram(opts)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := generatedDocuments[i%numGeneratedDocs]
		triSegment.Insert(d)
	}
}
