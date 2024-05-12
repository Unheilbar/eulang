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
	eulStmtKindIf
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

type eulType uint8

const (
	eulTypei64 eulType = iota
)

type eulVarDef struct {
	name  string
	etype eulType
	loc   eulLoc
}

type eulTopKind uint8

const (
	eulTopKindFunc = iota
	eulTopKindVar
)

type eulTopAs struct {
	vdef eulVarDef
	fdef eulFuncDef
}

type eulTop struct {
	kind eulTopKind
	as   eulTopAs
}

type eulModule struct {
	tops []eulTop
}

func parseEulModule(lex *lexer) eulModule {
	var mod eulModule
	var t token
	for lex.peek(&t) {
		if t.kind != eulTokenKindName {
			log.Fatalf("%s:%d:%d expected var or func definition but got %s", t.loc.filepath, t.loc.row, t.loc.col, tokenKindNames[t.kind])
		}

		var top eulTop
		switch t.view {
		case "func":
			fdef := parseFuncDef(lex)

			top.as.fdef = fdef
			top.kind = eulTopKindFunc

		case "var":
			vdef := parseVarDef(lex)

			top.as.vdef = vdef
			top.kind = eulTopKindVar
		default:
			log.Fatalf("%s:%d:%d expected module definitions but got keyword %s", t.loc.filepath, t.loc.row, t.loc.col, t.view)
		}
		mod.tops = append(mod.tops, top)
	}

	return mod
}

func parseEulTop(lex *lexer) eulTop {
	return eulTop{}
}

func parseVarDef(lex *lexer) eulVarDef {
	// eulang:
	// var i: i32;
	//
	var vdef eulVarDef

	lex.expectKeyword("var")
	t := lex.expectToken(eulTokenKindName)
	vdef.loc = t.loc
	vdef.name = t.view
	lex.expectToken(eulTokenKindColon)
	vdef.etype = parseEulType(lex)
	lex.expectToken(eulTokenKindSemicolon)

	return vdef
}

func parseFuncDef(lex *lexer) eulFuncDef {
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
	// Parse then
	{
		t := lex.expectKeyword("if")
		eif.loc = t.loc

		lex.expectToken(eulTokenKindOpenParen)
		eif.condition = parseEulExpr(lex)
		lex.expectToken(eulTokenKindCloseParen)
		eif.ethen = parseCurlyEulBlock(lex)
	}

	// Parse else if exists
	{
		var t token
		if lex.peek(&t) && t.kind == eulTokenKindName && t.view == "else" {
			lex.next(&t)
			eif.elze = parseCurlyEulBlock(lex)
		}
	}

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
		} else {
			funcall := parseFuncCall(lex)
			expr.kind = eulExprKindFuncCall
			expr.as.funcCall = funcall
		}
	case eulTokenKindLitStr:
		strLit := parseStrLit(lex)
		expr.kind = eulExprKindStrLit
		expr.as.strLit = strLit

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

func parseEulType(lex *lexer) eulType {
	//TODO Euler implement other types here later
	lex.expectKeyword("i64")

	return eulTypei64
}

func parseStrLit(lex *lexer) string {
	return lex.expectToken(eulTokenKindLitStr).view
}
