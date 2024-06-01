package state

import (
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const txAmount = 30
const conflictRate = 0 // how many txes in window are in conflict

func Benchmark_window(b *testing.B) {
	b.StopTimer()
	state := NewState()
	win := NewWindow(state, txAmount)
	txes := generateTxes(txAmount, conflictRate)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		win.Process(txes)
	}
}

func Test_windowFuzz(t *testing.T) {
	expResult := testWindow()
	for i := 0; i < 10000; i++ {
		res := testWindow()
		assert(t, expResult, res)
	}
}

func assert(t *testing.T, exp map[common.Hash]common.Hash, act map[common.Hash]common.Hash) {
	for k, v := range exp {
		if act[k] != v {
			t.Fatal("wrong result")
		}
	}
}

func testWindow() map[common.Hash]common.Hash {
	state := NewState()
	win := NewWindow(state, txAmount)
	txes := generateTxes(txAmount, conflictRate)

	win.Process(txes)

	for _, tx := range txes {
		_, ok := state.pending[tx.writeKey]
		if !ok {
			log.Println("key not found in pendings", tx.writeKey)
		}
	}

	return state.pending
}

func generateTxes(amount int64, conflictRate int) []*tx {
	txes := make([]*tx, 0)

	for i := int64(0); i < amount; i++ {
		h := common.BigToHash(big.NewInt(i))
		txes = append(txes, &tx{
			idx:      int(i),
			hash:     h,
			readKey:  common.BigToHash(big.NewInt(i)),
			writeKey: common.BigToHash(big.NewInt(i + amount)),
		})
	}

	for i := 1; i <= int(conflictRate); i++ {
		txes[i].readKey = txes[i-1].writeKey

	}

	return txes
}
