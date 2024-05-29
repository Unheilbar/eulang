package state

import (
	"github.com/ethereum/go-ethereum/common"
)

// slotState represents the state of one worker instance
type slotState struct {
	cache *StateDB

	// dirties current transaction potential writings
	dirties     map[common.Hash]common.Hash
	copyDirties map[common.Hash]common.Hash
	reads       map[common.Hash]common.Hash

	// not nil only if tx is getting reexecuted
	updatedDirties map[common.Hash]common.Hash
}

func newSlotState(cache *StateDB) *slotState {
	return &slotState{
		dirties:     make(map[common.Hash]common.Hash, 30),
		copyDirties: make(map[common.Hash]common.Hash, 30),
		reads:       make(map[common.Hash]common.Hash, 30),
		cache:       cache,
	}
}

func (ss *slotState) GetState(key common.Hash) common.Hash {
	if ss.updatedDirties != nil {
		return ss.updatedDirties[key]
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
	ss.copyDirties[key] = val
	ss.dirties[key] = val
}

func (ss *slotState) getDirties() map[common.Hash]common.Hash {
	return ss.copyDirties
}

func (ss *slotState) validate(upd map[common.Hash]common.Hash) bool {
	for _, key := range ss.reads {
		if _, ok := upd[key]; ok {
			// one of our touches was found in a dirtyfall
			// tx needs to be reexecuted
			return false
		}
	}
	return true
}

func (ss *slotState) mergeIntoDirtyFall(upd map[common.Hash]common.Hash) {
	for k, v := range ss.dirties {
		upd[k] = v
	}
	ss.copyDirties = upd
}

func (ss *slotState) revert() {
	ss.dirties = make(map[common.Hash]common.Hash, 30)
	ss.reads = make(map[common.Hash]common.Hash, 30)
	ss.updatedDirties = nil
}

// all dirties go to pending of underlying cache layer
func (ss *slotState) finilize() {}
