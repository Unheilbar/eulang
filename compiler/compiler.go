package compiler

import "github.com/Unheilbar/eulang/eulvm"

func CompileFromSource(filename string) eulvm.Program {
	lex := NewLexerFromFile(filename)
	module := parseEulModule(lex)
	easm := NewEasm()
	eulang := NewEulang()

	eulang.compileModuleIntoEasm(easm, module)

	//TODO later add it somewhere else
	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.STOP,
	})

	return easm.GetProgram()
}
