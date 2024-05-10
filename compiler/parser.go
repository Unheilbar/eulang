package compiler

import (
	"fmt"
	"log"
	"strings"
)

const (
	eulExprKindStrLit uint8 = iota
	eulExprKindFuncCall

	//... to be continued
)

type eulFuncCallArg struct {
	value eulExpr
	next  *eulFuncCallArg
}

type eulFuncCall struct {
	name string
	args *eulFuncCallArg
}

// NOTE eulang another potential use case for union
// NOTE eulang also potentially can be using interface
type eulExprValue struct {
	asStrLit   string
	asFuncCall eulFuncCall
	asU64      uint64
	//... to be continued
}

type eulExpr struct {
	kind  uint8
	value eulExprValue
}

type eulStatement struct {
	expr eulExpr
	kind uint8
}

type eulBlock struct {
	statement eulStatement
	next      *eulBlock
}

type eulFuncDef struct {
	name string
	body *eulBlock
}

func parseFuncDef(lex *lexer) {
	//var module can be used in the future

	lex.expectKeyword("func")
	funcName := lex.expectToken(eulTokenKindName)
	lex.expectToken(eulTokenKindOpenParen)
	lex.expectToken(eulTokenKindCloseParen)
	fmt.Println("name", funcName)
	parseCurlyEulBlock(lex)
}

func parseCurlyEulBlock(lex *lexer) *eulBlock {
	lex.expectToken(eulTokenKindOpenCurly)
	var t = &token{}

	var begin = &eulBlock{}
	var end = &eulBlock{}

	for lex.peek(t) && t.kind != eulTokenKindCloseCurly {
		node := &eulBlock{}
		node.statement.expr = parseEulExpr(lex)
		if end != nil {
			end = node
			end.next = node
		} else {
			if begin != nil {
				panic("exp not nil begin")
			}
			begin = node
			end = node
		}

		lex.expectToken(eulTokenKindSemicolon)
	}

	lex.expectToken(eulTokenKindCloseCurly)

	return begin
}

func parseEulExpr(lex *lexer) eulExpr {
	var t token

	if !lex.peek(&t) {
		log.Fatalf("%s:%d:%d expected expression but got EOF", lex.filepath, lex.row, lex.lineStart)
	}

	var expr eulExpr

	switch t.kind {
	case eulTokenKindName:
		funcall := parseFuncCall(lex)
		expr.kind = eulExprKindFuncCall
		expr.value = eulExprValue{
			asFuncCall: funcall,
		}
	case eulTokenKindLitStr:
		strLit := parseStrLit(lex)
		expr.kind = eulExprKindStrLit
		expr.value = eulExprValue{
			asStrLit: strLit,
		}

	// TODO cases of other token kinds
	default:
		log.Fatalf("%s:%d:%d no expression starts with %s", lex.filepath, lex.row, lex.lineStart, t.view)

	}

	return eulExpr{}
}

func parseFuncCall(lex *lexer) eulFuncCall {
	var funcCall eulFuncCall

	funcCall.name = lex.expectToken(eulTokenKindName).view

	funcCall.args = parseFuncCallArgs(lex)

	return funcCall
}

func parseFuncCallArgs(lex *lexer) *eulFuncCallArg {
	var res = &eulFuncCallArg{}

	lex.expectToken(eulTokenKindOpenParen)
	// TODO parse fun Call currently supports only 1 argument
	res.value = parseEulExpr(lex)
	lex.expectToken(eulTokenKindCloseParen)
	return res
}

func parseStrLit(lex *lexer) string {
	t := lex.expectToken(eulTokenKindLitStr)

	return strings.Trim(t.view, "\"")
}
