package compiler

import (
	"fmt"
	"testing"
)

func Test_LexNext(t *testing.T) {
	var code = []string{
		"    ",
		"  ",
		"    func main () {",
		`       event("hello")`,
		"}",
	}

	lex := NewLexer(code, "testfile")
	var tok = &token{}
	//lex.next(tok)
	lim := 10
	for lex.next(tok) && lim > 0 {
		lim--
		fmt.Println(tok.view)
	}
	//expTok := &token{
	//	kind: eulTokenKindName,
	//	view: "func",
	//	loc: eulLoc{
	//		row:      2,
	//		col:      4,
	//		filepath: "testfile",
	//	},
	//}

	//assert.Equal(t, expTok, tok)
	lex.next(tok)
}
