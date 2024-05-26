package compiler

import (
	"github.com/Unheilbar/eulang/eulvm"
)

func CompileFromSource(eulang *eulang, filename string) eulvm.Program {
	lex := NewLexerFromFile(filename)
	module := parseEulModule(lex)
	easm := NewEasm()

	eulang.prepareVarStack(easm, eulvm.StackCapacity)

	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.CALLDATA,
	})

	eulang.pushNewScope(nil)
	eulang.compileModuleIntoEasm(easm, module)
	eulang.popScope()

	return easm.GetProgram()
}
