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
// all copies or substantial portions of the Softwarw.
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

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment/fs/encoding"
)

const (
	initialDataEncoderLen     = 1024
	dataEncoderWriteThreshold = 512
)

// DataWriter writes the data file for stored fields.
type DataWriter struct {
	writer io.Writer
	enc    *encoding.Encoder

	idx uint64
}

// NewDataWriter returns a new DataWriter.
func NewDataWriter(w io.Writer) *DataWriter {
	return &DataWriter{
		writer: w,
		enc:    encoding.NewEncoder(initialDataEncoderLen),
	}
}

func (w *DataWriter) Write(d doc.Document) (int, error) {
	n := w.enc.PutBytes(d.ID)
	n += w.enc.PutUvarint(uint64(len(d.Fields)))
	for _, f := range d.Fields {
		n += w.enc.PutBytes(f.Name)
		n += w.enc.PutBytes(f.Value)
	}

	if w.enc.Len() > dataEncoderWriteThreshold {
		if err := w.write(); err != nil {
			return 0, err
		}
	}

	return n, nil
}

func (w *DataWriter) write() error {
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

// Close closes the DataWriter and ensures any buffered data is written out.
func (w *DataWriter) Close() error {
	if w.enc.Len() == 0 {
		return nil
	}
	return w.write()
}

// DataReader is a reader for the data file for stored fields.
type DataReader struct {
	data []byte
	dec  *encoding.Decoder
}

// NewDataReader returns a new DataReader.
func NewDataReader(data []byte) *DataReader {
	return &DataReader{
		data: data,
		dec:  encoding.NewDecoder(nil),
	}
}

func (r *DataReader) Read(offset uint64) (doc.Document, error) {
	if offset > uint64(len(r.data)) {
		return doc.Document{}, fmt.Errorf("invalid offset: %v is past the end of the data file", offset)
	}
	r.dec.Reset(r.data[int(offset):])

	id, err := r.dec.Bytes()
	if err != nil {
		return doc.Document{}, err
	}

	x, err := r.dec.Uvarint()
	if err != nil {
		return doc.Document{}, err
	}
	n := int(x)

	d := doc.Document{
		ID:     id,
		Fields: make([]doc.Field, n),
	}

	for i := 0; i < n; i++ {
		name, err := r.dec.Bytes()
		if err != nil {
			return doc.Document{}, err
		}
		val, err := r.dec.Bytes()
		if err != nil {
			return doc.Document{}, err
		}
		d.Fields[i] = doc.Field{
			Name:  name,
			Value: val,
		}
	}

	return d, nil
}
