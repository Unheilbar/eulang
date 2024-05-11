package compiler

import (
	"fmt"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/holiman/uint256"
)

type compiledFuncs struct{}

type eulang struct {
	funcs []compiledFuncs
}

func newEulang() *eulang {
	return &eulang{}
}

func (e *eulang) compileFuncCallIntoEasm(easm *easm, fd eulFuncDef) {
	for _, statement := range fd.body.statements {
		e.compileStatementIntoEasm(easm, statement)
	}
}

func (e *eulang) compileStatementIntoEasm(easm *easm, stmt eulStatement) {
	e.compileExprIntoEasm(easm, stmt.expr)
}

func (e *eulang) compileExprIntoEasm(easm *easm, expr eulExpr) {
	switch expr.kind {
	case eulExprKindFuncCall:
		// TODO temporary solution hard code just one function
		if expr.value.asFuncCall.name != "write" {
			fmt.Println("unexpected name", expr.value.asFuncCall.name)
			panic("implement other functions later")
		}
		e.compileExprIntoEasm(easm, expr.value.asFuncCall.args[0].value)
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.OpCode(eulvm.NATIVE),
			Operand: *uint256.NewInt(eulvm.NativeWrite),
		})
	case eulExprKindStrLit:
		var addr eulvm.Word = easm.pushStringToMemory(expr.value.asStrLit)

		pushStrAddrInst := eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: addr,
		}

		easm.pushInstruction(pushStrAddrInst)

		pushStrSizeInst := eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *uint256.NewInt(uint64(len(expr.value.asStrLit))),
		}

		easm.pushInstruction(pushStrSizeInst)
	default:
		panic("unsupported expression kind")
	}
}
