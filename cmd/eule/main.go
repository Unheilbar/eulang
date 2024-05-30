package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Unheilbar/eulang/compiler"
	"github.com/Unheilbar/eulang/eulvm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func main() {
	file := os.Args[1]
	eulang := compiler.NewEulang()
	prog := compiler.CompileFromSource(eulang, file)
	//e := eulvm.New(prog).WithDebug()
	e := eulvm.New(prog)
	input := eulang.GenerateInput(os.Args[2], os.Args[3:])
	err := e.Run(input)
	if err != nil {
		log.Fatal(err)
	}

	//todo normal return parser
	retData1 := e.GetReternData()[0]
	dec := uint256.Int(retData1)
	d := dec.Uint64()
	if d == 0 {
		return
	}

	retData2 := e.GetReternData()[1]
	dec2 := uint256.Int(retData2)
	d2 := dec2.Uint64()

	retData3 := e.GetReternData()[2]
	dec3 := uint256.Int(retData3)
	d3 := dec3.Uint64()

	var addr common.Address

	retData4 := e.GetReternData()[3]

	addr.SetBytes(retData4.Bytes())

	fmt.Println("return: ", d, d2, d3, addr)

}
