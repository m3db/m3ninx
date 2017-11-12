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
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"

	"github.com/m3db/m3ninx/doc"
)

var words []string

// GeneratorSpec provides knobs to generate random documents.
type GeneratorSpec struct {
	Rng                      *rand.Rand
	MinNumFieldsPerDoc       int
	MaxNumFieldsPerDoc       int
	MinNumWordsPerFieldValue int
	MaxNumWordsPerFieldValue int

	words []string
}

// Generate generates the specified number of documnets using the rng and
// the embedded `words.txt` file.
func (g GeneratorSpec) Generate(numDocs int) []doc.Document {
	docs := make([]doc.Document, 0, numDocs)
	for i := 0; i < numDocs; i++ {
		docs = append(docs, g.generate())
	}
	return docs
}

func (g GeneratorSpec) generate() doc.Document {
	numFields := g.generateUniform(g.MinNumFieldsPerDoc, g.MaxNumFieldsPerDoc)
	fields := make([]doc.Field, 0, numFields)
	for i := 0; i < numFields; i++ {
		fields = append(fields, g.generateField())
	}
	id := g.generateID(fields)
	return doc.Document{
		ID:     id,
		Fields: fields,
	}
}

func (g GeneratorSpec) generateID(fields []doc.Field) doc.ID {
	var (
		buf   bytes.Buffer
		first = true
	)
	for _, f := range fields {
		if !first {
			buf.WriteString(",")
		}
		first = false
		buf.WriteString(fmt.Sprintf("%s=%s", string(f.Name), string(f.Value)))
	}
	return doc.ID(buf.Bytes())
}

func (g GeneratorSpec) generateField() doc.Field {
	var buf bytes.Buffer
	fieldName := g.generateWord()
	numWords := g.generateUniform(g.MinNumWordsPerFieldValue, g.MaxNumWordsPerFieldValue)
	for i := 0; i < numWords; i++ {
		buf.WriteString(g.generateWord())
	}

	return doc.Field{
		Name:      []byte(fieldName),
		Value:     buf.Bytes(),
		ValueType: doc.StringValueType,
	}
}

func (g GeneratorSpec) generateUniform(min, max int) int {
	return min + g.Rng.Intn(max-min)
}

func (g GeneratorSpec) generateWord() string {
	n := g.Rng.Intn(len(words))
	return words[n]
}

func init() {
	f, err := os.Open("words.txt")
	if err != nil {
		println(err)
		os.Exit(1)
	}
	fs := bufio.NewScanner(f)
	for fs.Scan() {
		words = append(words, fs.Text())
	}
}
