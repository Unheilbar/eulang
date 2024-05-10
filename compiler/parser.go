package compiler

import "fmt"

const (
	eulExprKindStrLit uint8 = iota
	eulExprKindFuncCall

	//... to be continued
)

type eulFuncArg struct {
	value eulExpr
	next  *eulFuncArg
}

type eulFuncCall struct {
	name string
	args *eulFuncArg
}

// TODO eulang another potential use case for union
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

	panic("not implemented")
}
