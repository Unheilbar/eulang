package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LexNext(t *testing.T) {
	var code = []string{
		"    ",
		"  ",
		"    func () {",
		"	write(1);",
		"}",
	}

	lex := NewLexer(code, "testfile")
	var tok = &token{}
	lex.next(tok)

	expTok := &token{
		kind: eulTokenKindName,
		view: "func",
		loc: eulLoc{
			row:      2,
			col:      4,
			filepath: "testfile",
		},
	}
	assert.Equal(t, expTok, tok)
	lex.next(tok)
}
