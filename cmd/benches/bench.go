package main

import (
	"fmt"
	"math/big"

	"github.com/Unheilbar/eulang/compiler"
	"github.com/Unheilbar/eulang/eulvm"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	_ = prepareBench()
}

type benchSuit struct {
	vm *eulvm.EulVM

	wallet1 common.Hash
	wallet2 common.Hash

	transferCall []byte
	checkCall    []byte
}

const emissionAmount = 1000000
const transferAmount = 1

func prepareBench() *benchSuit {
	file := "/home/allison/src/eulang/examples/demo_transfer.eul"
	eulang := compiler.NewEulang()
	prog := compiler.CompileFromSource(eulang, file)
	//e := eulvm.New(prog).WithDebug()
	e := eulvm.New(prog)

	var bs = &benchSuit{}

	bs.vm = e

	bs.wallet1 = common.BigToHash(big.NewInt(1))
	bs.wallet2 = common.BigToHash(big.NewInt(2))

	emission := eulang.GenerateInput("emission", []string{bs.wallet1.Hex(), fmt.Sprint(emissionAmount)})

	err := e.Run(emission)
	if err != nil {
		panic(err)
	}
	e.Reset()

	bs.transferCall = eulang.GenerateInput("transfer", []string{bs.wallet1.Hex(), bs.wallet2.Hex(), fmt.Sprint(transferAmount)})
	bs.checkCall = eulang.GenerateInput("checkBalance", []string{bs.wallet2.Hex(), fmt.Sprint(1)})

	return bs
}
