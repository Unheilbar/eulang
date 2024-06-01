package state

import (
	"maps"

	"github.com/ethereum/go-ethereum/common"
)

// slotState represents the state of one worker instance
type slotState struct {
	idx   int
	cache *StateDB

	// local dirties current transaction potential writings
	dirties map[common.Hash]common.Hash

	reads map[common.Hash]common.Hash

	// flag set true when it's reexecution
	reexec bool
}

func newSlotState(cache *StateDB, idx int) *slotState {
	return &slotState{
		idx:     idx,
		dirties: make(map[common.Hash]common.Hash, 30),
		reads:   make(map[common.Hash]common.Hash, 30),
		cache:   cache,
	}
}

func (ss *slotState) GetState(key common.Hash) common.Hash {
	if ss.reexec {
		if val, ok := ss.cache.mergeDirties[key]; ok {
			return val
		}
	}

	if val, ok := ss.dirties[key]; ok {
		return val
	}

	if val, ok := ss.reads[key]; ok {
		return val
	}

	val := ss.cache.get(key)

	ss.reads[key] = val

	return val
}

func (ss *slotState) SetState(key common.Hash, val common.Hash) {
	ss.dirties[key] = val
}

func (ss *slotState) validate() bool {
	for key := range ss.reads {
		if _, ok := ss.cache.mergeDirties[key]; ok {
			// one of our touches was found in a dirtyfall
			// tx needs to be reexecuted
			return false
		}
	}
	return true
}

// adds current worker dirties to upper worker dirties
func (ss *slotState) mergeDirties() {
	maps.Copy(ss.cache.mergeDirties, ss.dirties)
}

func (ss *slotState) setReexec() {
	ss.reexec = true
}

// doesn't remove logs
func (ss *slotState) revert() {
	ss.reexec = false
	for k := range ss.dirties {
		delete(ss.dirties, k)
	}
	for k := range ss.reads {
		delete(ss.reads, k)
	}
}

// removes logs
func (ss *slotState) reset() {
	ss.reexec = false
	for k := range ss.dirties {
		delete(ss.dirties, k)
	}
	for k := range ss.reads {
		delete(ss.reads, k)
	}
}
