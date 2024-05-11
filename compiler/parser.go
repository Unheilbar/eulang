package compiler

import (
	"log"
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
type eulExprAs struct {
	strLit   string
	funcCall eulFuncCall
	boolean  bool
	//... to be continued
}

type eulIf struct {
	loc       eulLoc
	condition eulExpr
	ethen     *eulBlock
	elze      *eulBlock //because else is busy by golang
}

type eulExprKind uint8

const (
	eulExprKindStrLit eulExprKind = iota
	eulExprKindBoolLit
	eulExprKindFuncCall

	//... to be continued
)

type eulExpr struct {
	kind eulExprKind
	as   eulExprAs
}

type eulStmtKind uint8

const (
	eulStmtKindExpr eulStmtKind = iota
	eulStmtKindIf               = iota
)

type eulStatementAs struct {
	expr eulExpr
	eif  eulIf
}

type eulStatement struct {
	as   eulStatementAs
	kind eulStmtKind
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
	f.body = *parseCurlyEulBlock(lex)
	return f
}

func parseEulIf(lex *lexer) eulIf {
	var eif eulIf
	t := lex.expectKeyword("if")
	eif.loc = t.loc

	lex.expectToken(eulTokenKindOpenParen)
	eif.condition = parseEulExpr(lex)
	lex.expectToken(eulTokenKindCloseParen)
	eif.ethen = parseCurlyEulBlock(lex)

	//TODO eulang if doesnt support else

	return eif
}

func parseEulStmt(lex *lexer) eulStatement {

	var t token
	if !lex.peek(&t) {
		log.Fatalf("%s:%d:%d expected statement but got EOF", lex.filepath, lex.row, lex.lineStart)
	}

	if t.kind == eulTokenKindName && t.view == "if" {
		var stmt eulStatement
		stmt.kind = eulStmtKindIf
		stmt.as.eif = parseEulIf(lex)
		return stmt
	}

	var stmt eulStatement
	stmt.kind = eulStmtKindExpr
	stmt.as.expr = parseEulExpr(lex)
	lex.expectToken(eulTokenKindSemicolon)

	return stmt
}

func parseCurlyEulBlock(lex *lexer) *eulBlock {
	lex.expectToken(eulTokenKindOpenCurly)
	var t = &token{}
	var result eulBlock

	for lex.peek(t) && t.kind != eulTokenKindCloseCurly {
		result.statements = append(result.statements, parseEulStmt(lex))

	}

	lex.expectToken(eulTokenKindCloseCurly)

	return &result
}

func parseEulExpr(lex *lexer) eulExpr {
	var t token

	if !lex.peek(&t) {
		log.Fatalf("%s:%d:%d expected expression but got EOF", lex.filepath, lex.row, lex.lineStart)
	}

	var expr eulExpr

	switch t.kind {
	case eulTokenKindName:
		if t.view == "true" {
			lex.next(&t)
			expr.as.boolean = true
			expr.kind = eulExprKindBoolLit
		} else if t.view == "false" {
			lex.next(&t)
			expr.as.boolean = false
			expr.kind = eulExprKindBoolLit
		} else if t.view == "false" {
		} else {
			funcall := parseFuncCall(lex)
			expr.kind = eulExprKindFuncCall
			expr.as = eulExprAs{
				funcCall: funcall,
			}
		}
	case eulTokenKindLitStr:
		strLit := parseStrLit(lex)
		expr.kind = eulExprKindStrLit
		expr.as = eulExprAs{
			strLit: strLit,
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
