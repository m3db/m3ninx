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

package codec

import (
	"bytes"
	"io"
	"testing"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/util"

	"github.com/stretchr/testify/require"
)

func TestDocumentsRoundtrip(t *testing.T) {
	tests := []struct {
		docs []doc.Document
	}{
		{
			docs: util.MustReadDocs("../util/testdata/node_exporter.json", 2000),
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)

		w := newDocWriter(buf)
		require.NoError(t, w.Init())
		for i := 0; i < len(test.docs); i++ {
			require.NoError(t, w.Write(test.docs[i]))
		}
		require.NoError(t, w.Close())

		r := newDocReader(buf.Bytes())
		require.NoError(t, r.Init())
		for i := 0; i < len(test.docs); i++ {
			actual, err := r.Read()
			require.NoError(t, err)
			require.Equal(t, test.docs[i], actual)
		}
		_, err := r.Read()
		require.Equal(t, io.EOF, err)
		require.NoError(t, r.Close())
	}
}
