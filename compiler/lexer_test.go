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
		`       event("hello");`,
		"}",
	}

	lex := NewLexer(code, "testfile")
	var tok = &token{}
	//lex.next(tok)
	lim := 12
	lex.peek(tok, 1)

	for lex.next(tok) && lim > 0 {
		lim--
		fmt.Println("view", tok.view, "\tkind", tokenKindNames[tok.kind])
	}

	lex.next(tok)
}
