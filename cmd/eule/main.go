package main

import (
	"log"
	"os"

	"github.com/Unheilbar/eulang/compiler"
	"github.com/Unheilbar/eulang/eulvm"
)

func main() {
	file := os.Args[1]

	lex := compiler.NewLexerFromFile(file)
	funcDef := compiler.ParseFuncDef(lex)
	eulang := compiler.NewEulang()
	easm := compiler.NewEasm()
	eulang.CompileFuncCallIntoEasm(easm, funcDef)

	// TODO later eulang will push stop instruction
	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.STOP,
	})

	prog := easm.GetProgram()

	e := eulvm.New(prog)
	err := e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
