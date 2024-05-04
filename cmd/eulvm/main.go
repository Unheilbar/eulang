package main

import (
	"os"

	"github.com/Unheilbar/eulang/compiler"
	"github.com/Unheilbar/eulang/euvm"
)

func main() {
	interpreter := euvm.Interpreter{}
	file := os.Args[1]
	code := compiler.Compile(file)

	interpreter.Run(code)
}
