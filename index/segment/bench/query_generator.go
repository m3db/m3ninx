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
	"bytes"

	"github.com/m3db/m3ninx/index/segment"
)

// QueryGeneratorOpts is a collection of knobs to tweak query generation.
type QueryGeneratorOpts struct {
	MaxNumFiltersPerQuery int
	PercentNegated        float64
	PercentRegex          float64
	GeneratorSpec         *GeneratorSpec
}

// QueryGenerator generates random queries.
type QueryGenerator struct {
	opts QueryGeneratorOpts
}

// NewQueryGenerator returns a new query generator.
func NewQueryGenerator(opts QueryGeneratorOpts) *QueryGenerator {
	return &QueryGenerator{opts: opts}
}

// GenerateN generates N queries.
func (q *QueryGenerator) GenerateN(numQueries int) []segment.Query {
	queries := make([]segment.Query, 0, numQueries)
	for i := 0; i < numQueries; i++ {
		queries = append(queries, q.Generate())
	}
	return queries
}

// Generate generates a single query.
func (q *QueryGenerator) Generate() segment.Query {
	numFilters := 1 + q.opts.GeneratorSpec.intn(q.opts.MaxNumFiltersPerQuery)
	filters := make([]segment.Filter, 0, numFilters)
	for i := 0; i < numFilters; i++ {
		filters = append(filters, q.generateFilter())
	}
	return segment.Query{
		Conjunction: segment.AndConjunction,
		Filters:     filters,
	}
}

func (q *QueryGenerator) generateFilter() segment.Filter {
	fieldName := q.opts.GeneratorSpec.generateFieldName()
	regex := q.opts.GeneratorSpec.float64() < q.opts.PercentRegex
	var fieldValue []byte
	if regex {
		fieldValue = q.generateRegex()
	} else {
		fieldValue = q.opts.GeneratorSpec.generateFieldValue()
	}
	negated := q.opts.GeneratorSpec.float64() < q.opts.PercentNegated
	return segment.Filter{
		FieldName:        []byte(fieldName),
		FieldValueFilter: fieldValue,
		Negate:           negated,
		Regexp:           regex,
	}
}

func (q *QueryGenerator) generateRegex() []byte {
	numRegexOrs := 1 + q.opts.GeneratorSpec.intn(20)
	var buf bytes.Buffer
	first := true
	for i := 0; i < numRegexOrs; i++ {
		if !first {
			buf.WriteString("|")
		}
		first = false
		buf.Write(q.opts.GeneratorSpec.generateFieldValue())
	}
	return buf.Bytes()
}
