package compiler

import "testing"

func Test_compileFuncCallIntoEasm(t *testing.T) {
	code := []string{
		"    ",
		`   func main(){`,
		`write("hello");`,
		`write("hello");`,
		"}",
	}

	lex := NewLexer(code, "testfile")
	funcDef := parseFuncDef(lex)

	elang := newEulang()
	easm := newEasm()
	elang.compileFuncCallIntoEasm(easm, funcDef)
	easm.program.Dump()
	easm.memory.Dump()
}
