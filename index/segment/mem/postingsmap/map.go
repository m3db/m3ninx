// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

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

package postingsmap

import "github.com/m3db/m3ninx/postings"

// MapHash is the hash for a given map entry, this is public to support
// iterating over the map using a native Go for loop.
type MapHash uint64

// HashFn is the hash function to execute when hashing a key.
type HashFn func([]byte) MapHash

// EqualsFn is the equals key function to execute when detecting equality of a key.
type EqualsFn func([]byte, []byte) bool

// CopyFn is the copy key function to execute when copying the key.
type CopyFn func([]byte) []byte

// FinalizeFn is the finalize key function to execute when finished with a key.
type FinalizeFn func([]byte)

// Map uses the genny package to provide a generic hash map that can be specialized
// by running the following command from this root of the repository:
// ```
// make hashmap-gen pkg=outpkg key_type=Type value_type=Type out_dir=/tmp
// ```
// Or if you would like to use bytes or ident.ID as keys you can use the
// partially specialized maps to generate your own maps as well:
// ```
// make byteshashmap-gen pkg=outpkg value_type=Type out_dir=/tmp
// make idhashmap-gen pkg=outpkg value_type=Type out_dir=/tmp
// ```
// This will output to stdout the generated source file to use for your map.
// It uses linear probing by incrementing the number of the hash created when
// hashing the identifier if there is a collision.
// Map is a value type and not an interface to allow for less painful
// upgrades when adding/removing methods, it is not likely to need mocking so
// an interface would not be super useful either.
type Map struct {
	mapOptions

	// lookup uses hash of the identifier for the key and the MapEntry value
	// wraps the value type and the key (used to ensure lookup is correct
	// when dealing with collisions), we use uint64 for the hash partially
	// because lookups of maps with uint64 keys has a fast path for Go.
	lookup map[MapHash]MapEntry
}

// mapOptions is a set of options used when creating an identifier map, it is kept
// private so that implementers of the generated map can specify their own options
// that partially fulfill these options.
type mapOptions struct {
	// hash is the hash function to execute when hashing a key.
	hash HashFn
	// equals is the equals key function to execute when detecting equality.
	equals EqualsFn
	// copy is the copy key function to execute when copying the key.
	copy CopyFn
	// finalize is the finalize key function to execute when finished with a
	// key, this is optional to specify.
	finalize FinalizeFn
	// initialSize is the initial size for the map, use zero to use Go's std map
	// initial size and consequently is optional to specify.
	initialSize int
}

// MapEntry is an entry in the map, this is public to support iterating
// over the map using a native Go for loop.
type MapEntry struct {
	// key is used to check equality on lookups to resolve collisions
	key mapKey
	// value type stored
	value postings.MutableList
}

type mapKey struct {
	key      []byte
	finalize bool
}

// Key returns the map entry key.
func (e MapEntry) Key() []byte {
	return e.key.key
}

// Value returns the map entry value.
func (e MapEntry) Value() postings.MutableList {
	return e.value
}

// newMap is a non-exported function so that when generating the source code
// for the map you can supply a public constructor that sets the correct
// hash, equals, copy, finalize options without users of the map needing to
// implement them themselves.
// nolint: deadcode
func newMap(opts mapOptions) *Map {
	m := &Map{mapOptions: opts}
	m.Reallocate()
	return m
}

func (m *Map) newMapKey(k []byte, opts mapKeyOptions) mapKey {
	key := mapKey{key: k, finalize: opts.finalizeKey}
	if !opts.copyKey {
		return key
	}

	key.key = m.copy(k)
	return key
}

func (m *Map) removeMapKey(hash MapHash, key mapKey) {
	delete(m.lookup, hash)
	if key.finalize {
		m.finalize(key.key)
	}
}

// Get returns a value in the map for an identifier if found.
func (m *Map) Get(k []byte) (postings.MutableList, bool) {
	hash := m.hash(k)
	for entry, ok := m.lookup[hash]; ok; entry, ok = m.lookup[hash] {
		if m.equals(entry.key.key, k) {
			return entry.value, true
		}
		// Linear probe to "next" to this entry (really a rehash)
		hash++
	}
	var empty postings.MutableList
	return empty, false
}

// Set will set the value for an identifier.
func (m *Map) Set(k []byte, v postings.MutableList) {
	m.set(k, v, mapKeyOptions{
		copyKey:     true,
		finalizeKey: m.finalize != nil,
	})
}

// SetUnsafeOptions is a set of options to use when setting a value with
// the SetUnsafe method.
type SetUnsafeOptions struct {
	NoCopyKey     bool
	NoFinalizeKey bool
}

// SetUnsafe will set the value for an identifier with unsafe options for how
// the map treats the key.
func (m *Map) SetUnsafe(k []byte, v postings.MutableList, opts SetUnsafeOptions) {
	m.set(k, v, mapKeyOptions{
		copyKey:     !opts.NoCopyKey,
		finalizeKey: !opts.NoFinalizeKey,
	})
}

type mapKeyOptions struct {
	copyKey     bool
	finalizeKey bool
}

func (m *Map) set(k []byte, v postings.MutableList, opts mapKeyOptions) {
	hash := m.hash(k)
	for entry, ok := m.lookup[hash]; ok; entry, ok = m.lookup[hash] {
		if m.equals(entry.key.key, k) {
			m.lookup[hash] = MapEntry{
				key:   entry.key,
				value: v,
			}
			return
		}
		// Linear probe to "next" to this entry (really a rehash)
		hash++
	}

	m.lookup[hash] = MapEntry{
		key:   m.newMapKey(k, opts),
		value: v,
	}
}

// Iter provides the underlying map to allow for using a native Go for loop
// to iterate the map, however callers should only ever read and not write
// the map.
func (m *Map) Iter() map[MapHash]MapEntry {
	return m.lookup
}

// Len returns the number of map entries in the map.
func (m *Map) Len() int {
	return len(m.lookup)
}

// Contains returns true if value exists for key, false otherwise, it is
// shorthand for a call to Get that doesn't return the value.
func (m *Map) Contains(k []byte) bool {
	_, ok := m.Get(k)
	return ok
}

// Delete will remove a value set in the map for the specified key.
func (m *Map) Delete(k []byte) {
	hash := m.hash(k)
	for entry, ok := m.lookup[hash]; ok; entry, ok = m.lookup[hash] {
		if m.equals(entry.key.key, k) {
			m.removeMapKey(hash, entry.key)
			return
		}
		// Linear probe to "next" to this entry (really a rehash)
		hash++
	}
}

// Reset will reset the map by simply deleting all keys to avoid
// allocating a new map.
func (m *Map) Reset() {
	for hash, entry := range m.lookup {
		m.removeMapKey(hash, entry.key)
	}
}

// Reallocate will avoid deleting all keys and reallocate a new
// map, this is useful if you believe you have a large map and
// will not need to grow back to a similar size.
func (m *Map) Reallocate() {
	if m.initialSize > 0 {
		m.lookup = make(map[MapHash]MapEntry, m.initialSize)
	} else {
		m.lookup = make(map[MapHash]MapEntry)
	}
}
