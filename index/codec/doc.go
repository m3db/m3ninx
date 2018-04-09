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

package codec

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"

	"github.com/m3db/m3ninx/doc"
)

// Documents file format metadata.
const (
	docsFileFormatVersion = 1

	docsMagicNumber = 0x6D33D0C5

	docsHeaderSize = 0 +
		4 + // magic number
		4 + // version
		4 + // checksum of header fields
		0

	docsTrailerSize = 0 +
		8 + // number of documents
		4 + // checksum of trailer fields
		0
)

const initialEncoderSize = 1024

var (
	errEmptyDocument      = errors.New("cannot write an empty document")
	errInvalidMagicNumber = errors.New("encountered invalid magic number")
	errChecksumMismatch   = errors.New("encountered invalid checksum")
)

type docWriter struct {
	version uint32

	writer io.Writer
	enc    *encoder

	docs     uint64
	checksum *checksum
}

// TODO: remove the nolint
// nolint: deadcode
func newDocWriter(w io.Writer) *docWriter {
	return &docWriter{
		version:  docsFileFormatVersion,
		writer:   w,
		enc:      newEncoder(initialEncoderSize),
		checksum: new(checksum),
	}
}

func (w *docWriter) Init() error {
	w.enc.putUint32(docsMagicNumber)
	w.enc.putUint32(w.version)
	if err := w.write(); err != nil {
		return err
	}

	w.enc.putUint32(w.checksum.get())
	if err := w.write(); err != nil {
		return err
	}
	w.checksum.reset()

	return nil
}

func (w *docWriter) Write(d doc.Document) error {
	if len(d.Fields) == 0 {
		return errEmptyDocument
	}

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
	w.enc.putUint32(w.checksum.get())
	if err := w.write(); err != nil {
		return err
	}
	w.checksum.reset()

	w.enc.putUint64(w.docs)
	if err := w.write(); err != nil {
		return err
	}

	w.enc.putUint32(w.checksum.get())
	if err := w.write(); err != nil {
		return err
	}
	w.checksum.reset()

	return nil
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
	w.checksum.update(b)
	w.enc.reset()
	return nil
}

type docReader struct {
	version uint32

	data []byte
	dec  *decoder

	curr     uint64
	total    uint64
	checksum *checksum
}

// TODO: remove the nolint
// nolint: deadcode
func newDocReader(data []byte) *docReader {
	return &docReader{
		data:     data,
		dec:      new(decoder),
		checksum: new(checksum),
	}
}

func (r *docReader) Init() error {
	if len(r.data) < docsHeaderSize+docsTrailerSize {
		return io.ErrShortBuffer
	}

	if err := r.decodeHeader(); err != nil {
		return err
	}
	if err := r.decodeTrailer(); err != nil {
		return err
	}
	if err := r.verifyPayload(); err != nil {
		return err
	}

	return nil
}

func (r *docReader) decodeHeader() error {
	data := r.data[:docsHeaderSize]
	r.dec.reset(data)

	// Verify magic number.
	n, err := r.dec.uint32()
	if err != nil {
		return err
	}
	if n != docsMagicNumber {
		return errInvalidMagicNumber
	}

	// Verify file format version.
	v, err := r.dec.uint32()
	if err != nil {
		return err
	}
	if v != docsFileFormatVersion {
		return fmt.Errorf("version of file format: %v does not match expected version: %v", v, docsFileFormatVersion)
	}
	r.version = v

	r.checksum.update(data[:8])
	return r.verifyChecksum()
}

func (r *docReader) decodeTrailer() error {
	data := r.data[len(r.data)-docsTrailerSize:]
	r.dec.reset(data)

	// Get number of documents.
	n, err := r.dec.uint64()
	if err != nil {
		return err
	}
	r.total = n

	r.checksum.update(data[:8])
	return r.verifyChecksum()
}

func (r *docReader) verifyPayload() error {
	start := docsHeaderSize
	end := len(r.data) - docsTrailerSize - 4
	payload := r.data[start:end]
	r.checksum.update(payload)

	checksum := r.data[end : end+4]
	r.dec.reset(checksum)
	if err := r.verifyChecksum(); err != nil {
		return err
	}

	// Update decoder for use in subsequent calls to Read.
	r.dec.reset(payload)
	return nil
}

func (r *docReader) verifyChecksum() error {
	c, err := r.dec.uint32()
	if err != nil {
		return err
	}
	if c != r.checksum.get() {
		return errChecksumMismatch
	}
	r.checksum.reset()
	return nil
}

func (r *docReader) Read() (doc.Document, error) {
	if r.curr == r.total {
		return doc.Document{}, io.EOF
	}

	x, err := r.dec.uvarint()
	if err != nil {
		return doc.Document{}, err
	}
	n := int(x)

	d := doc.Document{
		Fields: make([]doc.Field, n),
	}

	for i := 0; i < n; i++ {
		name, err := r.dec.bytes()
		if err != nil {
			return doc.Document{}, err
		}
		val, err := r.dec.bytes()
		if err != nil {
			return doc.Document{}, err
		}
		d.Fields[i] = doc.Field{
			Name:  name,
			Value: val,
		}
	}

	r.curr++
	return d, nil
}

func (r *docReader) Close() error {
	return nil
}

type checksum uint32

func (c *checksum) get() uint32 { return uint32(*c) }
func (c *checksum) reset()      { *c = 0 }

func (c *checksum) update(b []byte) {
	*c = checksum(crc32.Update(c.get(), crc32.IEEETable, b))
}
