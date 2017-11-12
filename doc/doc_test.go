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

package doc_test

import (
	"testing"

	"github.com/m3db/m3ninx/doc"

	"github.com/stretchr/testify/require"
)

func TestNewMetricWithNoBytes(t *testing.T) {
	metricID := []byte("some-random-id")
	tags := map[string]string{
		"abc": "one",
		"def": "two",
	}

	d := doc.New(metricID, tags)
	require.Equal(t, metricID, []byte(d.ID))

	found := 0
	for _, f := range d.Fields {
		require.Equal(t, doc.StringValueType, f.ValueType)
		name, value := string(f.Name), string(f.Value)
		if tags[name] == value {
			found++
		}
	}
	require.Equal(t, len(tags), found)
}
