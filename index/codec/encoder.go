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

import "encoding/binary"

var byteOrder = binary.BigEndian

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

func (e *encoder) putVarint(x uint64) {
	n := binary.PutUvarint(e.tmp[:], x)
	e.buf = append(e.buf, e.tmp[:n]...)
}

func (e *encoder) putBytes(b []byte) {
	e.putUvarint(uint64(len(b)))
	e.buf = append(e.buf, b...)
}
