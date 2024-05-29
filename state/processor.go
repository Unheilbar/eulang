package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type tx struct {
	hash common.Hash
	key  common.Hash
	val  common.Hash
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
	w.state.revert() // validation failed, revert all tx changes and exec it with updatedDirties
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
		close(w.result)
	}
	close(w.doneCh)
}

// some exec in a future using vm
func (w *worker) tryExec(tx *tx) {
	w.status = workerStatusInProgress
	val := w.state.GetState(tx.key)
	nval := new(uint256.Int)
	nval.SetBytes(val.Bytes())
	nval.Add(nval, uint256.NewInt(9))
	w.state.SetState(tx.key, nval.Bytes32())
}

func newWorker(state *StateDB, result chan *tx, idx int, pool []*worker) *worker {
	return &worker{
		pool:         pool,
		idx:          idx,
		result:       result,
		state:        newSlotState(state),
		doneCh:       make(chan struct{}),
		upperDirties: make(chan map[common.Hash]common.Hash),
	}
}

type window struct {
	result chan *tx // txes that are ready for finilize

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
		go win.workers[i].process(tx)
	}

	for range win.result {
	}

}

func (win *window) Reset() {
	for i := 0; i < win.size; i++ {
		win.workers[i].doneCh = make(chan struct{})
	}
}
