package state

import "github.com/ethereum/go-ethereum/common"

// slotState represents the state of one worker instance
type slotState struct {
	priority int

	cache *cacheLayer

	// context of current execution
	tx         *tx
	dirties    map[common.Hash]common.Hash
	stateReads map[common.Hash]common.Hash
}

func newSlotState(cache *cacheLayer) *slotState {
	return &slotState{
		dirties: make(map[common.Hash]common.Hash, 30),
		cache:   cache,
	}
}

func (ss *slotState) GetState(key common.Hash) common.Hash {
	if val, ok := ss.dirties[key]; ok {
		return val
	}
	val := ss.cache.get(key)
	ss.stateReads[key] = val
	return val
}

func (ss *slotState) SetState(key common.Hash, val common.Hash) {
	ss.dirties[key] = val
}

func (ss *slotState) revert() {
	ss.dirties = make(map[common.Hash]common.Hash)
	ss.stateReads = make(map[common.Hash]common.Hash)
}

// all dirties go to pending of underlying cache layer, report all dirties
func (ss *slotState) finilize() {
	for k, v := range ss.dirties {
		ss.cache.setPending(k, v)
	}
}

func (ss *slotState) reset()
