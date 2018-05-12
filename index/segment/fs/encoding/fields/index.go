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
	"fmt"
	"io"

	"github.com/m3db/m3ninx/index/segment/fs/encoding"
	"github.com/m3db/m3ninx/postings"
)

const (
	indexHeaderSize = 8 // Base postings ID as a uint64.

	initialIndexEncoderLen     = 1024
	indexEncoderWriteThreshold = 512
)

// IndexWriter ...
type IndexWriter struct {
	writer io.Writer
	enc    *encoding.Encoder
	ready  bool
	prev   postings.ID
}

// NewIndexWriter ...
func NewIndexWriter(w io.Writer) *IndexWriter {
	iw := &IndexWriter{
		writer: w,
		enc:    encoding.NewEncoder(initialIndexEncoderLen),
	}
	return iw
}

func (w *IndexWriter) Write(id postings.ID, offset uint64) error {
	if !w.ready {
		w.enc.PutUint64(uint64(id))
		w.ready = true
	} else {
		if id != w.prev+1 {
			return fmt.Errorf("non-sequential postings ID: receieved %v but previous ID was %v", id, w.prev)
		}
	}

	w.enc.PutUint64(offset)
	w.prev = id

	if w.enc.Len() > indexEncoderWriteThreshold {
		if err := w.write(); err != nil {
			return err
		}
	}

	return nil
}

func (w *IndexWriter) write() error {
	b := w.enc.Bytes()
	n, err := w.writer.Write(b)
	if err != nil {
		return err
	}
	if n < len(b) {
		return io.ErrShortWrite
	}
	w.enc.Reset()
	return nil
}

// Close ...
func (w *IndexWriter) Close() error {
	if w.enc.Len() == 0 {
		return nil
	}
	return w.write()
}

// IndexReader ...
type IndexReader struct {
	data  []byte
	dec   *encoding.Decoder
	base  postings.ID
	limit postings.ID
}

// NewIndexReader ...
func NewIndexReader(data []byte) (*IndexReader, error) {
	if len(data) < indexHeaderSize {
		return nil, io.ErrShortBuffer
	}

	payloadLen := len(data) - indexHeaderSize
	if payloadLen%8 != 0 {
		return nil, fmt.Errorf("stored fields index payload should be a multiple of 8, found %v", payloadLen%8)
	}
	count := payloadLen / 8

	r := &IndexReader{
		data: data,
	}

	r.dec = encoding.NewDecoder(data[:8])
	base, err := r.dec.Uint64()
	if err != nil {
		return nil, fmt.Errorf("could not read base postings ID: %v", err)
	}
	r.base = postings.ID(base)
	r.limit = r.base + postings.ID(count)
	return r, nil
}

func (r *IndexReader) Read(id postings.ID) (uint64, error) {
	if id < r.base || id > r.limit {
		return 0, fmt.Errorf("invalid postings ID %v, must be between [%v, %v]", id, r.base, r.limit)
	}

	idx := (int(id-r.base) * 8) + indexHeaderSize
	r.dec.Reset(r.data[idx:])
	offset, err := r.dec.Uint64()
	if err != nil {
		return 0, err
	}

	return offset, nil
}
