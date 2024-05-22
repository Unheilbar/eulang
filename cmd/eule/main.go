package main

import (
	"log"
	"os"

	"github.com/Unheilbar/eulang/compiler"
	"github.com/Unheilbar/eulang/eulvm"
)

func main() {
	file := os.Args[1]
	eulang := compiler.NewEulang()
	prog := compiler.CompileFromSource(eulang, file)
	//e := eulvm.New(prog).WithDebug()
	e := eulvm.New(prog)
	input := eulang.GenerateInput(os.Args[2])

	err := e.Run(input)
	if err != nil {
		log.Fatal(err)
	}
}
