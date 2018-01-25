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
)

type testCase struct {
	input  []byte
	output [][]byte
}

func TestComputeTrigrams(t *testing.T) {
	testCases := []testCase{
		testCase{
			input: []byte("a=b"),
			output: [][]byte{
				[]byte("a=b"),
			},
		},
		testCase{
			input: []byte("fieldName=fieldValue"),
			output: [][]byte{
				[]byte("eld"),
				[]byte("iel"),
				[]byte("fie"),
				[]byte("lue"),
				[]byte("dNa"),
				[]byte("Val"),
				[]byte("dVa"),
				[]byte("ame"),
				[]byte("ldN"),
				[]byte("Nam"),
				[]byte("e=f"),
				[]byte("me="),
				[]byte("=fi"),
				[]byte("alu"),
				[]byte("ldV"),
			},
		},
	}

	for _, tc := range testCases {
		found := make(map[string]bool)
		for _, t := range tc.output {
			found[string(t)] = false
		}

		trigrams := computeTrigrams(tc.input)
		for _, t := range trigrams {
			found[string(t)] = true
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
	f := []byte("fieldName=fieldValue")
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		computeTrigrams(f)
	}
}
