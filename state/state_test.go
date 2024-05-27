package state

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type tx struct {
	hash common.Hash
	key  common.Hash
	val  common.Hash
}

const txAmount = 10
const conflictRate = 5 //every 10s key repeats
const workers = 10

func Test_State(t *testing.T) {
	revert := make(chan common.Hash)
	state := New(revert)

	txes := generateTxes()
	incoming := make(chan tx, workers)

	var processedTx atomic.Int64
	{ // write txes into incoming
		go func() {
			for _, v := range txes {
				incoming <- v
			}
		}()
	}

	{
		go func() {
			for tx := range revert {
				fmt.Println("revert")
				processedTx.Add(-1)
				incoming <- txes[tx]
			}
		}()

	}
	var wg sync.WaitGroup
	wg.Add(workers)

	{ // read from incoming and update state
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for tx := range incoming {
					val := state.GetState(tx.hash, tx.key)
					nval := new(uint256.Int)
					nval.SetBytes(val.Bytes())
					nval.Add(nval, nval)
					state.SetState(tx.hash, tx.key, nval.Bytes32())
					processedTx.Add(1)
				}
			}()
		}
	}

	{ // wait till all be processed
		go func() {
			for processedTx.Load() != txAmount {
			}
			close(incoming)
		}()
	}

	wg.Wait()

}

func generateTxes() map[common.Hash]tx {
	txes := make(map[common.Hash]tx)

	for i := int64(0); i < txAmount; i++ {
		h := common.BigToHash(big.NewInt(i))

		txes[h] = tx{
			hash: h,
			key:  common.BigToHash(big.NewInt(i % conflictRate)),
			val:  common.BigToHash(big.NewInt(i * 10)),
		}
	}

	return txes
}
