package compiler

import (
	"fmt"
	"log"

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
	gvars map[string]eulGlobalVar
}

func NewEulang() *eulang {
	return &eulang{
		gvars: make(map[string]eulGlobalVar),
	}
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

// TODO we don't check uniquness of global variables yet
func (e *eulang) compileVarDefIntoEasm(easm *easm, vd eulVarDef) {
	var gv eulGlobalVar
	gv.addr = easm.pushByteArrToMemory([]byte{0})
	gv.name = vd.name

	e.gvars[gv.name] = gv
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
	case eulStmtKindVarAssign:
		e.compileVarAssignIntoEasm(easm, stmt.as.varAssign)
	case eulStmtKindWhile:
		e.compileWhileIntoEasm(easm, stmt.as.while)
	default:
		panic(fmt.Sprintf("stmt kind doesn't exist kind %d", stmt.kind))
	}
}

func (e *eulang) compileWhileIntoEasm(easm *easm, w eulWhile) {
	condAddr := easm.program.Size()
	e.compileExprIntoEasm(easm, w.condition)
	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.NOT,
	})

	jumpWhileAddr := easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPI,
	})

	e.compileBlockIntoEasm(easm, &w.body)
	easm.PushInstruction(eulvm.Instruction{
		OpCode:  eulvm.JUMPDEST,
		Operand: *uint256.NewInt(uint64(condAddr)),
	})
	bodyEnd := easm.program.Size()

	// resolve deferred
	easm.program.Instrutions[jumpWhileAddr].Operand = *uint256.NewInt(uint64(bodyEnd))
}

func (e *eulang) compileVarAssignIntoEasm(easm *easm, expr eulVarAssign) {
	vari, ok := e.gvars[expr.name]

	if !ok {
		log.Fatalf("%s:%d:%d ERROR cannot assign not declared variable '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.PUSH,
		Operand: vari.addr,
	})
	e.compileExprIntoEasm(easm, expr.value)
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MSTORE256,
	})
}

func (e *eulang) compileBinaryOpIntoEasm(easm *easm, binOp binaryOp) {
	switch binOp.kind {
	case binaryOpLess:
		e.compileExprIntoEasm(easm, binOp.rhs)
		e.compileExprIntoEasm(easm, binOp.lhs)
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.LT,
		})
	case binaryOpPlus:
		e.compileExprIntoEasm(easm, binOp.lhs)
		e.compileExprIntoEasm(easm, binOp.rhs)
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.ADD,
		})
	default:
		panic("compiling bin op unreachable")
	}
}

func (e *eulang) compileVarReadIntoEasm(easm *easm, expr varRead) {
	glVar, ok := e.gvars[expr.name]
	if !ok {
		log.Fatalf("%s:%d:%d ERROR cannot read not declared variable '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.PUSH,
		Operand: glVar.addr,
	})
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MLOAD,
	})
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

// returns address of the end of the block
func (e *eulang) compileBlockIntoEasm(easm *easm, block *eulBlock) int {
	if block == nil {
		return easm.program.Size()
	}
	for _, stmt := range block.statements {
		e.compileStatementIntoEasm(easm, stmt)
	}

	return easm.program.Size()
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
	case eulExprKindIntLit:
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *uint256.NewInt(uint64(expr.as.intLit)),
		})
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
	case eulExprKindVarRead:
		e.compileVarReadIntoEasm(easm, expr.as.varRead)
	case eulExprKindBinaryOp:
		e.compileBinaryOpIntoEasm(easm, *expr.as.binaryOp)
	default:
		panic("unsupported expression kind")
	}
}
