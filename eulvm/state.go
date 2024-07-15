package eulvm

import (
	"github.com/ethereum/go-ethereum/common"
)

type State interface {
	SetState(key common.Hash, val common.Hash)
	GetState(key common.Hash) (val common.Hash)
	Reset()
}

type MockState struct {
	store map[common.Hash]common.Hash
}

func NewMopState() *MockState {
	return &MockState{
		store: make(map[common.Hash]common.Hash),
	}
}

func (ms *MockState) SetState(key common.Hash, val common.Hash) {
	ms.store[key] = val
}
func (ms *MockState) GetState(key common.Hash) (val common.Hash) {
	return ms.store[key]
}
func (ms *MockState) Reset() {
	clear(ms.store)
}
