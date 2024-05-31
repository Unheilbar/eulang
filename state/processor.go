package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type tx struct {
	idx      int
	hash     common.Hash
	readKey  common.Hash
	writeKey common.Hash
}

type workerStatus uint8

const (
	workerStatusInProgress workerStatus = iota // executing tx
	workerStatusIdle                           // waiting for all workers with higher priority to finish their executions
	workerStatusDone                           //
)

type worker struct {
	idx          int
	pool         []*worker
	status       workerStatus
	state        *slotState
	doneCh       chan struct{}
	upperDirties chan map[common.Hash]common.Hash

	result chan *tx
}

func (w *worker) process(t *tx) {
	w.tryExec(t)

	if w.idx == len(w.pool)-1 {
		w.state.mergeIntoDirtyFall(w.state.dirties)
		w.done(t)
		return
	}

	w.status = workerStatusIdle
	<-w.pool[w.idx+1].doneCh
	fd := w.pool[w.idx+1].dirties()
	if ok := w.state.validate(fd); ok {
		w.state.mergeIntoDirtyFall(fd)
		w.done(t)
		return
	}

	w.status = workerStatusInProgress
	w.state.reset() // validation failed, revert all tx changes and exec it with updatedDirties
	w.state.updatedDirties = fd
	w.tryExec(t)
	w.state.mergeIntoDirtyFall(fd)
	w.done(t)
}

func (w *worker) dirties() map[common.Hash]common.Hash {
	return w.state.getDirties()
}

func (w *worker) done(t *tx) {
	w.status = workerStatusDone
	w.result <- t
	// report dirties to lower priority worker
	if w.idx == 0 {
		return
	}
	w.doneCh <- struct{}{}
}

// some exec in a future using vm
func (w *worker) tryExec(tx *tx) {
	w.status = workerStatusInProgress
	w.state.GetState(tx.readKey)
	nval := uint256.NewInt(200)
	w.state.SetState(tx.writeKey, nval.Bytes32())
}

func (w *worker) reset() {
	w.state.reset()
}

func newWorker(state *StateDB, result chan *tx, idx int, pool []*worker) *worker {
	return &worker{
		pool:         pool,
		idx:          idx,
		result:       result,
		state:        newSlotState(state, idx),
		doneCh:       make(chan struct{}),
		upperDirties: make(chan map[common.Hash]common.Hash),
	}
}

type window struct {
	result  chan *tx // txes that are ready for finilize
	state   *StateDB
	size    int
	workers []*worker
}

func NewWindow(state *StateDB, size int) *window {
	result := make(chan *tx)

	workers := make([]*worker, size, size)
	for i := 0; i < size; i++ {
		workers[i] = newWorker(state, result, i, workers)
	}

	return &window{
		state:   state,
		result:  result,
		size:    size,
		workers: workers,
	}
}

func (win *window) Process(txes []*tx) {
	if len(txes) != win.size {
		panic("invalid usage of window")
	}

	for i, tx := range txes {
		wIdx := win.size - i - 1
		go win.workers[wIdx].process(tx)
	}

	// gather tx receipts here
	var acceptedTx int
	for range win.result {
		acceptedTx++
		if acceptedTx == win.size {
			break
		}
	}
	acceptedTx = 0

	win.finilize()
}

func (win *window) finilize() {
	win.state.updatePendings(win.workers[0].state.mergedDirties)
	for i := 0; i < win.size; i++ {
		win.workers[i].reset()
	}
}
