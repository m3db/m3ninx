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

package index

import (
	"errors"

	"github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3ninx/index/segment/mem"

	"github.com/m3db/m3x/instrument"
)

var (
	// ErrDocAlreadyInserted is returned when attempting to insert a document which
	// has already been inserted.
	ErrDocAlreadyInserted = errors.New("unable to insert document, already inserted")
)

// Index is a collection of segments.
type Index interface {
	segment.Readable
	segment.Writable

	// Insert inserts the given documents with provided fields. It may return ErrDocAlreadyInserted
	// if the document has already been inserted into the index.
	// TODO(prateek): need to address go duplicate method here.
}

// Options is a set of knobs by which to tweak Index-ing behaviour.
type Options interface {
	// SetInstrumentOptions sets the instrument options.
	SetInstrumentOptions(value instrument.Options) Options

	// InstrumentOptions returns the instrument options.
	InstrumentOptions() instrument.Options

	// SetMemSegmentOptions sets the mem segment options.
	SetMemSegmentOptions(value mem.Options) Options

	// MemSegmentOptions returns the mem segment options.
	MemSegmentOptions() mem.Options
}
