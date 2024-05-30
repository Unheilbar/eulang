package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
)

const N = 16
const touchesPrealloc = 20

var cacheCapacity = 1000

type StateDB struct {
	backendKV map[common.Hash]common.Hash // trieDB lifetime - persistent

	cache *lru.Cache[common.Hash, common.Hash] // lifetime - uptime

	pending map[common.Hash]common.Hash // lifetime - block
}

func NewState() *StateDB {
	return &StateDB{
		backendKV: make(map[common.Hash]common.Hash),
		pending:   make(map[common.Hash]common.Hash),
		cache:     lru.NewCache[common.Hash, common.Hash](cacheCapacity),
	}
}

func (sb *StateDB) get(key common.Hash) common.Hash {
	// first try to get pending
	if val, ok := sb.pending[key]; ok {
		return val
	}

	// try to get from cache
	if val, ok := sb.cache.Get(key); ok {
		return val
	}

	// get from backend
	val := sb.backendKV[key]

	return val
}

func (sb *StateDB) updatePendings(dirties map[common.Hash]common.Hash) {
	for k, v := range dirties {
		sb.pending[k] = v
	}
}
