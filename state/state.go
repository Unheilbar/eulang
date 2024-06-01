package state

import (
	"maps"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
)

var cacheCapacity = 1000

type StateDB struct {
	backendKV map[common.Hash]common.Hash // trieDB lifetime - persistent

	cache *lru.Cache[common.Hash, common.Hash] // lifetime - uptime

	pending map[common.Hash]common.Hash // lifetime - block

	mergeDirties map[common.Hash]common.Hash
}

func NewState() *StateDB {
	return &StateDB{
		backendKV:    make(map[common.Hash]common.Hash, 20),
		pending:      make(map[common.Hash]common.Hash, 20),
		mergeDirties: make(map[common.Hash]common.Hash, 200),
		cache:        lru.NewCache[common.Hash, common.Hash](cacheCapacity),
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

func (sb *StateDB) updatePendings() {
	maps.Copy(sb.pending, sb.mergeDirties)
	sb.clearDirties()
}

func (sb *StateDB) clearDirties() {
	clear(sb.mergeDirties)
}

func (sb *StateDB) Commit() {
	for k, v := range sb.pending {
		sb.cache.Add(k, v)
		sb.backendKV[k] = v
		delete(sb.pending, k)
		delete(sb.mergeDirties, k)
	}
}
