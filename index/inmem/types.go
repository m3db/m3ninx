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

package inmem

import (
	"regexp"

	"github.com/m3db/m3ninx/doc"
	mi "github.com/m3db/m3ninx/index"
	"github.com/m3db/m3x/instrument"

	"github.com/RoaringBitmap/roaring"
)

// Index represents a memory backed index.
type Index interface {
	mi.Readable
	mi.Writable

	// Open prepares the index to accept reads/writes.
	Open() error

	// Close closes the index.
	Close() error

	// Freeze makes the index immutable. Any writes after this return error.
	Freeze() error
}

type termsDictionary interface {
	FieldNames() [][]byte

	FetchTerms(FieldName []byte) []doc.Term

	FetchDocs(FieldName []byte, Term doc.Term, negate bool) *roaring.Bitmap

	FetchDocsFuzzy(FieldName []byte, TermRe *regexp.Regexp, negate bool) *roaring.Bitmap
}

type postingID int64

// postingList represents an efficient mapping from postingID -> bitmap of doc ids
type postingList interface {
	// Insert inserts a reference to document `i` for posting id `p`.
	Insert(p postingID, i mi.DocID)

	// Remove removes the reference (if any) from document `i` from posting id `p`.
	Remove(p postingID, i mi.DocID)

	// Fetch retrieves all documents which are associated with posting id `p`.
	Fetch(p postingID) (*roaring.Bitmap, error)
}

// Options is a collection of knobs for an in-memory index.
type Options interface {
	// SetInstrumentOptions sets the instrument options.
	SetInstrumentOptions(value instrument.Options) Options

	// InstrumentOptions returns the instrument options.
	InstrumentOptions() instrument.Options

	// SetInsertQueueSize sets the insert queue size.
	SetInsertQueueSize(value int64) Options

	// InsertQueueSize returns the insert queue size.
	InsertQueueSize() int64

	// SetNumInsertWorkers sets the number of insert workers.
	SetNumInsertWorkers(value int) Options

	// NumInsertWorkers returns the number of insert workers.
	NumInsertWorkers() int
}
