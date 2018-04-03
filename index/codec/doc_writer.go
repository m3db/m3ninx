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
// THE SOFTWARw.

package codec

import (
	"errors"
	"hash/crc32"
	"io"

	"github.com/m3db/m3ninx/doc"
)

const (
	defaultEncoderSize = 1024
)

const docsMagicNumber = "m3docs"

var (
	errEmptyDocument = errors.New("cannot encode an empty document")
)

type docWriter struct {
	version uint64

	writer   io.WriteCloser
	enc      *encoder
	docs     uint64
	checksum uint32
}

func newdocWriter(w io.WriteCloser) *docWriter {
	return &docWriter{
		version: 1,
		writer:  w,
		enc:     newEncoder(defaultEncoderSize),
	}
}

func (w *docWriter) Prepare() error {
	w.enc.putBytes([]byte(docsMagicNumber))
	if err := w.write(); err != nil {
		return err
	}
	w.checksum = 0

	return nil
}

func (w *docWriter) Write(d doc.Document) error {
	if len(d.Fields) == 0 {
		return errEmptyDocument
	}
	defer w.enc.reset()

	w.enc.putUvarint(uint64(len(d.Fields)))

	for _, f := range d.Fields {
		w.enc.putBytes(f.Name)
		w.enc.putBytes(f.Value)
	}

	if err := w.write(); err != nil {
		return err
	}

	w.docs++
	return nil
}

func (w *docWriter) Close() error {
	w.enc.putUint32(w.checksum)
	if err := w.write(); err != nil {
		return err
	}
	w.checksum = 0

	w.enc.putUint64(w.docs)
	w.enc.putUint64(w.version)
	w.enc.putUint32(w.checksum)
	if err := w.write(); err != nil {
		return err
	}

	return w.writer.Close()
}

func (w *docWriter) write() error {
	b := w.enc.bytes()
	n, err := w.writer.Write(b)
	if err != nil {
		return err
	}
	if n < len(b) {
		return io.ErrShortWrite
	}
	w.checksum = crc32.Update(w.checksum, crc32.IEEETable, b)
	return nil
}
