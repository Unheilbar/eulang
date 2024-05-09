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

type hardcodedToken struct {
	kind uint8
	text string
}

var hardcodedTokens = []hardcodedToken{
	{eulTokenKindDotDot, ".."},
	{eulTokenKindName, "func"},
	{eulTokenKindEqEq, "=="},
	{eulTokenKindOr, "||"},
	{eulTokenKindAnd, "&&"},
	{eulTokenKindNe, "!="},
	{eulTokenKindGe, ">="},
	{eulTokenKindOpenParen, "("},
	{eulTokenKindCloseParen, ")"},
	{eulTokenKindOpenCurly, "{"},
	{eulTokenKindCloseCurly, "}"},
	{eulTokenKindSemicolon, ";"},
	{eulTokenKindColon, ":"},
	{eulTokenKindComma, ","},
	{eulTokenKindEq, "="},
	{eulTokenKindPlus, "+"},
	{eulTokenKindMinus, "-"},
	{eulTokenKindMult, "*"},
	{eulTokenKindLt, "<"},
}

// lexer turns native code into stream of tokens(lexemes).
// token or lexem is a reprentation of each item in code at simple level
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

	for len(lex.current) == 0 && len(lex.content) > 1 {
		lex.nextLine()
	}

	if len(lex.current) == 0 {
		return false
	}

	// Firstly try parsing hardcoded tokens
	{
		for _, htok := range hardcodedTokens {
			if strings.HasPrefix(lex.current, htok.text) {
				*t = lex.chopToken(htok.kind, len(htok.text))
				return true
			}
		}
	}

	// Name token
	{
		nameToken := chopUntil(lex.current, isName)
		if len(nameToken) != 0 && strings.HasPrefix(lex.current, nameToken) {
			*t = lex.chopToken(eulTokenKindName, len(nameToken))
			return true
		}
	}

	// String literal
	{
		//NOTE Euler lexer doesn't support new lines
		if len(lex.current) > 2 && lex.current[0] == '"' {
			strToken := chopUntil(lex.current[1:], isNotQuoteMark)
			if len(strToken) >= len(lex.current[1:]) {
				log.Fatalf("%s:%d:%d unclosed string literal '%s'", lex.filepath, lex.row, lex.lineStart, strToken)
			}
			strToken = "\"" + strToken + "\""
			*t = lex.chopToken(eulTokenKindLitStr, len(strToken))
			return true
		}
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

func chopUntil(s string, until func(r rune) bool) string {
	for i, r := range s {
		if !until(r) {
			return s[:i]
		}
	}
	return s
}

func isName(r rune) bool {
	return ('a' <= r && r <= 'z') ||
		('A' <= r && r <= 'Z') ||
		('0' < r && r < '9') ||
		'_' == r

}

func isNotQuoteMark(r rune) bool {
	return r != '"'
}
