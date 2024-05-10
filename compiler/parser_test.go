package compiler

import "testing"

func Test_parseFuncDef(t *testing.T) {
	code := []string{
		"    ",
		`   func main(){`,
		`write("hello");`,
		"}",
	}

	lex := NewLexer(code, "testfile")
	parseFuncDef(lex)
}
