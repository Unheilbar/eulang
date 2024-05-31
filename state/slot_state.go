package state

import (
	"github.com/ethereum/go-ethereum/common"
)

// slotState represents the state of one worker instance
type slotState struct {
	idx   int
	cache *StateDB

	// local dirties current transaction potential writings
	dirties       map[common.Hash]common.Hash
	mergedDirties map[common.Hash]common.Hash

	reads map[common.Hash]common.Hash

	// not nil only if tx is getting reexecuted
	updatedDirties map[common.Hash]common.Hash
}

func newSlotState(cache *StateDB, idx int) *slotState {
	return &slotState{
		idx:           idx,
		dirties:       make(map[common.Hash]common.Hash, 30),
		mergedDirties: make(map[common.Hash]common.Hash, 30),
		reads:         make(map[common.Hash]common.Hash, 30),
		cache:         cache,
	}
}

func (ss *slotState) GetState(key common.Hash) common.Hash {
	if ss.updatedDirties != nil {
		if val, ok := ss.updatedDirties[key]; ok {
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

func (ss *slotState) getDirties() map[common.Hash]common.Hash {
	return ss.mergedDirties
}

func (ss *slotState) validate(upd map[common.Hash]common.Hash) bool {
	for key := range ss.reads {
		if _, ok := upd[key]; ok {
			// one of our touches was found in a dirtyfall
			// tx needs to be reexecuted
			return false
		}
	}
	return true
}

// adds current worker dirties to upper worker dirties
func (ss *slotState) mergeIntoDirtyFall(upd map[common.Hash]common.Hash) {
	for k, v := range ss.dirties {
		upd[k] = v
	}
	ss.mergedDirties = upd
}

// doesn't remove logs
func (ss *slotState) revert() {
	ss.dirties = make(map[common.Hash]common.Hash, 30)
	ss.reads = make(map[common.Hash]common.Hash, 30)
}

// removes logs
func (ss *slotState) reset() {
	for k := range ss.dirties {
		delete(ss.dirties, k)
	}
	for k := range ss.reads {
		delete(ss.reads, k)
	}
	for k := range ss.mergedDirties {
		delete(ss.mergedDirties, k)
	}

	ss.updatedDirties = nil
}
