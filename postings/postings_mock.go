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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/m3db/m3ninx/postings/types.go

package postings

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of List interface
type MockList struct {
	ctrl     *gomock.Controller
	recorder *_MockListRecorder
}

// Recorder for MockList (not exported)
type _MockListRecorder struct {
	mock *MockList
}

func NewMockList(ctrl *gomock.Controller) *MockList {
	mock := &MockList{ctrl: ctrl}
	mock.recorder = &_MockListRecorder{mock}
	return mock
}

func (_m *MockList) EXPECT() *_MockListRecorder {
	return _m.recorder
}

func (_m *MockList) Contains(id ID) bool {
	ret := _m.ctrl.Call(_m, "Contains", id)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockListRecorder) Contains(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Contains", arg0)
}

func (_m *MockList) IsEmpty() bool {
	ret := _m.ctrl.Call(_m, "IsEmpty")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockListRecorder) IsEmpty() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsEmpty")
}

func (_m *MockList) Min() (ID, error) {
	ret := _m.ctrl.Call(_m, "Min")
	ret0, _ := ret[0].(ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockListRecorder) Min() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Min")
}

func (_m *MockList) Max() (ID, error) {
	ret := _m.ctrl.Call(_m, "Max")
	ret0, _ := ret[0].(ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockListRecorder) Max() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Max")
}

func (_m *MockList) Size() uint64 {
	ret := _m.ctrl.Call(_m, "Size")
	ret0, _ := ret[0].(uint64)
	return ret0
}

func (_mr *_MockListRecorder) Size() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Size")
}

func (_m *MockList) Iterator() Iterator {
	ret := _m.ctrl.Call(_m, "Iterator")
	ret0, _ := ret[0].(Iterator)
	return ret0
}

func (_mr *_MockListRecorder) Iterator() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Iterator")
}

func (_m *MockList) Clone() MutableList {
	ret := _m.ctrl.Call(_m, "Clone")
	ret0, _ := ret[0].(MutableList)
	return ret0
}

func (_mr *_MockListRecorder) Clone() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Clone")
}

// Mock of MutableList interface
type MockMutableList struct {
	ctrl     *gomock.Controller
	recorder *_MockMutableListRecorder
}

// Recorder for MockMutableList (not exported)
type _MockMutableListRecorder struct {
	mock *MockMutableList
}

func NewMockMutableList(ctrl *gomock.Controller) *MockMutableList {
	mock := &MockMutableList{ctrl: ctrl}
	mock.recorder = &_MockMutableListRecorder{mock}
	return mock
}

func (_m *MockMutableList) EXPECT() *_MockMutableListRecorder {
	return _m.recorder
}

func (_m *MockMutableList) Contains(id ID) bool {
	ret := _m.ctrl.Call(_m, "Contains", id)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockMutableListRecorder) Contains(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Contains", arg0)
}

func (_m *MockMutableList) IsEmpty() bool {
	ret := _m.ctrl.Call(_m, "IsEmpty")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockMutableListRecorder) IsEmpty() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsEmpty")
}

func (_m *MockMutableList) Min() (ID, error) {
	ret := _m.ctrl.Call(_m, "Min")
	ret0, _ := ret[0].(ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMutableListRecorder) Min() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Min")
}

func (_m *MockMutableList) Max() (ID, error) {
	ret := _m.ctrl.Call(_m, "Max")
	ret0, _ := ret[0].(ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMutableListRecorder) Max() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Max")
}

func (_m *MockMutableList) Size() uint64 {
	ret := _m.ctrl.Call(_m, "Size")
	ret0, _ := ret[0].(uint64)
	return ret0
}

func (_mr *_MockMutableListRecorder) Size() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Size")
}

func (_m *MockMutableList) Iterator() Iterator {
	ret := _m.ctrl.Call(_m, "Iterator")
	ret0, _ := ret[0].(Iterator)
	return ret0
}

func (_mr *_MockMutableListRecorder) Iterator() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Iterator")
}

func (_m *MockMutableList) Clone() MutableList {
	ret := _m.ctrl.Call(_m, "Clone")
	ret0, _ := ret[0].(MutableList)
	return ret0
}

func (_mr *_MockMutableListRecorder) Clone() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Clone")
}

func (_m *MockMutableList) Insert(i ID) error {
	ret := _m.ctrl.Call(_m, "Insert", i)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockMutableListRecorder) Insert(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Insert", arg0)
}

func (_m *MockMutableList) Intersect(other List) error {
	ret := _m.ctrl.Call(_m, "Intersect", other)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockMutableListRecorder) Intersect(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Intersect", arg0)
}

func (_m *MockMutableList) Difference(other List) error {
	ret := _m.ctrl.Call(_m, "Difference", other)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockMutableListRecorder) Difference(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Difference", arg0)
}

func (_m *MockMutableList) Union(other List) error {
	ret := _m.ctrl.Call(_m, "Union", other)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockMutableListRecorder) Union(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Union", arg0)
}

func (_m *MockMutableList) RemoveRange(min ID, max ID) error {
	ret := _m.ctrl.Call(_m, "RemoveRange", min, max)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockMutableListRecorder) RemoveRange(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RemoveRange", arg0, arg1)
}

func (_m *MockMutableList) Reset() {
	_m.ctrl.Call(_m, "Reset")
}

func (_mr *_MockMutableListRecorder) Reset() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Reset")
}

// Mock of Iterator interface
type MockIterator struct {
	ctrl     *gomock.Controller
	recorder *_MockIteratorRecorder
}

// Recorder for MockIterator (not exported)
type _MockIteratorRecorder struct {
	mock *MockIterator
}

func NewMockIterator(ctrl *gomock.Controller) *MockIterator {
	mock := &MockIterator{ctrl: ctrl}
	mock.recorder = &_MockIteratorRecorder{mock}
	return mock
}

func (_m *MockIterator) EXPECT() *_MockIteratorRecorder {
	return _m.recorder
}

func (_m *MockIterator) Next() bool {
	ret := _m.ctrl.Call(_m, "Next")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIteratorRecorder) Next() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Next")
}

func (_m *MockIterator) Current() ID {
	ret := _m.ctrl.Call(_m, "Current")
	ret0, _ := ret[0].(ID)
	return ret0
}

func (_mr *_MockIteratorRecorder) Current() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Current")
}

func (_m *MockIterator) Err() error {
	ret := _m.ctrl.Call(_m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockIteratorRecorder) Err() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Err")
}

func (_m *MockIterator) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockIteratorRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

// Mock of Pool interface
type MockPool struct {
	ctrl     *gomock.Controller
	recorder *_MockPoolRecorder
}

// Recorder for MockPool (not exported)
type _MockPoolRecorder struct {
	mock *MockPool
}

func NewMockPool(ctrl *gomock.Controller) *MockPool {
	mock := &MockPool{ctrl: ctrl}
	mock.recorder = &_MockPoolRecorder{mock}
	return mock
}

func (_m *MockPool) EXPECT() *_MockPoolRecorder {
	return _m.recorder
}

func (_m *MockPool) Get() MutableList {
	ret := _m.ctrl.Call(_m, "Get")
	ret0, _ := ret[0].(MutableList)
	return ret0
}

func (_mr *_MockPoolRecorder) Get() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get")
}

func (_m *MockPool) Put(pl MutableList) {
	_m.ctrl.Call(_m, "Put", pl)
}

func (_mr *_MockPoolRecorder) Put(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Put", arg0)
}
