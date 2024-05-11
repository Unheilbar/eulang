package compiler

import (
	"fmt"
	"testing"
)

func Test_parseFuncDef(t *testing.T) {
	code := []string{
		"    ",
		`   func main(){`,
		`write("hello");`,
		`write("hello");`,
		`  read("zadravstvuyte");`,
		"}",
	}

	lex := NewLexer(code, "testfile")
	funcDef := ParseFuncDef(lex)

	fmt.Println("funcName", funcDef.name)

	for i, stmt := range funcDef.body.statements {
		fmt.Println(i, "statement: ", stmt)
	}
}
