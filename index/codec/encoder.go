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
	"encoding/binary"
	"errors"
	"io"
)

var byteOrder = binary.BigEndian

var errUvarintOverflow = errors.New("uvarint overflows 64 bits")

type encoder struct {
	buf []byte
	tmp [binary.MaxVarintLen64]byte
}

func newEncoder(size int) *encoder {
	return &encoder{buf: make([]byte, 0, size)}
}

func (e *encoder) bytes() []byte { return e.buf }
func (e *encoder) reset()        { e.buf = e.buf[:0] }

func (e *encoder) putUint32(x uint32) {
	byteOrder.PutUint32(e.tmp[:], x)
	e.buf = append(e.buf, e.tmp[:4]...)
}

func (e *encoder) putUint64(x uint64) {
	byteOrder.PutUint64(e.tmp[:], x)
	e.buf = append(e.buf, e.tmp[:8]...)
}

func (e *encoder) putUvarint(x uint64) {
	n := binary.PutUvarint(e.tmp[:], x)
	e.buf = append(e.buf, e.tmp[:n]...)
}

func (e *encoder) putBytes(b []byte) {
	e.putUvarint(uint64(len(b)))
	e.buf = append(e.buf, b...)
}

type decoder struct {
	buf []byte
}

func (d *decoder) reset(buf []byte) { d.buf = buf }

func (d *decoder) uint32() (uint32, error) {
	if len(d.buf) < 4 {
		return 0, io.ErrShortBuffer
	}
	x := byteOrder.Uint32(d.buf)
	d.buf = d.buf[4:]
	return x, nil
}

func (d *decoder) uint64() (uint64, error) {
	if len(d.buf) < 8 {
		return 0, io.ErrShortBuffer
	}
	x := byteOrder.Uint64(d.buf)
	d.buf = d.buf[8:]
	return x, nil
}

func (d *decoder) uvarint() (uint64, error) {
	x, n := binary.Uvarint(d.buf)
	if n == 0 {
		return 0, io.ErrShortBuffer
	}
	if n < 0 {
		return 0, errUvarintOverflow
	}
	d.buf = d.buf[n:]
	return x, nil
}

func (d *decoder) bytes() ([]byte, error) {
	x, err := d.uvarint()
	if err != nil {
		return nil, err
	}
	// TODO: check for overflow here?
	n := int(x)
	if len(d.buf) < n {
		return nil, io.ErrShortBuffer
	}
	b := d.buf[:n]
	d.buf = d.buf[n:]
	return b, nil
}
