package main

import (
	"os"

	"github.com/Unheilbar/eulang/compiler"
	"github.com/Unheilbar/eulang/eulvm"
)

func main() {
	file := os.Args[1]
	program := compiler.CompileEasmFromFile(file, "")
	elvm := eulvm.New(program)

	elvm.Run()
}
