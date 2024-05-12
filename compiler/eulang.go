package compiler

import (
	"fmt"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/holiman/uint256"
)

type eulGlobalVar struct {
	addr uint256.Int //offset inside of preallocated memory
	name string
}

type compiledFuncs struct{}

// eulang stores all the context of euler compiler (compiled functions, scopes, etc.)
type eulang struct {
	funcs []compiledFuncs

	// TODO maybe better make a map [name]>globalVar
	gvars []eulGlobalVar
}

func NewEulang() *eulang {
	return &eulang{}
}

func (e *eulang) compileModuleIntoEasm(easm *easm, module eulModule) {
	for _, top := range module.tops {
		switch top.kind {
		case eulTopKindFunc:
			e.compileFuncDefIntoEasm(easm, top.as.fdef)
		case eulTopKindVar:
			e.compileVarDefIntoEasm(easm, top.as.vdef)
		default:
			panic("try to compile unexpected top kind")
		}
	}
}

func (e *eulang) compileVarDefIntoEasm(easm *easm, vd eulVarDef) {
	var gv eulGlobalVar
	gv.addr = easm.pushByteArrToMemory([]byte{0})
	gv.name = vd.name

	e.gvars = append(e.gvars, gv)
}

func (e *eulang) compileFuncDefIntoEasm(easm *easm, fd eulFuncDef) {
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

	e.compileExprIntoEasm(easm, eif.condition)

	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.NOT,
	})

	jmpThenAddr := easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPI,
	})
	e.compileBlockIntoEasm(easm, eif.ethen)

	jmpElseAddr := easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPDEST,
	})
	elseAddr := easm.program.Size()
	e.compileBlockIntoEasm(easm, eif.elze)

	endAddr := easm.program.Size()

	easm.program.Instrutions[jmpThenAddr].Operand = *uint256.NewInt(uint64(elseAddr))
	easm.program.Instrutions[jmpElseAddr].Operand = *uint256.NewInt(uint64(endAddr))
}

func (e *eulang) compileBlockIntoEasm(easm *easm, block *eulBlock) {
	if block == nil {
		return
	}
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
