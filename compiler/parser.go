package compiler

import (
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
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
	eulExprKindByte32Lit
	eulExprKindAddressLit
	eulExprKindBoolLit
	eulExprKindIntLit
	eulExprKindFuncCall
	eulExprKindVarRead
	eulExprKindBinaryOp
	//... to be continued
)

type eulExprAs struct {
	strLit     string
	funcCall   eulFuncCall
	addressLit common.Address
	boolean    bool
	intLit     int64
	varRead    varRead
	binaryOp   *binaryOp
	bytes32Lit common.Hash
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
	eulStmtKindMapWrite
	eulStmtKindWhile
	eulStmtKindVarDef
)

type eulStatementAs struct {
	expr      eulExpr
	eif       eulIf
	varAssign eulVarAssign
	mapWrite  eulMapWrite
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
	eulTypeBool
	eulTypeBytes32
	eulTypeAddress
	// to be continued
	//...

	eulTypeCount
)

var eulTypesView = map[string]eulType{
	"i64":     eulTypei64,
	"void":    eulTypeVoid,
	"bool":    eulTypeBool,
	"bytes32": eulTypeBytes32,
	"address": eulTypeAddress,
}

var eulTypes = map[eulType]string{
	eulTypei64:     "i64",
	eulTypeVoid:    "void",
	eulTypeBool:    "bool",
	eulTypeBytes32: "bytes32",
	eulTypeAddress: "address",
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

type eulMapWrite struct {
	name  string
	value eulExpr
	key   eulExpr
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
	binaryOpKindPlus eulBinaryOpKind = iota
	binaryOpKindLess
	binaryOpKindGreater
	binaryOpKindMinus
	binaryOpKindEqual
	binaryOpKindNotEqual
	binaryOpKindAnd
	binaryOpKindOr
	binaryOpKindMulti

	countBinaryOpKinds //keep it last
)

type binaryOp struct {
	loc  eulLoc
	kind eulBinaryOpKind
	lhs  eulExpr
	rhs  eulExpr
	// TODO
}

type binaryOpDef struct {
	prec  eulBinaryOpPrecedence
	token eulTokenKind
	kind  eulBinaryOpKind
}

var binaryOpDefs = map[eulBinaryOpKind]binaryOpDef{
	// multi/div
	binaryOpKindMulti: {
		token: eulTokenKindMult,
		prec:  eulBinOpPrecedence3,
		kind:  binaryOpKindMulti,
	},
	// arithmetc precedence 2
	binaryOpKindPlus: {
		token: eulTokenKindPlus,
		prec:  eulBinOpPrecedence2,
		kind:  binaryOpKindPlus,
	},
	binaryOpKindMinus: {
		token: eulTokenKindMinus,
		prec:  eulBinOpPrecedence2,
		kind:  binaryOpKindMinus,
	},
	// comparison precedence 1
	binaryOpKindLess: {
		token: eulTokenKindLt,
		prec:  eulBinOpPrecedence1,
		kind:  binaryOpKindLess,
	},
	binaryOpKindGreater: {
		token: eulTokenKindGt,
		prec:  eulBinOpPrecedence1,
		kind:  binaryOpKindGreater,
	},
	binaryOpKindEqual: {
		token: eulTokenKindEqEq,
		prec:  eulBinOpPrecedence1,
		kind:  binaryOpKindEqual,
	},
	binaryOpKindNotEqual: {
		token: eulTokenKindNe,
		prec:  eulBinOpPrecedence1,
		kind:  binaryOpKindNotEqual,
	},
	// logical precedence 0
	binaryOpKindAnd: {
		token: eulTokenKindAnd,
		prec:  eulBinOpPrecedence0,
		kind:  binaryOpKindAnd,
	},
	binaryOpKindOr: {
		token: eulTokenKindOr,
		prec:  eulBinOpPrecedence0,
		kind:  binaryOpKindOr,
	},
}

type eulBinaryOpPrecedence uint8

const (
	eulBinOpPrecedence0 eulBinaryOpPrecedence = iota
	eulBinOpPrecedence1
	eulBinOpPrecedence2
	eulBinOpPrecedence3

	eulCountBinaryOpPrecedence
)

type eulMapDef struct {
	loc     eulLoc
	name    string
	keyType eulType
	valType eulType
	// TODO later add storage kind for maps
}

type eulTopKind uint8

const (
	eulTopKindFunc = iota
	eulTopKindVar
	eulTopKindMap
)

type eulTopAs struct {
	vdef eulVarDef
	fdef eulFuncDef
	mdef eulMapDef
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
		case "map":
			mdef := parseMapDef(lex)

			top.as.mdef = mdef
			top.kind = eulTopKindMap
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

func parseMapDef(lex *lexer) eulMapDef {
	var mdef eulMapDef
	lex.expectKeyword("map")
	tok := lex.expectToken(eulTokenKindName)
	mdef.loc = tok.loc
	mdef.name = tok.view

	lex.expectToken(eulTokenKindOpenBrack)
	mdef.keyType = parseEulType(lex)
	lex.expectToken(eulTokenKindCloseBrack)
	mdef.valType = parseEulType(lex)

	lex.expectToken(eulTokenKindSemicolon)
	return mdef
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
			} else if nt.kind == eulTokenKindOpenBrack {
				var stmt eulStatement
				stmt.kind = eulStmtKindMapWrite

				stmt.as.mapWrite = parseMapWrite(lex)

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

func parseMapWrite(lex *lexer) eulMapWrite {
	var mwrite eulMapWrite

	mwrite.name = lex.expectToken(eulTokenKindName).view
	lex.expectToken(eulTokenKindOpenBrack)
	mwrite.key = parseEulExpr(lex)
	lex.expectToken(eulTokenKindCloseBrack)
	t := lex.expectToken(eulTokenKindEq)

	mwrite.loc = t.loc
	mwrite.value = parseEulExpr(lex)

	return mwrite
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
	case eulTokenKindOpenParen:
		lex.next(&t)
		expr = parseEulExpr(lex)
		lex.expectToken(eulTokenKindCloseParen)
	default:
		log.Fatalf("%s:%d:%d no primary expression starts with %s",
			lex.filepath, lex.row, lex.lineStart, t.view)

	}

	return expr
}

func parseEulExprWithPrecedence(lex *lexer, prec eulBinaryOpPrecedence) eulExpr {
	if prec >= eulCountBinaryOpPrecedence {
		return parsePrimaryExpr(lex)
	}

	lhs := parseEulExprWithPrecedence(lex, prec+1)
	var t token
	var def binaryOpDef
	for lex.peek(&t, 0) && binDefByToken(t, &def) && def.prec == prec {
		ok := lex.next(&t)
		if !ok {
			panic("broken parser")
		}
		var expr eulExpr
		expr.loc = t.loc
		expr.kind = eulExprKindBinaryOp
		binOp := &binaryOp{}

		{
			binOp.loc = t.loc
			binOp.kind = def.kind
			binOp.lhs = lhs
			binOp.rhs = parseEulExprWithPrecedence(lex, prec+1)
		}

		expr.as.binaryOp = binOp

		lhs = expr
	}

	return lhs
}

func parseEulExpr(lex *lexer) eulExpr {
	return parseEulExprWithPrecedence(lex, 0)
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

func binDefByToken(t token, def *binaryOpDef) bool {
	for _, binDef := range binaryOpDefs {
		if t.kind == binDef.token {
			*def = binDef
			return true
		}
	}
	return false
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
