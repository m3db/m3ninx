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
	"bytes"
	"crypto/md5"
	"sort"

	"github.com/m3db/m3ninx/doc"
	"github.com/m3db/m3ninx/index/segment"

	"github.com/couchbaselabs/vellum"
	vregex "github.com/couchbaselabs/vellum/regexp"
)

const fieldJoinChar = '='

type fstIndex struct {
	opts       Options
	concatChar byte // TODO(prateek): migrate to Options

	id  segment.ID
	fst *vellum.FST

	documents       map[segment.DocID]document // TODO(prateek): measure perf impact of slice v map here
	postingsManager postingsManager
}

// NewFST returns a new FST backed segment.
func NewFST(id segment.ID, segments []segment.Segment, opts Options) (segment.Readable, error) {
	hashFn := func(i doc.ID) doc.Hash {
		// TODO: evaluate impact of other hash functions on perf/correctness
		return md5.Sum([]byte(i))
	}
	f := &fstIndex{
		id:              id,
		opts:            opts,
		concatChar:      fieldJoinChar,
		postingsManager: newPostingsManager(opts),
		documents:       make(map[segment.DocID]document, opts.InitialCapacity()),
	}
	err := f.initialize(segments, hashFn)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// TODO(prateek): deconstruct type into Builder + Readable
// this would allow more testable code, and it's probably possible to
// have memIndex's internals optimised to enable the Builder primitives more
// efficiently, e.g. we could get all field names, sort, and then retrieve all
// the values a field name a time, sort, and then do the insert. This would
// greatly reduce the overall memory footprint required of the fstIndex initialise
// method.

func (f *fstIndex) initialize(
	segments []segment.Segment,
	hashFn doc.HashFn,
) error {
	// sort segments by ID
	sortedSegments := segmentsOrderedByID(segments)
	sort.Sort(sortedSegments)

	// TODO(prateek): address inefficiencies in fst construction:
	//   - making the entire corpus' reverse index fit into memory is terrible.
	//   - we have a posting list constructed after the first loop, but re-construct it in the second.
	var (
		idGenerator       segment.DocID
		fieldNameToValues = make(map[string]map[string]segment.PostingsList, f.opts.InitialCapacity())
		hashedDocuments   = make(map[doc.Hash]document, f.opts.InitialCapacity())
	)
	for _, seg := range sortedSegments {
		iter := seg.Iter()
		for iter.Next() {
			newDoc, tombstoned, _ := iter.Current()
			hashID := hashFn(newDoc.ID)
			oldDoc, oldDocExists := hashedDocuments[hashID]

			// handle new doc insertion
			if !oldDocExists {
				id := idGenerator
				idGenerator++
				d := document{
					Document:   newDoc,
					docID:      id,
					tombstoned: tombstoned,
				}
				f.documents[id] = d
				hashedDocuments[hashID] = d
				for _, field := range newDoc.Fields {
					fieldName := string(field.Name)
					fieldValue := string(field.Value)
					vals, ok := fieldNameToValues[fieldName]
					if !ok {
						vals = make(map[string]segment.PostingsList)
						fieldNameToValues[fieldName] = vals
					}
					docIDs, ok := vals[fieldValue]
					if !ok {
						docIDs = f.opts.PostingsListPool().Get()
						vals[fieldValue] = docIDs
					}
					docIDs.Insert(id)
				}

				// done with insertion, move to next document
				continue
			}

			// i.e. we have a document for the current hash, handle merge/update/delete

			// check if the newDoc has a tombstone
			if tombstoned {
				// if so, delete it
				delete(f.documents, oldDoc.docID)
				delete(hashedDocuments, hashID)
				// delete all the corresponding entries in the map
				for _, field := range oldDoc.Fields {
					fieldName := string(field.Name)
					fieldValue := string(field.Value)
					vals, ok := fieldNameToValues[fieldName]
					if !ok {
						continue
					}
					ids, ok := vals[fieldValue]
					if !ok {
						continue
					}
					ids.Remove(oldDoc.docID)
				}
				continue
			}

			// if it's not tombstoned, we don't need to do anything because we don't
			// have stored fields, and indexed fields do not change with updates.
		}
	}

	// now to create the fst
	var fstBuffer bytes.Buffer
	// TODO(prateek): builderopts for vellum
	builder, err := vellum.New(&fstBuffer, nil)
	if err != nil {
		return err
	}

	// need to sort the keys and values in lexicographic order
	names := make([]string, 0, len(fieldNameToValues))
	for name := range fieldNameToValues {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		values := fieldNameToValues[name]
		sortedValues := make([]string, 0, len(values))
		for value := range values {
			sortedValues = append(sortedValues, value)
		}
		sort.Strings(sortedValues)
		for _, value := range sortedValues {
			docIds := values[value]
			if docIds.IsEmpty() {
				continue
			}

			offset, err := f.postingsManager.InsertSet(docIds)
			if err != nil {
				return err
			}

			key := f.computeIndexID([]byte(name), []byte(value))
			err = builder.Insert(key, uint64(offset))
			if err != nil {
				return err
			}
		}
	}

	// finish construction
	if err := builder.Close(); err != nil {
		return err
	}

	fst, err := vellum.Load(fstBuffer.Bytes())
	if err != nil {
		return err
	}

	f.fst = fst
	return nil
}

func (f *fstIndex) Query(query segment.Query) ([]doc.Document, error) {
	if err := validateQuery(query); err != nil {
		return nil, err
	}

	// order filters to ensure the first filter has no-negation
	filters := orderFiltersByNonNegated(query.Filters)
	sort.Sort(filters)

	var (
		candidateDocIds segment.PostingsList
	)

	// TODO(prateek): Query search optimizations
	// - need to add a filter cache per segment
	// - need to share the code for the searchers
	// - pool fst iterators/regex compilations
	for _, filter := range filters {
		filter := filter
		fetchedIds, err := f.filterAndFetchIDs(filter.FieldName, filter.FieldValueFilter, filter.Regexp)
		if err != nil {
			return nil, err
		}

		// i.e. we don't have any documents for the given filter, can early terminate entire fn
		if fetchedIds == nil {
			return nil, nil
		}

		if candidateDocIds == nil {
			candidateDocIds = fetchedIds
			continue
		}

		// update candidate set
		if filter.Negate {
			candidateDocIds.Difference(fetchedIds)
		} else {
			candidateDocIds.Intersect(fetchedIds)
		}

		// early terminate if we don't have any docs in candidate set
		if candidateDocIds.IsEmpty() {
			return nil, nil
		}
	}

	docs := make([]doc.Document, 0, candidateDocIds.Size())
	iter := candidateDocIds.Iter()

	// retrieve all the filtered document ids
	for iter.Next() {
		id := iter.Current()
		if d, ok := f.fetchDocument(id); ok {
			docs = append(docs, d.Document)
		}
	}

	// TODO(prateek): emit same statistics as trigram query

	return docs, nil
}

func (f *fstIndex) filterAndFetchIDs(
	filterName, filterValue []byte,
	isRegex bool,
) (segment.PostingsList, error) {
	if isRegex {
		return f.filterAndFetchRegex(filterName, filterValue)
	}
	return f.filterAndFetchExact(filterName, filterValue)
}

func (f *fstIndex) filterAndFetchExact(filterName, filterValue []byte) (segment.PostingsList, error) {
	var (
		fetchedIds = f.opts.PostingsListPool().Get()
		minKey     = f.computeIndexID(filterName, filterValue)
		maxKey     = append(minKey, byte(0))
	)
	// TODO(prateek): need to pool these iterators
	iter, err := f.fst.Iterator(minKey, maxKey)
	if err != nil {
		return nil, err
	}
	for ; err == nil; err = iter.Next() {
		_, offset := iter.Current()
		// TODO(prateek): test key to ensure it matches expected value | assert otherwise
		// key, offset := iter.Current()
		ids, err := f.postingsManager.Fetch(postingsManagerOffset(offset))
		if err != nil {
			return nil, err
		}
		fetchedIds.Union(ids)
	}
	return fetchedIds, nil
}

func (f *fstIndex) filterAndFetchRegex(filterName, filterValue []byte) (segment.PostingsList, error) {
	var (
		regexExpr  = f.computeIndexID(filterName, filterValue)
		minKey     = f.computeIndexID(filterName, nil)
		maxKey     = f.computeNexIndexIDBoundary(filterName)
		fetchedIds = f.opts.PostingsListPool().Get()
	)
	// TODO(prateek): use NewWithLimit and provide options hook for the same.
	re, err := vregex.New(string(regexExpr))
	if err != nil {
		return nil, err
	}
	iter, err := f.fst.Search(re, minKey, maxKey)
	if err != nil {
		return nil, err
	}
	for ; err == nil; err = iter.Next() {
		_, offset := iter.Current()
		ids, err := f.postingsManager.Fetch(postingsManagerOffset(offset))
		if err != nil {
			return nil, err
		}
		fetchedIds.Union(ids)
	}
	return fetchedIds, nil
}

func (f *fstIndex) fetchDocument(id segment.DocID) (document, bool) {
	d, ok := f.documents[id]
	return d, ok
}

func (f *fstIndex) computeIndexID(fieldName, fieldValue []byte) []byte {
	return concatenateBytes(fieldName, f.concatChar, fieldValue)
}

func (f *fstIndex) computeNexIndexIDBoundary(fieldName []byte) []byte {
	return concatenateBytes(fieldName, 1+f.concatChar, nil)
}

func concatenateBytes(a []byte, b byte, c []byte) []byte {
	buf := make([]byte, len(a)+len(c)+1)
	copy(buf, a)
	copy(buf[len(a):], string(b))
	copy(buf[len(a)+1:], c)
	return buf
}
