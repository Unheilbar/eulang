package state

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type touchWrite struct {
	tx      *tx
	prevVal common.Hash
}

type storageChange struct {
	key       common.Hash
	prevValue common.Hash
}

type ParallelState struct {
	mx sync.Mutex
	kv map[common.Hash]common.Hash // kv represents any key val storage underhood our state. (for instance triedb)

	allTouches   map[common.Hash][]*tx // touches represent access to state in current execution window where key is touched memory slot
	touchesWrite map[common.Hash][]touchWrite

	storageChanges map[common.Hash]storageChange
	revert         chan common.Hash
}

func New() *ParallelState {
	return &ParallelState{
		kv: make(map[common.Hash]common.Hash, 2),

		allTouches:   make(map[common.Hash][]*tx, 2),
		touchesWrite: make(map[common.Hash][]touchWrite, 2),

		storageChanges: make(map[common.Hash]storageChange, 2),
	}
}

func (s *ParallelState) SetState(t *tx, key common.Hash, val common.Hash) {
	s.mx.Lock()
	defer s.mx.Unlock()

	redo := make(map[common.Hash]*tx, 0)

	if touches, ok := s.allTouches[key]; ok {
		// slot was touched by other txes. revert and redo all txes with less priority than ours
		// TODO should also remove all touches by that tx
		for _, touchTx := range touches {
			if touchTx.priority < t.priority {
				sch, ok := s.storageChanges[touchTx.hash]
				if ok {
					s.kv[sch.key] = sch.prevValue
				}
				redo[touchTx.hash] = touchTx
			}
		}
	}

	for _, t := range redo {
		go func(t *tx) {
			t.redo <- t
		}(t)
	}

	s.allTouches[key] = append(s.allTouches[key], t)
	s.touchesWrite[key] = append(s.touchesWrite[key], touchWrite{
		tx:      t,
		prevVal: s.kv[key],
	})

	if _, ok := s.storageChanges[t.hash]; !ok {
		s.storageChanges[t.hash] = storageChange{
			key:       key,
			prevValue: s.kv[key],
		}
	}

	s.kv[key] = val
}

func (s *ParallelState) GetState(tx *tx, key common.Hash) (val common.Hash) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.allTouches[key] = append(s.allTouches[key], tx)

	if txs, ok := s.touchesWrite[key]; ok {
		// key was written by other txs. if tx priority of any of them is higher than ours we just take the last updated
		// otherwise we get the prev value from the lowest tx priority
		var minVal common.Hash
		var minPriority = tx.priority
		for _, write := range txs {
			if write.tx.priority < minPriority {
				minVal = write.prevVal
			} else {
				return s.kv[key]
			}
		}
		return minVal
	}

	return s.kv[key]
}

func (s *ParallelState) Reset() {
	s.allTouches = make(map[common.Hash][]*tx, 2)
	s.touchesWrite = make(map[common.Hash][]touchWrite, 2)
	s.storageChanges = make(map[common.Hash]storageChange)
}
