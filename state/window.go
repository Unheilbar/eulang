package state

import (
	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	relIdx int // relative tx index in current execution slot
	input  []byte
}
type Receipt struct {
	idx int
}

var contractAddress = common.BytesToAddress([]byte("Cats have to be worshiped as gods!"))

type window struct {
	result chan *Receipt // txes that are ready for finilize
	state  *StateDB
	size   int
	slots  []*executionSlot
}

func NewWindow(state *StateDB, size int) *window {
	result := make(chan *Receipt)

	executionSlots := make([]*executionSlot, size, size)
	for i := 0; i < size; i++ {
		executionSlots[i] = newSlot(state, result, i, executionSlots)
	}

	return &window{
		state:  state,
		result: result,
		size:   size,
		slots:  executionSlots,
	}
}

func (win *window) process(txes []*Transaction) []*Receipt {
	if len(txes) > win.size {
		panic("invalid usage of window")
	}

	for i, tx := range txes {
		wIdx := win.size - i - 1
		go win.slots[wIdx].process(tx)
	}

	var receipts []*Receipt
	// gather tx receipts here
	for r := range win.result {
		receipts = append(receipts, r)
		if len(receipts) == len(txes) {
			break
		}
	}

	win.finilize()
	return receipts
}

func (win *window) finilize() {
	win.state.updatePendings()
	for i := 0; i < win.size; i++ {
		win.slots[i].reset()
	}
}

func (win *window) commit() {

}
