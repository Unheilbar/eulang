package compiler

import (
	"log"
	"strconv"
)

var binaryOpTokens = map[eulBinaryOpKind]uint8{
	binaryOpLess: eulTokenKindLt,
	binaryOpPlus: eulTokenKindPlus,
}

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

type eulWhile struct {
	loc       eulLoc
	condition eulExpr
	body      eulBlock
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
	eulExprKindIntLit
	eulExprKindVarRead
	eulExprKindBinaryOp
	//... to be continued
)

type eulExprAs struct {
	strLit   string
	funcCall eulFuncCall
	boolean  bool
	intLit   int64
	varRead  varRead
	binaryOp *binaryOp
	//... to be continued
}

type eulExpr struct {
	loc  eulLoc
	kind eulExprKind
	as   eulExprAs
}

type eulStmtKind uint8

const (
	eulStmtKindExpr eulStmtKind = iota
	eulStmtKindIf
	eulStmtKindVarAssign
	eulStmtKindWhile
	eulStmtKindVarDef
)

type eulStatementAs struct {
	expr      eulExpr
	eif       eulIf
	varAssign eulVarAssign
	while     eulWhile
	vardef    eulVarDef
}

type eulStatement struct {
	as   eulStatementAs
	kind eulStmtKind
}

type eulBlock struct {
	statements []eulStatement
}

type eulFuncModifier uint8

const (
	eulModifierKindInternal eulFuncModifier = iota
	eulModifierKindExternal
	//.. to be continued
)

type eulFuncDef struct {
	name     string
	modifier eulFuncModifier

	body   eulBlock
	loc    eulLoc
	params []eulFuncParam
}

type eulType uint8

const (
	eulTypei64 eulType = iota
	eulTypeVoid
	eulTypeBytes32
	// to be continued
	//...
	eulTypeCount
)

var eulTypesView = map[string]eulType{
	"i64":     eulTypei64,
	"void":    eulTypeVoid,
	"bytes32": eulTypeBytes32,
}

var eulTypes = map[eulType]string{
	eulTypei64:     "i64",
	eulTypeVoid:    "void",
	eulTypeBytes32: "bytes32",
}

type eulVarDef struct {
	name  string
	etype eulType
	loc   eulLoc

	init    eulExpr
	hasInit bool
}

type eulVarAssign struct {
	name  string
	value eulExpr
	loc   eulLoc
}

type varRead struct {
	name string
	loc  eulLoc
}

type eulFuncParam struct {
	name  string
	typee eulType
	loc   eulLoc
}

type eulBinaryOpKind uint8

const (
	binaryOpPlus eulBinaryOpKind = iota
	binaryOpLess

	countBinaryOpKinds //keep it last
)

