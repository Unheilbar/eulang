package state

import "github.com/ethereum/go-ethereum/common"

// slotState represents the state of one worker instance
type slotState struct {
	priority int

	cache *cacheLayer

	// context of current execution
	tx      *tx
	dirties map[common.Hash]common.Hash

	stateReads   map[common.Hash]struct{}
	stateUpdates map[common.Hash]struct{}
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

	ss.cache
}

func (ss *slotState) SetState(key common.Hash, val common.Hash) {

}

func (ss *slotState) reset()
