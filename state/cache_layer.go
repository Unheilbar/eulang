package state

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
)

const N = 16
const touchesPrealloc = 20

var cacheCapacity = 1000

type cacheLayer struct {
	backendKV map[common.Hash]common.Hash // trieDB

	cache *lru.Cache[common.Hash, common.Hash]

	pmx     sync.RWMutex
	pending map[common.Hash]common.Hash

	mx sync.RWMutex
}

func NewCacheLayer() *cacheLayer {
	return &cacheLayer{
		backendKV: make(map[common.Hash]common.Hash),
		cache:     lru.NewCache[common.Hash, common.Hash](cacheCapacity),
	}
}

// returns value and list of invalidated tx
func (cl *cacheLayer) get(key common.Hash) common.Hash {
	if val, ok := cl.pending[key]; ok {
		return val
	}

	if v, ok := cl.getFromCache(key); ok {
		return v
	}

	val := cl.backendKV[key]
	cl.addToCache(key, val)
	return val
}

func (cl *cacheLayer) addPending(key common.Hash, val common.Hash) {
	cl.pending[key] = val
}

func (cl *cacheLayer) addToCache(key common.Hash, val common.Hash) {
	cl.mx.Lock()
	cl.cache.Add(key, val)
	cl.mx.Unlock()
}

func (cl *cacheLayer) getFromCache(key common.Hash) (val common.Hash, ok bool) {
	cl.mx.RLock()
	val, ok = cl.cache.Get(key)
	cl.mx.RUnlock()
	return
}
