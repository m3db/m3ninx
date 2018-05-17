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

package fs

import (
	sgmt "github.com/m3db/m3ninx/index/segment"
	"github.com/m3db/m3x/context"
	xerrors "github.com/m3db/m3x/errors"

	"github.com/couchbase/vellum"
)

type newFstTermsIterOpts struct {
	ctx         context.Context
	opts        Options
	fst         *vellum.FST
	finalizeFST bool
}

func (o *newFstTermsIterOpts) Close() error {
	o.ctx.Close()
	if o.finalizeFST {
		return o.fst.Close()
	}
	return nil
}

func newFSTTermsIter(opts newFstTermsIterOpts) *fstTermsIter {
	return &fstTermsIter{iterOpts: opts}
}

// nolint: maligned
type fstTermsIter struct {
	iterOpts newFstTermsIterOpts

	initialized bool
	iter        *vellum.FSTIterator
	err         error
	done        bool

	current      []byte
	currentValue uint64
}

var _ sgmt.OrderedBytesSliceIterator = &fstTermsIter{}

func (f *fstTermsIter) Next() bool {
	if f.done || f.err != nil {
		return false
	}

	var err error

	if !f.initialized {
		f.initialized = true
		f.iter = &vellum.FSTIterator{}
		err = f.iter.Reset(f.iterOpts.fst, minByteKey, maxByteKey, nil)
	} else {
		err = f.iter.Next()
	}

	if err != nil {
		f.done = true
		if err != vellum.ErrIteratorDone {
			f.err = err
		}
		return false
	}

	nextBytes, nextValue := f.iter.Current()
	// taking a copy of the bytes to avoid referring to mmap'd memory
	// TODO(prateek): back this by a bytes pool.
	f.current = append([]byte(nil), nextBytes...)
	f.currentValue = nextValue
	return true
}

func (f *fstTermsIter) CurrentExtended() ([]byte, uint64) {
	return f.current, f.currentValue
}

func (f *fstTermsIter) Current() []byte {
	return f.current
}

func (f *fstTermsIter) Err() error {
	return f.err
}

func (f *fstTermsIter) Close() error {
	f.current = nil

	multiErr := xerrors.MultiError{}

	if f.iter != nil {
		multiErr = multiErr.Add(f.iter.Close())
	}
	f.iter = nil

	multiErr = multiErr.Add(f.iterOpts.Close())
	return multiErr.FinalError()
}

type noOpCloser struct{}

func (n noOpCloser) Close() error { return nil }
