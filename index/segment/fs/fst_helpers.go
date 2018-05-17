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
	"fmt"
	"unicode/utf8"

	"github.com/m3db/m3ninx/doc"
)

var (
	privateCodePoint    = doc.ReservedCodePoint
	privateCodePointLen = len(privateCodePoint)

	minByteKey = []byte{}
	maxByteKey = []byte(string(utf8.MaxRune))
)

func extractFieldAndTerm(entry []byte) (fieldAndTerm, error) {
	result := fieldAndTerm{}
	idx := bytes.Index(entry, privateCodePoint)
	if idx == -1 {
		return result, fmt.Errorf("invalid entry, did not find separator")
	}
	result.field = entry[:idx]
	if idx+privateCodePointLen > len(entry) {
		return result, fmt.Errorf("entry is too small, did not find term")

	}
	result.term = entry[idx+privateCodePointLen:]
	return result, nil
}

func computeFSTKeyWithField(field []byte) []byte {
	return append(
		append([]byte(nil), field...), privateCodePoint...)
}

func computeFSTKey(ft fieldAndTerm) []byte {
	return append(computeFSTKeyWithField(ft.field), ft.term...)
}

func computeFSTBoundsForField(field []byte) (min []byte, max []byte) {
	min = computeFSTKeyWithField(field)
	max = append(min, maxByteKey...)
	return min, max
}
