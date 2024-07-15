package state

import (
	"github.com/Unheilbar/eulang/eulvm"
	"github.com/ethereum/go-ethereum/common"
)

type executionSlotStatus uint8

const (
	executionSlotStatusInProgress executionSlotStatus = iota // executing tx
	executionSlotStatusIdle                                  // waiting for all workers with higher priority to finish their executions
	executionSlotStatusDone                                  //
)

type executionSlot struct {
	idx          int
	pool         []*executionSlot
	status       executionSlotStatus
	state        *slotState
	doneCh       chan struct{}
	upperDirties chan map[common.Hash]common.Hash
	vm           *eulvm.EulVM

	result chan *Receipt
}

func (w *executionSlot) process(t *Transaction) {
	w.tryExec(t)

	if w.idx == len(w.pool)-1 {
		w.state.mergeDirties()
		w.done()
		return
	}

	w.status = executionSlotStatusIdle
	<-w.pool[w.idx+1].doneCh
	if ok := w.state.validate(); ok {
		w.state.mergeDirties()
		w.done()
		return
	}

	w.status = executionSlotStatusInProgress
	w.vm.Reset() // validation failed, revert all tx changes and exec it with updatedDirties
	w.state.reset()
	w.state.setReexec()
	w.tryExec(t)
	w.state.mergeDirties()
	w.done()
}

func (w *executionSlot) done() {
	w.status = executionSlotStatusDone
	// TODO later form receipt
	w.result <- &Receipt{}
	// report dirties to lower priority executionSlot
	if w.idx == 0 {
		return
	}

	w.doneCh <- struct{}{}
}

// some exec in a future using vm
func (w *executionSlot) tryExec(tx *Transaction) {
	w.vm.Run(tx.input)
}

func (w *executionSlot) reset() {
	w.state.reset()
	w.vm.Reset()
}

func newSlot(state *StateDB, result chan *Receipt, idx int, pool []*executionSlot) *executionSlot {
	slotState := newSlotState(state, idx)
	vm := eulvm.New(slotState,
		slotState.cache.GetCode())
	return &executionSlot{
		pool:   pool,
		idx:    idx,
		result: result,
		state:  slotState,
		doneCh: make(chan struct{}),
		vm:     vm,
	}
}
