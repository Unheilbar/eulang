package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type tx struct {
	priority int
	hash     common.Hash
	key      common.Hash
	val      common.Hash
	redo     chan *tx
}

type workerStatus uint8

const (
	workerStatusInProgress workerStatus = iota // executing tx
	workerStatusIdle                           // waiting for all workers with higher priority to finish their executions
	workerStatusDone                           //
)

type worker struct {
	idx    int
	pool   []*worker
	status workerStatus
	state  *slotState
	done   chan struct{}

	result chan *tx
}

func (w *worker) process(t *tx) {
	t.priority = w.idx
	w.tryExec(t)

	if w.idx == len(w.pool)-1 {
		w.status = workerStatusDone
		w.result <- t
		close(w.done)
		return
	}

	w.status = workerStatusIdle
idle_loop:
	for {
		select {
		case <-t.redo:
			w.tryExec(t)
		case <-w.pool[w.idx+1].done:
			w.status = workerStatusDone
			close(w.done)
			w.state.finilize() // tx executed the tx with lower priority can be executed
			w.result <- t
			break idle_loop
		}
	}

}

// some exec in a future using vm
func (w *worker) tryExec(tx *tx) {
	w.status = workerStatusInProgress
	val := w.state.GetState(tx.key)
	nval := new(uint256.Int)
	nval.SetBytes(val.Bytes())
	nval.Add(nval, nval)
	w.state.SetState(tx.key, nval.Bytes32())
}

func newWorker(state *cacheLayer, result chan *tx, idx int, pool []*worker) *worker {
	return &worker{
		pool:   pool,
		idx:    idx,
		result: result,
		state:  newSlotState(state),
		done:   make(chan struct{}),
	}
}

type window struct {
	result chan *tx // txes that are ready for finilize

	size    int
	workers []*worker
}

func NewWindow(state *cacheLayer, size int) *window {
	result := make(chan *tx)

	workers := make([]*worker, size, size)
	for i := 0; i < size; i++ {
		workers[i] = newWorker(state, result, i, workers)
	}

	return &window{
		result:  result,
		size:    size,
		workers: workers,
	}
}

func (win *window) Process(txes []*tx) {
	if len(txes) != win.size {
		panic("invalid usage of window")
	}

	for i := len(txes) - 1; i >= 0; i-- {
		txes[i].redo = make(chan *tx)
		go win.workers[i].process(txes[i])
	}

	var got int
	for range win.result {
		got++
		if got == len(txes) {
			break
		}
	}

}

func (win *window) Reset() {
	for i := 0; i < win.size; i++ {
		win.workers[i].done = make(chan struct{})
	}
}
