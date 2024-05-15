package compiler

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strconv"
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
	//{eulTokenKindName, "func"}, // all keywords will be handled on the parser level
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

var tokenKindNames = map[uint8]string{
	eulTokenKindName:       "name",
	eulTokenKindNumber:     "number",
	eulTokenKindOpenParen:  "(",
	eulTokenKindCloseParen: ")",
	eulTokenKindOpenCurly:  "{",
	eulTokenKindCloseCurly: "}",
	eulTokenKindSemicolon:  ";",
	eulTokenKindColon:      ":",
	eulTokenKindComma:      ",",
	eulTokenKindEq:         "=",
	eulTokenKindEqEq:       "==",
	eulTokenKindPlus:       "+",
	eulTokenKindMinus:      "-",
	eulTokenKindMult:       "*",
	eulTokenKindLt:         "<",
	eulTokenKindGe:         ">=",
	eulTokenKindNe:         "!=",
	eulTokenKindAnd:        "&&",
	eulTokenKindOr:         "||",
	eulTokenKindDotDot:     "..",
	eulTokenKindLitStr:     "string literal",
}

type peekBuffer struct {
	begin  int
	count  int
	tokens []token
}

// lexer turns native code into stream of tokens(lexemes).
// token or lexem is a reprentation of each item in code at simple level
type lexer struct {
	content []string

	current string

	row       int
	lineStart int //to keep current position on line

	filepath string

	peekBuffer *peekBuffer
}

func NewLexer(content []string, filepath string) *lexer {
	return &lexer{
		content:  content,
		filepath: filepath,
		peekBuffer: &peekBuffer{
			tokens: make([]token, 0),
		},
	}
}

func NewLexerFromFile(filename string) *lexer {
	fi, err := os.Open(filename)
	if err != nil {
		log.Fatalf("err create lexer open file %s err %s", filename, err)
	}

	defer fi.Close()

	scanner := bufio.NewScanner(fi)
	var content []string

	// FIXME doesn't work without it
	content = append(content, " ")
	re, _ := regexp.Compile("//.*")
	for scanner.Scan() {
		line := strings.ReplaceAll(scanner.Text(), "	", " ")
		line = string(re.ReplaceAll([]byte(line), []byte("")))
		content = append(content, line)
	}

	return NewLexer(content, filename)
}

// NOTE euler this function is for using from the inside of lexer only! Don't tru to call it from parser it won't end good for you
func (lex *lexer) tokenByPassPeekBuffer(t *token) bool {
	// Extract next token
	// TODO doesn't trim tabs
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

	// Name token and Number tokens
	{
		nameToken := chopUntil(lex.current, isName)
		if len(nameToken) != 0 && strings.HasPrefix(lex.current, nameToken) {
			*t = lex.chopToken(eulTokenKindName, len(nameToken))
			if isNumber(t.view) {
				t.kind = eulTokenKindNumber
			}
			return true
		}
	}

	// String literal
	{
		//NOTE Euler lexer doesn't support new lines for string literals
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

func (lex *lexer) expectToken(expKind uint8) token {
	var t token

	if !lex.next(&t) {
		log.Fatalf("%s expected token %s but reached EOF", lex.filepath, tokenKindNames[expKind])
	}

	if t.kind != expKind {
		log.Fatalf("%s:%d:%d expected token kind %s but got %s", lex.filepath, t.loc.row, t.loc.col, tokenKindNames[expKind], tokenKindNames[t.kind])
	}

	return t
}

func (lex *lexer) expectKeyword(keyword string) token {
	token := lex.expectToken(eulTokenKindName)
	if token.view != keyword {
		log.Fatalf("%s:%d:%d expected keyword %s but got %s", lex.filepath, token.loc.row, token.loc.col, keyword, token.view)
	}

	return token
}

// next iterates through tokens in lexer content
func (lex *lexer) next(t *token) bool {
	if !lex.peek(t, 0) {
		return false
	}

	lex.peekBuffer.dq()
	return true
}

// peek gets next token in tokenizer content without changing current state of the lexer
func (lex *lexer) peek(t *token, offset int) bool {
	lex.fillPeekBuffer()

	if offset < lex.peekBuffer.count {
		*t = lex.peekBuffer.get(offset)
		return true
	}
	return false
}

func (lex *lexer) fillPeekBuffer() {
	var t token
	for lex.tokenByPassPeekBuffer(&t) {
		lex.peekBuffer.nq(t)
	}
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
		panic("invalid chop token call, token size bigger than current line")
	}

	var t token
	t.kind = kind
	t.view = strings.Trim(lex.current[:size], "\"")
	t.loc = eulLoc{
		row:      lex.row,
		col:      lex.lineStart,
		filepath: lex.filepath,
	}
	if t.kind == eulTokenKindLitStr {
		//TODO eulang dirty hack
		t.view = strings.ReplaceAll(t.view, "\\n", "\n")
	}
	lex.current = lex.current[size:]
	lex.lineStart += size

	return t
}

func (pb *peekBuffer) nq(t token) {
	pb.tokens = append(pb.tokens, t)
	pb.count += 1
}

func (pb *peekBuffer) dq() token {
	internalIdx := pb.begin + pb.count - 1

	if internalIdx > len(pb.tokens)-1 || len(pb.tokens) == 0 {
		panic("invalid dq call")
	}

	res := pb.tokens[internalIdx]
	pb.begin += 1
	pb.count -= 1

	return res

}

func (pb *peekBuffer) get(idx int) token {
	interalIdx := pb.begin + idx
	if interalIdx >= len(pb.tokens) {
		panic("wrong call pb get")
	}
	return pb.tokens[interalIdx]
}

func chopUntil(s string, until func(r rune) bool) string {
	for i, r := range s {
		if !until(r) {
			return s[:i]
		}
	}
	return s
}

func isNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

func isName(r rune) bool {
	return ('a' <= r && r <= 'z') ||
		('A' <= r && r <= 'Z') ||
		('0' <= r && r <= '9') ||
		'_' == r

}

func isNotQuoteMark(r rune) bool {
	return r != '"'
}
