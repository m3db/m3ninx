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

package fields

import (
	"bytes"
	"testing"

	"github.com/m3db/m3ninx/postings"
	"github.com/stretchr/testify/require"
)

func TestStoredFieldsIndex(t *testing.T) {
	tests := []struct {
		name    string
		base    postings.ID
		offsets []uint64
	}{
		{
			name:    "valid offsets",
			base:    postings.ID(0),
			offsets: []uint64{13, 29, 42, 57, 83},
		},
		{
			name:    "valid offsets with non-zero base",
			base:    postings.ID(176),
			offsets: []uint64{13, 29, 42, 57, 83},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			w := NewIndexWriter(buf)
			for i := range test.offsets {
				id := test.base + postings.ID(i)
				err := w.Write(id, test.offsets[i])
				require.NoError(t, err)
			}

			r, err := NewIndexReader(buf.Bytes())
			require.NoError(t, err)
			for i := range test.offsets {
				id := test.base + postings.ID(i)
				actual, err := r.Read(id)
				require.NoError(t, err)
				require.Equal(t, test.offsets[i], actual)
			}
		})
	}
}
