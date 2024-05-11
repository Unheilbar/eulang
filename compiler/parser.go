package compiler

import (
	"log"
)

const (
	eulExprKindStrLit uint8 = iota
	eulExprKindFuncCall

	//... to be continued
)

type eulFuncCallArg struct {
	value eulExpr
	//TODO probably makes sense to represent it as linked list so we can iterate through arguments
}

type eulFuncCall struct {
	name string
	args []eulFuncCallArg

	loc eulLoc
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
	statements []eulStatement
}

type eulFuncDef struct {
	name string
	body eulBlock
}

func ParseFuncDef(lex *lexer) eulFuncDef {
	//var module can be used in the future
	var f eulFuncDef
	lex.expectKeyword("func")
	f.name = lex.expectToken(eulTokenKindName).view
	lex.expectToken(eulTokenKindOpenParen)
	//TODO eulang func def do not support args yet
	lex.expectToken(eulTokenKindCloseParen)
	f.body = parseCurlyEulBlock(lex)
	return f
}

func parseCurlyEulBlock(lex *lexer) eulBlock {
	lex.expectToken(eulTokenKindOpenCurly)
	var t = &token{}
	var result eulBlock

	for lex.peek(t) && t.kind != eulTokenKindCloseCurly {
		result.statements = append(result.statements, eulStatement{
			expr: parseEulExpr(lex),
		})

		lex.expectToken(eulTokenKindSemicolon)
	}

	lex.expectToken(eulTokenKindCloseCurly)

	return result
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

	return expr
}

func parseFuncCall(lex *lexer) eulFuncCall {
	var funcCall eulFuncCall

	t := lex.expectToken(eulTokenKindName)
	funcCall.name = t.view
	funcCall.loc = t.loc
	funcCall.args = parseFuncCallArgs(lex)

	return funcCall
}

func parseFuncCallArgs(lex *lexer) []eulFuncCallArg {
	var res []eulFuncCallArg

	lex.expectToken(eulTokenKindOpenParen)
	// TODO parse fun Call currently supports only 1 argument fix later
	arg := parseEulExpr(lex)
	res = append(res, eulFuncCallArg{arg})
	lex.expectToken(eulTokenKindCloseParen)
	return res
}

func parseStrLit(lex *lexer) string {
	return lex.expectToken(eulTokenKindLitStr).view
}
