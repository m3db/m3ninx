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

	"github.com/stretchr/testify/require"

	"github.com/m3db/m3ninx/doc"
)

type testCase struct {
	input  doc.Field
	output []string
}

func TestComputeTrigrams(t *testing.T) {
	testCases := []testCase{
		testCase{
			input: doc.Field{
				Name:  []byte("a"),
				Value: doc.Value("b"),
			},
			output: []string{"a=b"},
		},
		testCase{
			input: doc.Field{
				Name:  []byte("fieldName"),
				Value: doc.Value("fieldValue"),
			},
			output: []string{"eld", "iel", "fie", "lue", "dNa", "Val", "dVa", "ame", "ldN", "Nam", "e=f", "me=", "=fi", "alu", "ldV"},
		},
	}

	for _, tc := range testCases {
		found := make(map[string]bool)
		for _, t := range tc.output {
			found[t] = false
		}

		trigrams := computeTrigrams(tc.input)
		for _, t := range trigrams {
			found[t] = true
		}

		missing := []string{}
		for tri, ok := range found {
			if !ok {
				missing = append(missing, tri)
			}
		}
		require.Empty(t, missing, "Did not find expected trigrams", missing)
	}
}

func BenchmarkComputeTrigram(b *testing.B) {
	f := doc.Field{
		Name:  []byte("fieldName"),
		Value: doc.Value("fieldValue"),
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		computeTrigrams(f)
	}
}
