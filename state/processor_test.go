package state

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const txAmount = 16
const conflictPercentage = 10 // how many txes in window are in conflict

func Benchmark_window(b *testing.B) {
	b.StopTimer()
	state := NewState()
	win := NewWindow(state, txAmount)
	txes := generateTxes(txAmount, conflictPercentage)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		win.Process(txes)
	}
}

func Test_windowFuzz(t *testing.T) {
	expResult := testWindow()
	for i := 0; i < 100000; i++ {
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
	txes := generateTxes(txAmount, conflictPercentage)

	win.Process(txes)

	return state.pending
}

func generateTxes(amount int64, conflictRate int64) []*tx {
	txes := make([]*tx, 0)

	var mod int64
	if conflictRate == 0 {
		mod = amount
	} else if conflictRate >= 100 {
		mod = 1
	} else {
		mod = amount - int64(float64(conflictRate)/100*float64(amount))
	}

	for i := int64(0); i < amount; i++ {
		h := common.BigToHash(big.NewInt(i))

		txes = append(txes, &tx{
			hash: h,
			key:  common.BigToHash(big.NewInt(i % mod)),
			val:  common.BigToHash(big.NewInt(i*2 + 10)),
		})
	}

	return txes
}
