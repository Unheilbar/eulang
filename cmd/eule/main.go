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
	//TODO later elang will export method to compile from file
	eulang.CompileFuncCallIntoEasm(easm, funcDef)

	// TODO later eulang will push stop instruction
	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.STOP,
	})

	prog := easm.GetProgram()
	//for idx, inst := range prog.Instrutions {
	//	fmt.Println(idx, eulvm.OpCodes[inst.OpCode], inst.Operand.Uint64())

	//}
	e := eulvm.New(prog)
	err := e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
