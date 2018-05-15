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

package fs

import (
	"bytes"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/stretchr/testify/require"
)

func TestFSTHelpersBijective(t *testing.T) {
	a := []byte("abcd")
	b := []byte("cdef")
	entry := computeFSTKey(fieldAndTerm{a, b})
	ft, err := extractFieldAndTerm(entry)
	require.NoError(t, err)
	require.Equal(t, a, ft.field)
	require.Equal(t, b, ft.term)
}

func TestFSTExtractEmptyKey(t *testing.T) {
	a := []byte("abcd")
	b := []byte{}
	entry := computeFSTKey(fieldAndTerm{a, b})
	ft, err := extractFieldAndTerm(entry)
	require.NoError(t, err)
	require.Equal(t, a, ft.field)
	require.Equal(t, b, ft.term)
}

func TestComputeFSTBounds(t *testing.T) {
	a := []byte("abcd")
	min, max := computeFSTBoundsForField(a)
	require.True(t, bytes.Compare(a, min) < 0, string(min))
	require.True(t, bytes.Compare(min, max) < 0, string(min))
	require.True(t, bytes.Compare(a, max) < 0, string(max))
}

func TestFSTHelperBijectiveProp(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	seed := time.Now().UnixNano()
	parameters.MinSuccessfulTests = 10000
	parameters.MaxSize = 40
	parameters.Rng = rand.New(rand.NewSource(seed))
	parameters.Workers = 1
	properties := gopter.NewProperties(parameters)

	properties.Property("fst key -> extract is 1-1", prop.ForAll(
		func(field []byte, term []byte) (bool, error) {
			entry := computeFSTKey(fieldAndTerm{field, term})
			ft, err := extractFieldAndTerm(entry)
			if err != nil {
				return false, err
			}
			return bytes.Equal(ft.field, field) && bytes.Equal(ft.term, term), nil
		},
		genBytes(), genBytes(),
	))

	properties.Property("fst field key compute is 1-1", prop.ForAll(
		func(field []byte) (bool, error) {
			entry := computeFSTKey(fieldAndTerm{field, nil})
			ft, err := extractFieldAndTerm(entry)
			if err != nil {
				return false, err
			}
			return bytes.Equal(ft.field, field), nil
		},
		genBytes(),
	))

	reporter := gopter.NewFormatedReporter(true, 160, os.Stdout)
	if !properties.Run(reporter) {
		t.Errorf("failed with initial seed: %d", seed)
	}
}

func genBytes() gopter.Gen {
	return gen.AnyString().
		Map(func(s string) []byte {
			return []byte(s)
		}).
		SuchThat(func(b []byte) bool {
			return len(b) > 0 && bytes.Index(b, privateCodePoint) == -1
		})
}
