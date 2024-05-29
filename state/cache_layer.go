package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
)

const N = 16
const touchesPrealloc = 20

var cacheCapacity = 1000

type StateDB struct {
	backendKV map[common.Hash]common.Hash // trieDB

	cache *lru.Cache[common.Hash, common.Hash]

	pending map[common.Hash]common.Hash
}

func NewState() *StateDB {
	return &StateDB{
		backendKV: make(map[common.Hash]common.Hash),
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
