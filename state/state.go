package state

import (
	"github.com/ethereum/go-ethereum/common"
)

type State struct {
	state map[common.Hash]common.Hash // TODO later use actual stateDB as storage backend. map can be used for temporary map storage inside of smart contract
}

func New() *State {
	return &State{
		state: make(map[common.Hash]common.Hash),
	}
}

func (s *State) SetState(key common.Hash, val common.Hash) {
	s.state[key] = val
}
func (s *State) GetState(key common.Hash) (val common.Hash) {
	return s.state[key]
}
