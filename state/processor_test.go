package state

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const txAmount = 10
const conflictPercentage = 30 // how many txes in window are in conflict

func Test_window(t *testing.T) {
	win := NewWindow(txAmount)
	txes := generateTxes(txAmount, conflictPercentage)

	win.Process(txes)
}

func generateTxes(amount int64, conflictRate int64) []*tx {
	txes := make([]*tx, 0)

	var mod int64
	if conflictRate == 0 {
		mod = amount
	} else if conflictRate == 100 {
		mod = 1
	} else {
		mod = amount - int64(float64(conflictRate)/100*float64(amount))
	}

	for i := int64(0); i < amount; i++ {
		h := common.BigToHash(big.NewInt(i))

		txes = append(txes, &tx{
			hash: h,
			key:  common.BigToHash(big.NewInt(i % mod)),
			val:  common.BigToHash(big.NewInt(i * 10)),
		})
	}

	return txes
}
