package state

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type touchWrite struct {
	txHash common.Hash
	prev   common.Hash
}

type storageChange struct {
	key       common.Hash
	prevValue common.Hash
}

type ParallelState struct {
	mx sync.Mutex
	kv map[common.Hash]common.Hash // kv represents any key val storage underhood our state. (for instance triedb)

	allTouches   map[common.Hash][]common.Hash // touches represent access to state in current execution window where key is touched memory slot
	touchesWrite map[common.Hash][]touchWrite

	storageChanges map[common.Hash]storageChange
	revert         chan common.Hash
}

func New(revert chan common.Hash) *ParallelState {
	return &ParallelState{
		revert:         revert,
		kv:             make(map[common.Hash]common.Hash, 2),
		allTouches:     make(map[common.Hash][]common.Hash, 2),
		touchesWrite:   make(map[common.Hash][]touchWrite, 2),
		storageChanges: make(map[common.Hash]storageChange, 2),
	}
}

func (s *ParallelState) SetState(txHash common.Hash, key common.Hash, val common.Hash) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if touches, ok := s.allTouches[key]; ok {
		// slot was touched by other txes. revert and redo all txes with less priority than ours
		for _, touchTxHash := range touches {
			if touchTxHash.Cmp(txHash) < 0 {
				sch := s.storageChanges[touchTxHash]
				s.kv[sch.key] = sch.prevValue
				s.revert <- touchTxHash
			}
		}
	}

	s.allTouches[key] = append(s.allTouches[key], txHash)
	s.touchesWrite[key] = append(s.touchesWrite[key], touchWrite{
		txHash: txHash,
		prev:   s.kv[key],
	})

	if _, ok := s.storageChanges[txHash]; !ok {
		s.storageChanges[txHash] = storageChange{
			key:       key,
			prevValue: s.kv[key],
		}
	}

	s.kv[key] = val
}

func (s *ParallelState) GetState(txHash common.Hash, key common.Hash) (val common.Hash) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.allTouches[key] = append(s.allTouches[key], txHash)

	if txs, ok := s.touchesWrite[key]; ok {
		// key was written by other txs. if tx priority of any of them is higher than ours we just take the last updated
		// otherwise we get the key from the lowest tx priority
		var minVal common.Hash
		var minTxHash = txHash
		for _, write := range txs {
			if write.txHash.Cmp(minTxHash) < 0 {
				minTxHash = write.txHash
				minVal = write.prev
			} else {
				return s.kv[key]
			}
		}
		return minVal
	}

	return s.kv[key]
}

func (s *ParallelState) Reset() {
	s.allTouches = make(map[common.Hash][]common.Hash, 2)
	s.touchesWrite = make(map[common.Hash][]touchWrite, 2)
	s.storageChanges = make(map[common.Hash]storageChange)
}
