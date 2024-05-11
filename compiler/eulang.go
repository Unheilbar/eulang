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

func NewEulang() *eulang {
	return &eulang{}
}

func (e *eulang) compileFuncCallIntoEasm(easm *easm, fd eulFuncDef) {
	for _, statement := range fd.body.statements {
		e.compileStatementIntoEasm(easm, statement)
	}
}

// TODO temp later should be public
func (e *eulang) CompileFuncCallIntoEasm(easm *easm, fd eulFuncDef) {
	for _, statement := range fd.body.statements {
		e.compileStatementIntoEasm(easm, statement)
	}
}

func (e *eulang) compileStatementIntoEasm(easm *easm, stmt eulStatement) {
	switch stmt.kind {
	case eulStmtKindExpr:
		e.compileExprIntoEasm(easm, stmt.as.expr)
	case eulStmtKindIf:
		e.compileIfIntoEasm(easm, stmt.as.eif)
	default:
		panic(fmt.Sprintf("stmt kind doesn't exist kind %d", stmt.kind))
	}
}

func (e *eulang) compileIfIntoEasm(easm *easm, eif eulIf) {
	if eif.elze != nil {
		panic("else for expressions is not supported by the compiler")
	}

	e.compileExprIntoEasm(easm, eif.condition)

	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.NOT,
	})

	jumpIfaddr := easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPI,
	})
	e.compileBlockIntoEasm(easm, eif.ethen)
	bodyEndAddr := easm.program.Size()

	easm.program.Instrutions[jumpIfaddr].Operand = *uint256.NewInt(uint64(bodyEndAddr))
}

func (e *eulang) compileBlockIntoEasm(easm *easm, block *eulBlock) {
	for _, stmt := range block.statements {
		e.compileStatementIntoEasm(easm, stmt)
	}
}

func (e *eulang) compileExprIntoEasm(easm *easm, expr eulExpr) {
	switch expr.kind {
	case eulExprKindFuncCall:
		// TODO temporary solution hard code just one function
		if expr.as.funcCall.name == "write" {
			e.compileExprIntoEasm(easm, expr.as.funcCall.args[0].value)
			easm.pushInstruction(eulvm.Instruction{
				OpCode:  eulvm.OpCode(eulvm.NATIVE),
				Operand: *uint256.NewInt(eulvm.NativeWrite),
			})
		} else if expr.as.funcCall.name == "true" {
			easm.pushInstruction(eulvm.Instruction{
				OpCode:  eulvm.PUSH,
				Operand: *uint256.NewInt(1),
			})
		} else if expr.as.funcCall.name == "false" {
			easm.pushInstruction(eulvm.Instruction{
				OpCode: eulvm.PUSH,
			})
		} else {
			panic("unexpected name")
		}
	case eulExprKindStrLit:
		var addr eulvm.Word = easm.pushStringToMemory(expr.as.strLit)

		pushStrAddrInst := eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: addr,
		}

		easm.pushInstruction(pushStrAddrInst)

		pushStrSizeInst := eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *uint256.NewInt(uint64(len(expr.as.strLit))),
		}

		easm.pushInstruction(pushStrSizeInst)
	case eulExprKindBoolLit:
		if expr.as.boolean {
			easm.pushInstruction(eulvm.Instruction{
				OpCode:  eulvm.PUSH,
				Operand: *uint256.NewInt(1),
			})
		} else {
			easm.pushInstruction(eulvm.Instruction{
				OpCode: eulvm.PUSH,
			})
		}
	default:
		panic("unsupported expression kind")
	}
}
