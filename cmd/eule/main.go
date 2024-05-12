package main

import (
	"log"
	"os"

	"github.com/Unheilbar/eulang/compiler"
	"github.com/Unheilbar/eulang/eulvm"
)

func main() {
	file := os.Args[1]

	prog := compiler.CompileFromSource(file)
	//for idx, inst := range prog.Instrutions {
	//	fmt.Println(idx, eulvm.OpCodes[inst.OpCode], inst.Operand.Uint64())

	//}
	e := eulvm.New(prog)
	err := e.Run()
	if err != nil {
		log.Fatal(err)
	}
	//e.Dump()
}
