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

type varStorage uint8

const (
	storageKindStack      = iota // offset from the stack frame
	storageKindStatic            // absolute address of the variable in static memory
	storageKindVersion           // address of the variable in merklelized persistent KV
	storageKindPersistent        // address of the variable in not merkilized persistent KV
)

type compiledVar struct {
	name  string
	loc   eulLoc
	etype eulType

	// indicates where is the location of the variable
	storage varStorage
	addr    uint256.Int // interpretation based on storageKind
}

type eulScope struct {
	parent *eulScope

	compiledVars map[string]compiledVar
}

// eulang stores all the context of euler compiler (compiled functions, scopes, etc.)
type eulang struct {
	funcs map[string]compiledFunc

	scope *eulScope
}

func NewEulang() *eulang {
	return &eulang{
		scope: &eulScope{
			compiledVars: make(map[string]compiledVar),
		},
		funcs: make(map[string]compiledFunc),
	}
}

func (e *eulang) compileModuleIntoEasm(easm *easm, module eulModule) {
	for _, top := range module.tops {
		switch top.kind {
		case eulTopKindFunc:
			e.compileFuncDefIntoEasm(easm, top.as.fdef)
		case eulTopKindVar:
			e.compileVarDefIntoEasm(easm, top.as.vdef, storageKindStatic)
		default:
			panic("try to compile unexpected top kind")
		}
	}
}

// TODO we don't check uniquness of global variables yet
func (e *eulang) compileVarDefIntoEasm(easm *easm, vd eulVarDef, storage varStorage) {
	_ = e.compileVarIntoEasm(easm, vd, storage)
	if vd.hasInit {
		panic("var initialization after declaration is not implemented yet")
	}

}

func (e *eulang) compileVarIntoEasm(easm *easm, vd eulVarDef, storage varStorage) compiledVar {
	var cv compiledVar

	if vd.etype == eulTypeVoid {
		log.Fatalf("%s:%d:%d ERROR define variable with type void is not allowed (maybe in the future) ",
			vd.loc.filepath, vd.loc.row, vd.loc.col)
	}
	// NOTE Eulang doesn't warn about shadowing?

	cv.name = vd.name
	cv.loc = vd.loc
	cv.etype = vd.etype
	cv.storage = storage

	switch storage {
	case storageKindStatic:
		cv.addr = easm.pushByteArrToMemory([]byte{0})
	case storageKindStack:
		panic("storing variables onto stack will be implemented after introducing stack frames")
	default:
		panic("other storage kinds are not implemented yet")
	}

	e.scope.compiledVars[cv.name] = cv
	return cv
}

func (e *eulang) compileFuncDefIntoEasm(easm *easm, fd eulFuncDef) {
	var f compiledFunc
	if _, ok := e.funcs[fd.name]; ok {
		log.Fatalf("%s:%d:%d ERROR double declaration. func '%s' was already defined",
			fd.loc.filepath, fd.loc.row, fd.loc.col, fd.name)
	}

	f.addr = easm.program.Size()
	f.name = fd.name
	f.loc = fd.loc
	e.funcs[f.name] = f
	e.pushNewScope()
	e.compileBlockIntoEasm(easm, &fd.body)
	e.popScope()
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
	case eulStmtKindVarDef:
		e.compileVarDefIntoEasm(easm, stmt.as.vardef, storageKindStatic)
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
	e.pushNewScope()
	e.compileBlockIntoEasm(easm, &w.body)
	e.popScope()
	easm.PushInstruction(eulvm.Instruction{
		OpCode:  eulvm.JUMPDEST,
		Operand: *uint256.NewInt(uint64(condExpr.addr)),
	})
	bodyEnd := easm.program.Size()

	// resolve deferred
	easm.program.Instrutions[jumpWhileAddr].Operand = *uint256.NewInt(uint64(bodyEnd))
}

func (e *eulang) compileVarAssignIntoEasm(easm *easm, expr eulVarAssign) {
	vari := e.getCompiledVarByName(expr.name)

	if vari == nil {
		log.Fatalf("%s:%d:%d ERROR cannot assign not declared variable '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	//TODO var reads and var write should be depending on type and storage
	if vari.storage != storageKindStatic {
		panic("assigning non-static variables is not implemented yet")
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

	cvar := e.getCompiledVarByName(expr.name)
	result.typee = cvar.etype
	result.loc = cvar.loc

	//TODO var reads and var write should be depending on type and storage
	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.PUSH,
		Operand: cvar.addr,
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
	e.pushNewScope()
	e.compileBlockIntoEasm(easm, eif.ethen)
	e.popScope()
	jmpElseAddr := easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPDEST,
	})
	elseAddr := easm.program.Size()
	e.pushNewScope()
	e.compileBlockIntoEasm(easm, eif.elze)
	e.popScope()
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
			//TODO add deffered compiled function addresses resolving later
			compiledFunc, ok := e.funcs[expr.as.funcCall.name]
			if !ok {
				panic(fmt.Sprintf("undefined compiled function %s", expr.as.funcCall.name))
			}
			easm.pushInstruction(eulvm.Instruction{
				OpCode:  eulvm.CALL,
				Operand: *uint256.NewInt(uint64(compiledFunc.addr)),
			})
			cExp.typee = eulTypeVoid
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

func (e *eulang) pushNewScope() {
	scope := eulScope{
		parent:       e.scope,
		compiledVars: make(map[string]compiledVar),
	}
	e.scope = &scope
}

func (e *eulang) popScope() {
	//TODO dealloc stack after introducing frames

	e.scope = e.scope.parent
}

func (e *eulang) getCompiledVarByName(name string) *compiledVar {
	for scope := e.scope; scope != nil; scope = scope.parent {
		if cvar, ok := scope.compiledVars[name]; ok {
			return &cvar
		}
	}

	return nil
}

// // TODO euler later can add here function arguments
func (e *eulang) GenerateInput(method string) []byte {
	k := uint256.NewInt(uint64(e.funcs[method].addr)).Bytes32()
	return k[:]
}
