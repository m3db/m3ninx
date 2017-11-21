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
	"math/rand"

	"github.com/m3db/m3ninx/index/segment"

	"github.com/m3db/m3ninx/doc"
)

// QueryGeneratorOpts is a collection of knobs to tweak query generation.
type QueryGeneratorOpts struct {
	Rng                   *rand.Rand
	MaxNumFiltersPerQuery int
	PercentNegated        float64
	PercentRegex          float64
}

// QueryGenerator generates random queries.
type QueryGenerator struct {
	opts              QueryGeneratorOpts
	fieldNames        []string
	fieldNameToValues map[string][]string
}

// NewQueryGenerator returns a new query generator.
func NewQueryGenerator(docs []doc.Document, opts QueryGeneratorOpts) *QueryGenerator {
	seenNamesAndValues := make(map[string]map[string]struct{}, len(docs))
	for _, d := range docs {
		for _, f := range d.Fields {
			name := string(f.Name)
			value := string(f.Value)
			valuesMap, ok := seenNamesAndValues[name]
			if !ok {
				valuesMap = make(map[string]struct{}, len(docs))
				seenNamesAndValues[name] = valuesMap
			}
			valuesMap[value] = struct{}{}
		}
	}

	fieldNames := make([]string, 0, len(seenNamesAndValues))
	fieldNameToValues := make(map[string][]string, len(seenNamesAndValues))
	for name, valuesMap := range seenNamesAndValues {
		fieldNames = append(fieldNames, name)
		for value := range valuesMap {
			fieldNameToValues[name] = append(fieldNameToValues[name], value)
		}
	}

	return &QueryGenerator{
		opts:              opts,
		fieldNames:        fieldNames,
		fieldNameToValues: fieldNameToValues,
	}
}

// Generate generates queries.
func (q *QueryGenerator) Generate(numQueries int) []segment.Query {
	queries := make([]segment.Query, 0, numQueries)
	for i := 0; i < numQueries; i++ {
		queries = append(queries, q.generate())
	}
	return queries
}

func (q *QueryGenerator) generate() segment.Query {
	numFilters := 1 + q.opts.Rng.Intn(q.opts.MaxNumFiltersPerQuery)
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
	numFields := len(q.fieldNames)
	fieldName := q.fieldNames[q.opts.Rng.Intn(numFields)]
	regex := q.opts.Rng.Float64() < q.opts.PercentRegex
	var fieldValue []byte
	if regex {
		fieldValue = q.generateRegexForField(fieldName)
	} else {
		numValues := len(q.fieldNameToValues[fieldName])
		value := q.fieldNameToValues[fieldName][q.opts.Rng.Intn(numValues)]
		fieldValue = []byte(value)
	}
	negated := q.opts.Rng.Float64() < q.opts.PercentNegated
	return segment.Filter{
		FieldName:        []byte(fieldName),
		FieldValueFilter: fieldValue,
		Negate:           negated,
		Regexp:           regex,
	}
}

func (q *QueryGenerator) generateRegexForField(fieldName string) []byte {
	var buf bytes.Buffer
	numValues := len(q.fieldNameToValues[fieldName])
	numRegexOrs := 1 + q.opts.Rng.Intn(numValues)
	first := true
	for i := 0; i < numRegexOrs; i++ {
		if !first {
			buf.WriteString("|")
		}
		first = false
		n := q.opts.Rng.Intn(numValues)
		buf.WriteString(q.fieldNameToValues[fieldName][n])
	}
	return buf.Bytes()
}
