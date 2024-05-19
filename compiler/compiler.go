package compiler

import "github.com/Unheilbar/eulang/eulvm"

func CompileFromSource(eulang *eulang, filename string) eulvm.Program {
	lex := NewLexerFromFile(filename)
	module := parseEulModule(lex)
	easm := NewEasm()

	eulang.prepareVarStack(easm, eulvm.StackCapacity)

	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.CALLDATA,
	})

	//TODO later add it somewhere else
	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.STOP,
	})

	eulang.pushNewScope()
	eulang.compileModuleIntoEasm(easm, module)
	eulang.popScope()

	return easm.GetProgram()
}
