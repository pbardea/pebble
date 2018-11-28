// Copyright 2018 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package pebble

// rangeDelLevel holds the state for a single level in rangeDelMap. Levels come
// in two flavors: single-table levels and multi-table levels. Single table
// labels are initialized with a range-del iterator when the rangeDelMap is
// created. Multi-table levels are connected to levelIter and are lazily
// populated as tables are loaded.
type rangeDelLevel struct {
	m    *rangeDelMap
	iter internalIterator
}

func (l *rangeDelLevel) init(iter internalIterator) {
	l.iter = iter
}

// load the range-del iterator for the specified table.
func (l *rangeDelLevel) load(meta *fileMetadata) error {
	if l.iter != nil {
		if err := l.iter.Close(); err != nil {
			return err
		}
	}

	var err error
	l.iter, err = l.m.newIter(meta)
	return err
}

// rangeDelMap provides a merged view of the range tombstones from multiple
// levels. The map is composed of a series of levels, mirroring the levels in
// the LSM tree, though L0 is exploded into a level per table, and each
// memtable is on its own level.
type rangeDelMap struct {
	// The sequence number at which reads are being performed. Tombstones that
	// are newer than this sequence number are ignored.
	seqNum uint64
	// The callback for creating new range-del iterators.
	newIter tableNewIter
	levels  []rangeDelLevel
}

func (m *rangeDelMap) init(seqNum uint64, newIter tableNewIter) {
	m.seqNum = seqNum
	m.newIter = newIter
}

func (m *rangeDelMap) addLevel(iter internalIterator) {
	m.levels = append(m.levels, rangeDelLevel{
		m:    m,
		iter: iter,
	})
}

func (m *rangeDelMap) addLevels(n int) []rangeDelLevel {
	for i := 0; i < n; i++ {
		m.levels = append(m.levels, rangeDelLevel{
			m: m,
		})
	}
	return m.levels[len(m.levels)-n:]
}

func (m *rangeDelMap) Close() error {
	var err error
	for i := range m.levels {
		if m.levels[i].iter != nil {
			if err1 := m.levels[i].iter.Close(); err1 != nil && err == nil {
				err = err1
			}
		}
	}
	return err
}