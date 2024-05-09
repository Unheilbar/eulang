package compiler

import (
	"log"
	"strings"
)

const (
	eulTokenKindName uint8 = iota
	eulTokenKindNumber
	eulTokenKindOpenParen
	eulTokenKindCloseParen
	eulTokenKindOpenCurly
	eulTokenKindCloseCurly
	eulTokenKindSemicolon
	eulTokenKindColon
	eulTokenKindComma
	eulTokenKindEq
	eulTokenKindLitStr
	eulTokenKindPlus
	eulTokenKindMinus
	eulTokenKindMult
	eulTokenKindLt
	eulTokenKindGe
	eulTokenKindNe
	eulTokenKindAnd
	eulTokenKindOr
	eulTokenKindEqEq
	eulTokenKindDotDot
	//add here

	eulTokenKindKinds
)

type eulLoc struct {
	row int
	col int

	filepath string
}

type token struct {
	kind uint8
	view string

	loc eulLoc
}

type lexer struct {
	content []string

	current string

	row       int
	lineStart int //to keep current position on line

	filepath string
}

func NewLexer(content []string, filepath string) *lexer {
	return &lexer{
		content:  content,
		filepath: filepath,
	}
}

func (lex *lexer) next(t *token) bool {
	lex.current = strings.TrimLeft(lex.current, " ")

	for len(lex.current) == 0 && len(lex.content) != 0 {
		lex.nextLine()
	}

	if len(lex.current) == 0 {
		return false
	}

	//
	tokenStr := "func"
	if strings.HasPrefix(lex.current, tokenStr) {
		*t = lex.chopToken(eulTokenKindName, len(tokenStr))
		return true
	}
	tokenStr = "("
	if strings.HasPrefix(lex.current, tokenStr) {
		*t = lex.chopToken(eulTokenKindName, len(tokenStr))
		return true
	}

	log.Fatalf("%s:%d:%d Unkown token start with '%s'", lex.filepath, lex.row, lex.lineStart, lex.current[:1])

	return false
}

func (lex *lexer) nextLine() {
	lex.content = lex.content[1:]
	lex.row++
	line := lex.content[0]
	lex.lineStart = 0
	lex.current = strings.TrimLeft(line, " ")
	lex.lineStart += len(line) - len(lex.current) //line start includes trimmmed spaces
}

func (lex *lexer) chopToken(kind uint8, size int) token {
	if size > len(lex.current) {
		panic("invalid chop token call")
	}
	var t token
	t.kind = kind
	t.view = lex.current[:size]
	t.loc = eulLoc{
		row:      lex.row,
		col:      lex.lineStart,
		filepath: lex.filepath,
	}

	lex.current = lex.current[size:]
	lex.lineStart += size
	return t
}
