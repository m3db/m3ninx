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

package mem

import (
	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"

	"github.com/m3db/m3x/instrument"
)

// Segment represents a memory backed index segment.
type Segment interface {
	segment.Segment
	segment.Readable
	segment.Writable
}

// Options is a collection of knobs for an in-memory segment.
type Options interface {
	// SetInstrumentOptions sets the instrument options.
	SetInstrumentOptions(value instrument.Options) Options

	// InstrumentOptions returns the instrument options.
	InstrumentOptions() instrument.Options

	// SetPostingsListPool sets the PostingsListPool.
	SetPostingsListPool(value segment.PostingsListPool) Options

	// PostingsListPool returns the PostingsListPool.
	PostingsListPool() segment.PostingsListPool

	// SetInitialCapacity sets the initial capacity.
	SetInitialCapacity(value int) Options

	// InitialCapacity returns the initial capacity.
	InitialCapacity() int
}

// termDictionary represents a mapping from doc.Field -> doc
type termsDictionary interface {
	// Insert inserts a mapping from field `f` to document ID `docID`.
	Insert(f doc.Field, docID segment.DocID) error

	// Fetch returns all the docIDs matching the given filter.
	// NB(prateek): the returned PostingsList is safe to modify.
	Fetch(filterName []byte, filterValue []byte, isRegex bool) (segment.PostingsList, error)
}

// postingsManagerOffset is an offset used to reference a postingsList in the postingsManager.
type postingsManagerOffset int

// postingsManager represents an efficient mapping from postingsManagerOffset -> PostingsList
type postingsManager interface {
	// Insert inserts document `i` at a new postingsManagerOffset, and returns the
	// offset used.
	Insert(i segment.DocID) postingsManagerOffset

	// Update updates postingsManagerOffset `p` to include a reference to document `i`.
	Update(p postingsManagerOffset, i segment.DocID) error

	// Fetch retrieves all documents which are associated with postingsManagerOffset `p`.
	Fetch(p postingsManagerOffset) (segment.ImmutablePostingsList, error)
}

// predicate returns a bool indicating if the document matched the
// provided criterion, or not.
type predicate func(d doc.Document) bool

// queryable is the base contract required for any mem segement implementation to be used by a `searcher`.
type queryable interface {
	// Filter retrieves the PostingsList for the filter, and any additional
	// filtering criterion that must be applied in post processing.
	Filter(
		filterName []byte,
		filterValue []byte,
		isRegex bool,
	) (
		candidateDocIDs segment.PostingsList,
		pendingFilterFn predicate,
		err error,
	)

	// Fetch retrieves the list of documents in the given postings list.
	Fetch(p segment.PostingsList) ([]doc.Document, error)
}

// searcher performs a search on known queryable(s).
type searcher interface {
	// Query retrieves the list of documents matching the given criterion.
	Query(q segment.Query) ([]doc.Document, error)
}
