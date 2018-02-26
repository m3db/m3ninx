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

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/m3db/m3ninx/index/types.go

package index

import (
	gomock "github.com/golang/mock/gomock"
	doc "github.com/m3db/m3ninx/doc"
	postings "github.com/m3db/m3ninx/postings"
	reflect "reflect"
	regexp "regexp"
)

// MockIndex is a mock of Index interface
type MockIndex struct {
	ctrl     *gomock.Controller
	recorder *MockIndexMockRecorder
}

// MockIndexMockRecorder is the mock recorder for MockIndex
type MockIndexMockRecorder struct {
	mock *MockIndex
}

// NewMockIndex creates a new mock instance
func NewMockIndex(ctrl *gomock.Controller) *MockIndex {
	mock := &MockIndex{ctrl: ctrl}
	mock.recorder = &MockIndexMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockIndex) EXPECT() *MockIndexMockRecorder {
	return _m.recorder
}

// Insert mocks base method
func (_m *MockIndex) Insert(d doc.Document) error {
	ret := _m.ctrl.Call(_m, "Insert", d)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert
func (_mr *MockIndexMockRecorder) Insert(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Insert", reflect.TypeOf((*MockIndex)(nil).Insert), arg0)
}

// Snapshot mocks base method
func (_m *MockIndex) Snapshot() (Snapshot, error) {
	ret := _m.ctrl.Call(_m, "Snapshot")
	ret0, _ := ret[0].(Snapshot)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Snapshot indicates an expected call of Snapshot
func (_mr *MockIndexMockRecorder) Snapshot() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Snapshot", reflect.TypeOf((*MockIndex)(nil).Snapshot))
}

// Seal mocks base method
func (_m *MockIndex) Seal() error {
	ret := _m.ctrl.Call(_m, "Seal")
	ret0, _ := ret[0].(error)
	return ret0
}

// Seal indicates an expected call of Seal
func (_mr *MockIndexMockRecorder) Seal() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Seal", reflect.TypeOf((*MockIndex)(nil).Seal))
}

// Close mocks base method
func (_m *MockIndex) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (_mr *MockIndexMockRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Close", reflect.TypeOf((*MockIndex)(nil).Close))
}

// MockSnapshot is a mock of Snapshot interface
type MockSnapshot struct {
	ctrl     *gomock.Controller
	recorder *MockSnapshotMockRecorder
}

// MockSnapshotMockRecorder is the mock recorder for MockSnapshot
type MockSnapshotMockRecorder struct {
	mock *MockSnapshot
}

// NewMockSnapshot creates a new mock instance
func NewMockSnapshot(ctrl *gomock.Controller) *MockSnapshot {
	mock := &MockSnapshot{ctrl: ctrl}
	mock.recorder = &MockSnapshotMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockSnapshot) EXPECT() *MockSnapshotMockRecorder {
	return _m.recorder
}

// Readers mocks base method
func (_m *MockSnapshot) Readers() []Reader {
	ret := _m.ctrl.Call(_m, "Readers")
	ret0, _ := ret[0].([]Reader)
	return ret0
}

// Readers indicates an expected call of Readers
func (_mr *MockSnapshotMockRecorder) Readers() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Readers", reflect.TypeOf((*MockSnapshot)(nil).Readers))
}

// Close mocks base method
func (_m *MockSnapshot) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (_mr *MockSnapshotMockRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Close", reflect.TypeOf((*MockSnapshot)(nil).Close))
}

// MockReader is a mock of Reader interface
type MockReader struct {
	ctrl     *gomock.Controller
	recorder *MockReaderMockRecorder
}

// MockReaderMockRecorder is the mock recorder for MockReader
type MockReaderMockRecorder struct {
	mock *MockReader
}

// NewMockReader creates a new mock instance
func NewMockReader(ctrl *gomock.Controller) *MockReader {
	mock := &MockReader{ctrl: ctrl}
	mock.recorder = &MockReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockReader) EXPECT() *MockReaderMockRecorder {
	return _m.recorder
}

// MatchExact mocks base method
func (_m *MockReader) MatchExact(name []byte, value []byte) (postings.List, error) {
	ret := _m.ctrl.Call(_m, "MatchExact", name, value)
	ret0, _ := ret[0].(postings.List)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MatchExact indicates an expected call of MatchExact
func (_mr *MockReaderMockRecorder) MatchExact(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "MatchExact", reflect.TypeOf((*MockReader)(nil).MatchExact), arg0, arg1)
}

// MatchRegex mocks base method
func (_m *MockReader) MatchRegex(name []byte, pattern []byte, re *regexp.Regexp) (postings.List, error) {
	ret := _m.ctrl.Call(_m, "MatchRegex", name, pattern, re)
	ret0, _ := ret[0].(postings.List)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MatchRegex indicates an expected call of MatchRegex
func (_mr *MockReaderMockRecorder) MatchRegex(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "MatchRegex", reflect.TypeOf((*MockReader)(nil).MatchRegex), arg0, arg1, arg2)
}

// Docs mocks base method
func (_m *MockReader) Docs(pl postings.List, names [][]byte) (doc.Iterator, error) {
	ret := _m.ctrl.Call(_m, "Docs", pl, names)
	ret0, _ := ret[0].(doc.Iterator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Docs indicates an expected call of Docs
func (_mr *MockReaderMockRecorder) Docs(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Docs", reflect.TypeOf((*MockReader)(nil).Docs), arg0, arg1)
}

// Close mocks base method
func (_m *MockReader) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (_mr *MockReaderMockRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Close", reflect.TypeOf((*MockReader)(nil).Close))
}