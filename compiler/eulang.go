package compiler

import (
	"fmt"
	"log"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/holiman/uint256"
)

type eulGlobalVar struct {
	addr  uint256.Int //offset inside of preallocated memory
	typee eulType
	loc   eulLoc
	name  string
}

type compiledFunc struct {
	loc  eulLoc
	addr int
	name string

	//TODO later extend with arguments and return type info
}

type compiledExpr struct {
	addr  int     // where it starts
	typee eulType // the type that compiled expression returns
	loc   eulLoc
}

// eulang stores all the context of euler compiler (compiled functions, scopes, etc.)
type eulang struct {
	funcs map[string]compiledFunc

	// TODO maybe better make a map [name]>globalVar
	gvars map[string]eulGlobalVar
}

func NewEulang() *eulang {
	return &eulang{
		gvars: make(map[string]eulGlobalVar),
		funcs: make(map[string]compiledFunc),
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
	if vd.etype == eulTypeVoid {
		log.Fatalf("%s:%d:%d ERROR define variable with type void is not allowed (maybe in the future) ",
			vd.loc.filepath, vd.loc.row, vd.loc.col)
	}

	var gv eulGlobalVar
	if _, ok := e.gvars[vd.name]; ok {
		log.Fatalf("%s:%d:%d ERROR variable '%s' was already defined",
			vd.loc.filepath, vd.loc.row, vd.loc.col, vd.name)
	}

	gv.addr = easm.pushByteArrToMemory([]byte{0})
	gv.name = vd.name
	gv.typee = vd.etype
	gv.loc = vd.loc

	e.gvars[gv.name] = gv
}

func (e *eulang) compileFuncDefIntoEasm(easm *easm, fd eulFuncDef) {
	var f compiledFunc
	f.addr = easm.program.Size()
	f.name = fd.name
	f.loc = fd.loc
	e.funcs[f.name] = f
	e.compileBlockIntoEasm(easm, &fd.body)
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.RET},
	)
}

func (e *eulang) compileStatementIntoEasm(easm *easm, stmt eulStatement) {
	switch stmt.kind {
	case eulStmtKindExpr:
		expr := e.compileExprIntoEasm(easm, stmt.as.expr)
		if expr.typee != eulTypeVoid {
			// WE need to drop any result of statement as expression from stack because function must have it's return address on the stack
			easm.pushInstruction(eulvm.Instruction{
				OpCode: eulvm.DROP,
			})
		}
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
	condExpr := e.compileExprIntoEasm(easm, w.condition)
	//TODO later make something like (checkCondExpression cause it seems like reusable)
	//TODO for now we don't have booleans
	if condExpr.typee != eulTypei64 {
		log.Fatalf("%s:%d:%d ERROR condition expression type should be boolean, got %s",
			condExpr.loc.filepath, condExpr.loc.row, condExpr.loc.col, eulTypes[condExpr.typee])
	}

	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.NOT,
	})

	jumpWhileAddr := easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPI,
	})

	e.compileBlockIntoEasm(easm, &w.body)
	easm.PushInstruction(eulvm.Instruction{
		OpCode:  eulvm.JUMPDEST,
		Operand: *uint256.NewInt(uint64(condExpr.addr)),
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

func (e *eulang) compileBinaryOpIntoEasm(easm *easm, binOp binaryOp) eulType {
	//TODO in the future probably better to return compiled expression
	var returnType eulType
	rhsCompiled := e.compileExprIntoEasm(easm, binOp.rhs)
	lhsCompiled := e.compileExprIntoEasm(easm, binOp.lhs)
	{
		if rhsCompiled.typee != lhsCompiled.typee {
			log.Fatalf("%s:%d:%d ERROR expression types on left and right side do not match '%s' != '%s'",
				binOp.loc.filepath, binOp.loc.row, binOp.loc.col, eulTypes[lhsCompiled.typee], eulTypes[rhsCompiled.typee])
		}
		if lhsCompiled.typee != eulTypei64 {
			log.Fatalf("%s:%d:%d ERROR invalid type for compare binary operation '%s'",
				binOp.loc.filepath, binOp.loc.row, binOp.loc.col, eulTypes[lhsCompiled.typee])
		}
	}

	switch binOp.kind {
	case binaryOpLess:
		//Typecheck TODO will become a separate function after refacting
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.LT,
		})

		//TODO booleans have no their own type (maybe they don't need one?)
		returnType = rhsCompiled.typee
	case binaryOpPlus:
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.ADD,
		})

		//TODO for now it's the only supported type
		returnType = rhsCompiled.typee
	default:
		panic("compiling bin op unreachable")
	}

	return returnType
}

func (e *eulang) compileVarReadIntoEasm(easm *easm, expr varRead) compiledExpr {
	var result compiledExpr
	result.addr = easm.program.Size()

	glVar, ok := e.gvars[expr.name]
	if !ok {
		log.Fatalf("%s:%d:%d ERROR cannot read not declared variable '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}
	result.typee = glVar.typee
	result.loc = glVar.loc

	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.PUSH,
		Operand: glVar.addr,
	})
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MLOAD,
	})

	return result
}

func (e *eulang) compileIfIntoEasm(easm *easm, eif eulIf) {
	condExpr := e.compileExprIntoEasm(easm, eif.condition)
	if condExpr.typee != eulTypei64 {
		log.Fatalf("%s:%d:%d ERROR condition expression type should be boolean, got %s",
			condExpr.loc.filepath, condExpr.loc.row, condExpr.loc.col, eulTypes[condExpr.typee])
	}

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

func (e *eulang) compileExprIntoEasm(easm *easm, expr eulExpr) compiledExpr {
	var cExp compiledExpr
	cExp.addr = easm.program.Size()
	cExp.loc = expr.loc
	switch expr.kind {
	case eulExprKindFuncCall:
		// TODO temporary solution hard code just one function
		if expr.as.funcCall.name == "write" {
			e.compileExprIntoEasm(easm, expr.as.funcCall.args[0].value)
			easm.pushInstruction(eulvm.Instruction{
				OpCode:  eulvm.OpCode(eulvm.NATIVE),
				Operand: *uint256.NewInt(eulvm.NativeWrite),
			})
			cExp.typee = eulTypeVoid
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

		//TODO strings dont have their own type. But for now they're just i64 pointers to memory
		cExp.typee = eulTypei64
	case eulExprKindIntLit:
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *uint256.NewInt(uint64(expr.as.intLit)),
		})
		cExp.typee = eulTypei64
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

		//TODO booleans have no their own type just yet, but they are saved as i64 words on stack
		cExp.typee = eulTypei64
	case eulExprKindVarRead:
		cExp = e.compileVarReadIntoEasm(easm, expr.as.varRead)
	case eulExprKindBinaryOp:
		cExp.typee = e.compileBinaryOpIntoEasm(easm, *expr.as.binaryOp)
	default:
		panic("unsupported expression kind")
	}

	return cExp
}
