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
	"github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/index/segment/mem"
)

func BenchmarkSimpleIndexSimpleQuery(b *testing.B) {
	if simpleErr != nil {
		b.Error(simpleErr)
	}
	benchmarkQuery(b, simpleSegment, generatedSimpleQueries)
}

func BenchmarkTrigramIndexSimpleQuery(b *testing.B) {
	if trigramErr != nil {
		b.Error(trigramErr)
	}
	benchmarkQuery(b, trigramSegment, generatedSimpleQueries)
}

func BenchmarkFSTIndexSimpleQuery(b *testing.B) {
	if fstErr != nil {
		b.Error(fstErr)
	}
	benchmarkQuery(b, fstSegment, generatedSimpleQueries)
}

func BenchmarkSimpleIndexSimpleNegatedQuery(b *testing.B) {
	if simpleErr != nil {
		b.Error(simpleErr)
	}
	benchmarkQuery(b, simpleSegment, generatedNegatedQueries)
}

func BenchmarkTrigramIndexSimpleNegatedQuery(b *testing.B) {
	if trigramErr != nil {
		b.Error(trigramErr)
	}
	benchmarkQuery(b, trigramSegment, generatedNegatedQueries)
}

func BenchmarkFSTIndexSimpleNegatedQuery(b *testing.B) {
	if fstErr != nil {
		b.Error(fstErr)
	}
	benchmarkQuery(b, fstSegment, generatedNegatedQueries)
}

func BenchmarkSimpleIndexRegexQuery(b *testing.B) {
	if simpleErr != nil {
		b.Error(simpleErr)
	}
	benchmarkQuery(b, simpleSegment, generatedRegexQueries)
}

func BenchmarkTrigramIndexRegexQuery(b *testing.B) {
	b.Error("takes too long, skipping")
	if trigramErr != nil {
		b.Error(trigramErr)
	}
	benchmarkQuery(b, trigramSegment, generatedRegexQueries)
}

func BenchmarkFSTIndexRegexQuery(b *testing.B) {
	if fstErr != nil {
		b.Error(fstErr)
	}
	benchmarkQuery(b, fstSegment, generatedRegexQueries)
}

func BenchmarkSimpleIndexComplexQuery(b *testing.B) {
	if simpleErr != nil {
		b.Error(simpleErr)
	}
	benchmarkQuery(b, simpleSegment, generatedComplexQueries)
}

func BenchmarkTrigramIndexComplexQuery(b *testing.B) {
	b.Error("takes too long, skipping")
	if trigramErr != nil {
		b.Error(trigramErr)
	}
	benchmarkQuery(b, trigramSegment, generatedComplexQueries)
}

func BenchmarkFSTIndexComplexQuery(b *testing.B) {
	if fstErr != nil {
		b.Error(fstSegment)
	}
	benchmarkQuery(b, fstSegment, generatedComplexQueries)
}

func benchmarkQuery(b *testing.B, seg segment.Readable, queries []segment.Query) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		seg.Query(q)
	}
}

var (
	ingestedDocs    []doc.Document
	numIngestedDocs = 10000
)
var (
	generatedSimpleQueries     []segment.Query
	numGeneratedSimpleQueries  = 1000
	generatedNegatedQueries    []segment.Query
	numGeneratedNegatedQueries = 1000
	generatedRegexQueries      []segment.Query
	numGeneratedRegexQueries   = 1000
	generatedComplexQueries    []segment.Query
	numGeneratedComplexQueries = 1000
)

var (
	simpleSegment, trigramSegment mem.Segment
	fstSegment                    segment.Readable
	simpleErr, trigramErr, fstErr error
)

func init() {
	// generate documents
	gs := GeneratorSpec{
		Rng:                      rand.New(rand.NewSource(1234567)),
		MinNumFieldsPerDoc:       3,
		MaxNumFieldsPerDoc:       10,
		MinNumWordsPerFieldValue: 1,
		MaxNumWordsPerFieldValue: 3,
	}
	ingestedDocs = gs.Generate(numIngestedDocs)

	// create segments
	// simple segment & ingested docs
	segOpts := mem.NewOptions()
	simpleSegment, _ = mem.New(segOpts)
	for i := 0; i < numIngestedDocs; i++ {
		d := ingestedDocs[i]
		simpleErr = simpleSegment.Insert(d)
		if simpleErr != nil {
			break
		}
	}

	// trigram segment & ingested docs
	trigramSegment = mem.NewTrigram(segOpts)
	for i := 0; i < numIngestedDocs; i++ {
		d := ingestedDocs[i]
		trigramErr = trigramSegment.Insert(d)
		if trigramErr != nil {
			break
		}
	}

	// fst segment & ingested docs
	fstSegment, fstErr = mem.NewFST(1, []segment.Segment{simpleSegment}, segOpts)

	// generate exact match queries
	opts := QueryGeneratorOpts{
		MaxNumFiltersPerQuery: 10,
		PercentNegated:        0.0,
		PercentRegex:          0.0,
		Rng:                   gs.Rng,
	}
	queryGen := NewQueryGenerator(generatedDocuments, opts)
	generatedSimpleQueries = queryGen.Generate(numGeneratedSimpleQueries)

	// generate exact match queries with negation
	opts = QueryGeneratorOpts{
		MaxNumFiltersPerQuery: 10,
		PercentNegated:        0.30,
		PercentRegex:          0.0,
		Rng:                   gs.Rng,
	}
	queryGen = NewQueryGenerator(generatedDocuments, opts)
	generatedNegatedQueries = queryGen.Generate(numGeneratedNegatedQueries)

	// generate regex queries with no negation
	opts = QueryGeneratorOpts{
		MaxNumFiltersPerQuery: 10,
		PercentNegated:        0.0,
		PercentRegex:          1,
		Rng:                   gs.Rng,
	}
	queryGen = NewQueryGenerator(generatedDocuments, opts)
	generatedRegexQueries = queryGen.Generate(numGeneratedRegexQueries)

	// generate complex workload with all flags
	opts = QueryGeneratorOpts{
		MaxNumFiltersPerQuery: 10,
		PercentNegated:        0.20,
		PercentRegex:          0.10,
		Rng:                   gs.Rng,
	}
	queryGen = NewQueryGenerator(generatedDocuments, opts)
	generatedComplexQueries = queryGen.Generate(numGeneratedComplexQueries)
}
