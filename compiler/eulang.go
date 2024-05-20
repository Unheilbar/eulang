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
	loc    eulLoc
	addr   int
	name   string
	params []eulFuncParam
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

	stackFrameAddr uint256.Int
	frameSize      uint64
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

	_, ok := e.scope.compiledVars[vd.name]
	if ok {
		log.Fatalf("%s:%d:%d ERROR variable '%s' was already delcared at current scope ",
			vd.loc.filepath, vd.loc.row, vd.loc.col, vd.name)
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
		e.frameSize += 32 // all var have the size of 1 machine word
		cv.addr = *uint256.NewInt(e.frameSize)
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
	f.params = fd.params
	e.funcs[f.name] = f
	e.pushNewScope()

	// compile func params
	{
		for _, param := range fd.params {
			var vd eulVarDef
			vd.name = param.name
			vd.loc = param.loc
			vd.etype = param.typee

			varr := e.compileVarIntoEasm(easm, vd, storageKindStack)
			easm.pushInstruction(eulvm.Instruction{
				OpCode:  eulvm.SWAP,
				Operand: *uint256.NewInt(1),
			})
			e.compileGetVarAddr(easm, &varr)
			easm.pushInstruction(eulvm.Instruction{
				OpCode:  eulvm.SWAP,
				Operand: *uint256.NewInt(1),
			})
			easm.PushInstruction(eulvm.Instruction{
				OpCode: eulvm.MSTORE256,
			})
		}
	}

	e.compileBlockIntoEasm(easm, &fd.body)
	e.popScope()

	//TODO euler later add external modifier
	if f.name != "entry" {
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.RET},
		)
	} else {
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.STOP},
		)
	}
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
		// TODO statements as vars should be compiling into stack storage for compile-time known variables
		e.compileVarDefIntoEasm(easm, stmt.as.vardef, storageKindStack)
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

	e.compileGetVarAddr(easm, vari)
	compiledExpr := e.compileExprIntoEasm(easm, expr.value)

	if compiledExpr.typee != vari.etype {
		log.Fatalf("%s:%d:%d ERROR do not match types on the left and right side in expression '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

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

func (e *eulang) compileVarReadIntoEasm(easm *easm, expr varRead) eulType {
	cvar := e.getCompiledVarByName(expr.name)

	if cvar == nil {
		log.Fatalf("%s:%d:%d ERROR condition expression type should be boolean, got %s",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	e.compileGetVarAddr(easm, cvar)

	//TODO var reads and var write should be depending on type and storage
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MLOAD,
	})

	return cvar.etype
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
			e.compileFuncCallIntoEasm(easm, expr.as.funcCall)
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
		cExp.typee = e.compileVarReadIntoEasm(easm, expr.as.varRead)
	case eulExprKindBinaryOp:
		cExp.typee = e.compileBinaryOpIntoEasm(easm, *expr.as.binaryOp)
	default:
		panic("unsupported expression kind")
	}

	return cExp
}

func (e *eulang) compileFuncCallIntoEasm(easm *easm, funcCall eulFuncCall) {
	//TODO add deffered compiled function addresses resolving later
	compiledFunc, ok := e.funcs[funcCall.name]
	if !ok {
		panic(fmt.Sprintf("undefined compiled function %s", funcCall.name))
	}

	//compile args
	{
		if len(funcCall.args) != len(compiledFunc.params) {
			log.Fatalf("%s:%d:%d ERROR funcall arity missmatch. Expected '%d' arguments but got '%d' instead ",
				funcCall.loc.filepath, funcCall.loc.row, funcCall.loc.col, len(compiledFunc.params), len(funcCall.args))
		}
		for i, arg := range funcCall.args {
			param := compiledFunc.params[i]
			expr := e.compileExprIntoEasm(easm, arg.value)
			if param.typee != expr.typee {
				log.Fatalf("%s:%d:%d ERROR funcall type missmatch. Expected '%s' type but got '%s' instead ",
					expr.loc.filepath, expr.loc.row, expr.loc.col, eulTypes[param.typee], eulTypes[expr.typee])
			}
		}
	}
	//framez
	e.compilePushNewFrame(easm)
	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.CALL,
		Operand: *uint256.NewInt(uint64(compiledFunc.addr)),
	})

	e.compilePopFrame(easm)
}

func (e *eulang) pushNewScope() {
	scope := eulScope{
		parent:       e.scope,
		compiledVars: make(map[string]compiledVar),
	}
	e.scope = &scope
}

// Scope operations
func (e *eulang) popScope() {
	if e.scope == nil {
		panic("try pop nil scope")
	}

	//framez
	var deallocs uint64
	for _, v := range e.scope.compiledVars {
		if v.storage == storageKindStack {
			deallocs += 32
		}
	}

	e.frameSize -= deallocs

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

// Frame operations
func (e *eulang) compilePushNewFrame(easm *easm) {
	// 1. Read frame addr
	e.compileReadFrameAddr(easm)

	// 2. Offset the frame addr to find the top of the stack
	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.PUSH,
		Operand: *uint256.NewInt(e.frameSize),
	})
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.SUB,
	})

	// 3. store prev stack frame address
	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.PUSH,
		Operand: *uint256.NewInt(32), // the size of the machine word
	})
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.SUB,
	})
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.DUP,
	})
	e.compileReadFrameAddr(easm)
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MSTORE256,
	})

	// 4. Redirect the current frame
	e.compileWriteFrameAddr(easm)

}

func (e *eulang) compilePopFrame(easm *easm) {
	// 1. read frame addr
	e.compileReadFrameAddr(easm)

	// 2. read prev frame addr
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MLOAD,
	})

	// 3. write prev frame addr
	e.compileWriteFrameAddr(easm)
}

// Reads current stack frame address from VM memory
func (e *eulang) compileReadFrameAddr(easm *easm) {
	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.PUSH,
		Operand: e.stackFrameAddr,
	})
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MLOAD,
	})
}

// Writes current stack frame address to VM memory
func (e *eulang) compileWriteFrameAddr(easm *easm) {
	easm.pushInstruction(eulvm.Instruction{ // [...stack...] -> [...stack...frameAddr]
		OpCode:  eulvm.PUSH,
		Operand: e.stackFrameAddr,
	})
	easm.pushInstruction(eulvm.Instruction{ // [...stack...frameAddr] -> [...stack...frameAddr...]
		OpCode:  eulvm.SWAP,
		Operand: *uint256.NewInt(1),
	})
	easm.pushInstruction(eulvm.Instruction{ // [...stack...frameAddr...] -> [...stack...]
		OpCode: eulvm.MSTORE256,
	})
}

func (e *eulang) compileGetVarAddr(easm *easm, cv *compiledVar) {
	switch cv.storage {
	case storageKindStatic:
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: cv.addr,
		})
	case storageKindStack:
		e.compileReadFrameAddr(easm)
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: cv.addr,
		})
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.SUB,
		})
	default:
		panic("compiling other storage types is not implemented yet")
	}
}

func (e *eulang) prepareVarStack(easm *easm, stackSize uint) {
	arr := make([]byte, stackSize*32)

	easm.pushByteArrToMemory(arr)
	result := easm.pushWordToMemory(*uint256.NewInt(uint64(stackSize * 32)))

	e.stackFrameAddr = result
}

// // TODO euler later can add here function arguments
func (e *eulang) GenerateInput(method string) []byte {
	meth := uint256.NewInt(uint64(e.funcs[method].addr)).Bytes32()
	return meth[:]
}
