package compiler

import (
	"log"
	"testing"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/Unheilbar/eulang/utils"
)

func Test_compileFuncCallIntoEasm(t *testing.T) {
	code := []string{
		"    ",
		`   func main(){`,
		`write("hello");`,
		`write("hello");`,
		"}",
	}

	lex := NewLexer(code, "testfile")
	funcDef := ParseFuncDef(lex)

	elang := NewEulang()
	easm := NewEasm()
	elang.compileFuncCallIntoEasm(easm, funcDef)
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.STOP,
	})
	//easm.program.Dump()
	//easm.memory.Dump()
	easm.dumpProgramToFile("test.gob")

	program, err := utils.LoadProgramFromFile("test.gob")
	if err != nil {
		panic(err)
	}

	e := eulvm.New(program)
	err = e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