type binaryOp struct {
	loc  eulLoc
	kind eulBinaryOpKind
	lhs  eulExpr
	rhs  eulExpr
	// TODO
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
	for lex.peek(&t, 0) {
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

func parseVarDef(lex *lexer) eulVarDef {
	// eulang:
	// var i: i32;
	//
	var vdef eulVarDef

	lex.expectKeyword("var")
	t := lex.expectToken(eulTokenKindName)
	vdef.loc = t.loc
	vdef.name = t.view
	vdef.etype = parseEulType(lex)

	// Try to parse assignment
	{
		var t token
		if lex.peek(&t, 0) && t.kind == eulTokenKindEq {
			lex.next(&t)
			vdef.hasInit = true
			vdef.init = parseEulExpr(lex)
		}
	}

	lex.expectToken(eulTokenKindSemicolon)

	return vdef
}

func parseFuncDef(lex *lexer) eulFuncDef {
	//var module can be used in the future
	var f eulFuncDef
	lex.expectKeyword("func")
	t := lex.expectToken(eulTokenKindName)
	f.loc = t.loc
	f.name = t.view

	f.params = parseFuncDefParams(lex)

	//If function has modifier then parse it
	{
		var t token
		lex.peek(&t, 0)
		if t.kind == eulTokenKindName {
			t = lex.expectToken(eulTokenKindName)
			switch t.view {
			case "external":
				f.modifier = eulModifierKindExternal
			case "internal":
				f.modifier = eulModifierKindInternal
			default:
				log.Fatalf("%s:%d:%d undefined modifier '%s'", t.loc.filepath, t.loc.row, t.loc.col, t.view)
			}
		} else {
			// NOTE default modifier is internal
			f.modifier = eulModifierKindInternal
		}
	}
	f.body = *parseCurlyEulBlock(lex)
	return f
}

func parseFuncDefParams(lex *lexer) []eulFuncParam {
	var result []eulFuncParam

	lex.expectToken(eulTokenKindOpenParen)

	var t token
	for lex.peek(&t, 0) && t.kind != eulTokenKindCloseParen {
		var p eulFuncParam
		tok := lex.expectToken(eulTokenKindName)
		p.name = tok.view
		p.loc = tok.loc
		p.typee = parseEulType(lex)
		result = append(result, p)
		if lex.peek(&t, 0) && t.kind != eulTokenKindCloseParen {
			lex.expectToken(eulTokenKindComma)
		}
	}
	lex.expectToken(eulTokenKindCloseParen)
	return result
}

func parseEulWhile(lex *lexer) eulWhile {
	var while eulWhile

	//Parse while
	{
		t := lex.expectKeyword("while")
		while.loc = t.loc

		while.condition = parseEulExpr(lex)
		while.body = *parseCurlyEulBlock(lex)
	}

	return while
}

func parseEulIf(lex *lexer) eulIf {
	var eif eulIf
	// Parse then
	{
		t := lex.expectKeyword("if")
		eif.loc = t.loc

		eif.condition = parseEulExpr(lex)

		eif.ethen = parseCurlyEulBlock(lex)
	}

	// Parse else if exists
	{
		var t token
		if lex.peek(&t, 0) && t.kind == eulTokenKindName && t.view == "else" {
			lex.next(&t)
			eif.elze = parseCurlyEulBlock(lex)
		}
	}

	return eif
}

func parseEulStmt(lex *lexer) eulStatement {

	var t token
	if !lex.peek(&t, 0) {
		log.Fatalf("%s:%d:%d expected statement but got EOF", lex.filepath, lex.row, lex.lineStart)
	}

	switch t.kind {
	case eulTokenKindName:
		switch t.view {
		case "if":
			var stmt eulStatement
			stmt.kind = eulStmtKindIf
			stmt.as.eif = parseEulIf(lex)
			return stmt
		case "while":
			var stmt eulStatement
			stmt.kind = eulStmtKindWhile
			stmt.as.while = parseEulWhile(lex)
			return stmt
		case "var":
			var stmt eulStatement
			stmt.kind = eulStmtKindVarDef
			stmt.as.vardef = parseVarDef(lex)
			return stmt
		default:
			var nt token
			if lex.peek(&nt, 1) && nt.kind == eulTokenKindEq {
				var stmt eulStatement
				stmt.kind = eulStmtKindVarAssign

				stmt.as.varAssign = parseVarAssign(lex)
				lex.expectToken(eulTokenKindSemicolon)
				return stmt
			}
		}
	}

	// if it's still there then just parse it as a statement

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

	for lex.peek(t, 0) && t.kind != eulTokenKindCloseCurly {
		result.statements = append(result.statements, parseEulStmt(lex))

	}

	lex.expectToken(eulTokenKindCloseCurly)

	return &result
}

func parseVarAssign(lex *lexer) eulVarAssign {
	var vas eulVarAssign

	vas.name = lex.expectToken(eulTokenKindName).view
	t := lex.expectToken(eulTokenKindEq)
	vas.value = parseEulExpr(lex)
	vas.loc = t.loc
	return vas
}

func parsePrimaryExpr(lex *lexer) eulExpr {
	var t token

	if !lex.peek(&t, 0) {
		log.Fatalf("%s:%d:%d expected expression but got EOF", lex.filepath, lex.row, lex.lineStart)
	}

	var expr eulExpr

	switch t.kind {
	case eulTokenKindName:
		if t.view == "true" {
			lex.next(&t)
			expr.as.boolean = true
			expr.kind = eulExprKindBoolLit
			expr.loc = t.loc
		} else if t.view == "false" {
			lex.next(&t)
			expr.as.boolean = false
			expr.kind = eulExprKindBoolLit
			expr.loc = t.loc
		} else {
			var nextTok token
			if lex.peek(&nextTok, 1) && nextTok.kind == eulTokenKindOpenParen {
				funcall := parseFuncCall(lex)
				expr.loc = t.loc
				expr.kind = eulExprKindFuncCall
				expr.as.funcCall = funcall
			} else {
				expr.kind = eulExprKindVarRead
				expr.loc = t.loc
				expr.as.varRead = parseVarRead(lex)
			}
		}
	case eulTokenKindLitStr:
		strLit := parseStrLit(lex)
		expr.kind = eulExprKindStrLit
		expr.as.strLit = strLit
		expr.loc = t.loc
	case eulTokenKindNumber:
		intLit := parseIntLit(lex)
		expr.kind = eulExprKindIntLit
		expr.as.intLit = intLit
		expr.loc = t.loc
	// TODO cases of other token kinds
	default:
		log.Fatalf("%s:%d:%d no primary expression starts with %s",
			lex.filepath, lex.row, lex.lineStart, t.view)

	}

	return expr
}

func parseEulExpr(lex *lexer) eulExpr {
	lhs := parsePrimaryExpr(lex)

	var t token
	if lex.peek(&t, 0) {
		for kind := eulBinaryOpKind(0); kind < countBinaryOpKinds; kind++ {
			if binaryOpTokens[kind] == t.kind {
				ok := lex.next(&t)
				if !ok {
					panic("your lexer is broken!!1111")
				}
				var binOp binaryOp

				// Initialize binary op
				{
					binOp.loc = t.loc
					binOp.kind = kind
					binOp.lhs = lhs

					binOp.rhs = parseEulExpr(lex)
				}

				var expr eulExpr
				expr.kind = eulExprKindBinaryOp
				expr.as.binaryOp = &binOp
				return expr
			}
		}
	}

	return lhs
}

func parseFuncCall(lex *lexer) eulFuncCall {
	var funcCall eulFuncCall

	t := lex.expectToken(eulTokenKindName)
	funcCall.name = t.view
	funcCall.loc = t.loc
	funcCall.args = parseFuncCallArgs(lex)

	return funcCall
}

func parseVarRead(lex *lexer) varRead {
	var vr varRead
	t := lex.expectToken(eulTokenKindName)
	vr.name = t.view
	vr.loc = t.loc
	return vr
}

func parseFuncCallArgs(lex *lexer) []eulFuncCallArg {
	var res []eulFuncCallArg

	lex.expectToken(eulTokenKindOpenParen)

	var t token
	for lex.peek(&t, 0) && t.kind != eulTokenKindCloseParen {
		res = append(res, eulFuncCallArg{
			value: parseEulExpr(lex),
		})
		if lex.peek(&t, 0) && t.kind != eulTokenKindCloseParen {
			lex.expectToken(eulTokenKindComma)
		}
	}

	lex.expectToken(eulTokenKindCloseParen)

	return res
}

func parseEulType(lex *lexer) eulType {
	lex.expectToken(eulTokenKindColon)

	tok := lex.expectToken(eulTokenKindName)
	for typee, view := range eulTypes {
		if tok.view == view {
			return typee
		}
	}

	log.Fatalf("%s:%d:%d undefined type '%s'",
		tok.loc.filepath, tok.loc.row, tok.loc.col, tok.view)

	return 99
}

func parseStrLit(lex *lexer) string {
	return lex.expectToken(eulTokenKindLitStr).view
}

func parseIntLit(lex *lexer) int64 {
	intLit, err := strconv.Atoi(lex.expectToken(eulTokenKindNumber).view)
	if err != nil {
		panic("parse int lit invalid call")
	}
	return int64(intLit)
}
