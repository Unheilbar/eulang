package compiler

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type eulGlobalVar struct {
	addr  uint256.Int //offset inside of preallocated memory
	typee eulType
	loc   eulLoc
	name  string
}

type compiledMap struct {
	loc     eulLoc
	name    string
	keyType eulType
	valType eulType
}

type compiledFunc struct {
	loc      eulLoc
	addr     int
	name     string
	params   []eulFuncParam
	returns  []eulType
	modifier eulFuncModifier
}

type preparedHeader struct {
	name     string
	params   []eulFuncParam
	returns  []eulType
	modifier eulFuncModifier
	body     eulBlock
}

type compiledExpr struct {
	addr  int       // where it starts
	types []eulType // the type that compiled expression returns
	loc   eulLoc
}

type varStorage uint8

const (
	storageKindStack      = iota // offset from the stack frame
	storageKindStatic            // absolute address of the variable in static memory
	storageKindCalldata          // absolute address of the variable in user input (dynamic types are not supported yet)
	storageKindVersion           // address of the variable in merklelized persistent KV
	storageKindPersistent        // address of the variable in not merkilized persistent KV
)

type compiledVar struct {
	name  string
	loc   eulLoc
	etype eulType

	// indicates where is the location of the variable
	storage varStorage
	addr    *uint256.Int // interpretation based on storageKind
}

type eulScope struct {
	parent *eulScope

	compiledVars map[string]compiledVar
	expReturn    []eulType
}

type compileOp struct {
	instruction eulvm.Instruction
	returns     eulType
}

// TODO
var binaryOpByType = map[eulType]map[eulBinaryOpKind]compileOp{
	eulTypeBool: {
		binaryOpKindAnd: {eulvm.Instruction{OpCode: eulvm.AND}, eulTypeBool},
		binaryOpKindOr:  {eulvm.Instruction{OpCode: eulvm.OR}, eulTypeBool},
	},
	eulTypeAddress: {
		binaryOpKindEqual:    {eulvm.Instruction{OpCode: eulvm.EQ}, eulTypeBool},
		binaryOpKindNotEqual: {eulvm.Instruction{OpCode: eulvm.NEQ}, eulTypeBool},
	},
	eulTypeBytes32: {
		binaryOpKindEqual:    {eulvm.Instruction{OpCode: eulvm.EQ}, eulTypeBool},
		binaryOpKindNotEqual: {eulvm.Instruction{OpCode: eulvm.NEQ}, eulTypeBool},
	},
	eulTypei64: {
		binaryOpKindEqual:    {eulvm.Instruction{OpCode: eulvm.EQ}, eulTypeBool},
		binaryOpKindNotEqual: {eulvm.Instruction{OpCode: eulvm.NEQ}, eulTypeBool},
		binaryOpKindLess:     {eulvm.Instruction{OpCode: eulvm.LT}, eulTypeBool},
		binaryOpKindGreater:  {eulvm.Instruction{OpCode: eulvm.GT}, eulTypeBool},
		binaryOpKindMulti:    {eulvm.Instruction{OpCode: eulvm.MUL}, eulTypei64},
		binaryOpKindPlus:     {eulvm.Instruction{OpCode: eulvm.ADD}, eulTypei64},
		binaryOpKindMinus:    {eulvm.Instruction{OpCode: eulvm.SUB}, eulTypei64},
	},
}

// eulang stores all the context of 0euler compiler (compiled functions, scopes, etc.)
type eulang struct {
	funcs       map[string]compiledFunc
	funcHeaders map[string]preparedHeader

	unresolvedAddrIndexes map[string][]int // key func name val index of unresolved call instructions list

	maps map[string]compiledMap

	scope *eulScope

	stackFrameAddr uint256.Int
	frameSize      uint64
}

func NewEulang() *eulang {
	return &eulang{
		scope: &eulScope{
			compiledVars: make(map[string]compiledVar),
		},

		funcs:                 make(map[string]compiledFunc),
		funcHeaders:           make(map[string]preparedHeader),
		unresolvedAddrIndexes: make(map[string][]int),

		maps: make(map[string]compiledMap),
	}
}

func (e *eulang) compileModuleIntoEasm(easm *easm, module eulModule) {
	e.prepareHeaders(module)
	for _, top := range module.tops {
		switch top.kind {
		case eulTopKindFunc:
			e.compileFuncDefIntoEasm(easm, top.as.fdef)
		case eulTopKindVar:
			e.compileVarDefIntoEasm(easm, top.as.vdef, storageKindStatic)
		case eulTopKindMap:
			e.addMapDef(top.as.mdef)
		default:
			panic("try to compile unexpected top kind")
		}
	}
}

func (e *eulang) prepareHeaders(module eulModule) {
	for _, top := range module.tops {
		if top.kind == eulTopKindFunc {
			fd := top.as.fdef
			if _, ok := e.funcHeaders[fd.name]; ok {
				log.Fatalf("%s:%d:%d ERROR double declaration. func '%s' was already defined",
					fd.loc.filepath, fd.loc.row, fd.loc.col, fd.name)
			}
			e.funcHeaders[fd.name] = preparedHeader{
				name:     fd.name,
				params:   fd.params,
				modifier: fd.modifier,
				returns:  fd.returns,
				body:     fd.body,
			}
		}
	}
}

func (e *eulang) addMapDef(mdef eulMapDef) {
	// TODO validate key and value types. Only few types are available for usage as map keys/values
	compMap, ok := e.maps[mdef.name]
	if ok {
		log.Fatalf("%s:%d:%d ERROR map '%s' was already declared. First declaration: %s:%d:%d",
			mdef.loc.filepath, mdef.loc.row, mdef.loc.col, mdef.name, compMap.loc.filepath, compMap.loc.row, compMap.loc.col)
	}

	e.maps[mdef.name] = compiledMap{
		loc:     mdef.loc,
		name:    mdef.name,
		keyType: mdef.keyType,
		valType: mdef.valType,
	}
}

// TODO we don't check uniquness of global variables yet
func (e *eulang) compileVarDefIntoEasm(easm *easm, vd eulVarDef, storage varStorage) {
	_ = e.compileVarIntoEasm(easm, vd, storage)
	if vd.hasInit {
		switch storage {
		case storageKindStatic:
			log.Fatalf("%s:%d:%d ERROR can't assign global variables yet",
				vd.loc.filepath, vd.loc.row, vd.loc.col)

		}
		e.compileVarAssignIntoEasm(easm, eulVarAssign{
			name:  vd.name,
			value: vd.init,
			loc:   vd.loc,
		})
	}
}

func (e *eulang) compileReturnIntoEasm(easm *easm, ret eulReturn) {
	// TODO currently doesn't support return for external functions

	if len(ret.returnExprs) != len(e.scope.expReturn) {
		log.Fatalf("%s:%d:%d ERROR expect return '%d' values but got '%d'",
			ret.loc.filepath, ret.loc.row, ret.loc.col, len(e.scope.expReturn), len(ret.returnExprs))
	}

	for i := len(ret.returnExprs) - 1; i >= 0; i-- {
		expr := ret.returnExprs[i]
		expType := e.scope.expReturn[i]

		cexp := e.compileExprIntoEasm(easm, expr)
		if expType != cexp.types[0] {
			log.Fatalf("%s:%d:%d ERROR expected return '%s' but got '%s'",
				cexp.loc.filepath, cexp.loc.row, cexp.loc.col, eulTypes[expType], eulTypes[cexp.types[0]])
		}

		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.SWAP,
			Operand: *uint256.NewInt(1),
		})
	}

	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.RET},
	)
	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.STOP},
	)
}

func (e *eulang) compileVarIntoEasm(easm *easm, vd eulVarDef, storage varStorage) compiledVar {
	var cv compiledVar
	cv.addr = new(uint256.Int)
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
		*cv.addr = easm.pushByteArrToMemory([]byte{0})
	case storageKindStack:
		e.frameSize += 32 // all var have the size of 1 machine word
		*cv.addr = *uint256.NewInt(e.frameSize)
	case storageKindCalldata:
		// for calldata address gets calculated in parent call
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
	if idxs, ok := e.unresolvedAddrIndexes[fd.name]; ok {
		for _, idx := range idxs {
			easm.program.Instrutions[idx].Operand = *uint256.NewInt(uint64(f.addr))
		}
	}
	f.name = fd.name
	f.loc = fd.loc
	f.params = fd.params
	f.modifier = fd.modifier
	f.returns = fd.returns

	e.funcs[f.name] = f
	e.pushNewScope(fd.returns)

	// compile func params
	if fd.modifier == eulModifierKindExternal && len(fd.params) != 0 {
		e.compileExternalFuncParams(easm, fd.params)
	} else {
		e.compileInternalFuncParams(easm, fd.params)
	}

	e.compileBlockIntoEasm(easm, &fd.body)

	e.popScope()

	// TODO check if function returns values it's supposed to return
	if fd.modifier != eulModifierKindExternal {
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.RET},
		)
	} else {
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.STOP},
		)
	}
}

func (e *eulang) compileExternalFuncParams(easm *easm, params []eulFuncParam) {
	// TODO for now each param has fixed size 32 bytes
	baseOffset := eulvm.WordLength.Clone() // first 32 bytes determine called method
	startAddress := baseOffset.Clone()
	for _, param := range params {
		var vd eulVarDef
		vd.name = param.name
		vd.loc = param.loc
		vd.etype = param.typee

		varr := e.compileVarIntoEasm(easm, vd, storageKindCalldata)
		*varr.addr = *startAddress
		startAddress.Add(startAddress, baseOffset.Clone())
		e.compileGetVarAddr(easm, &varr)
	}
}

func (e *eulang) compileInternalFuncParams(easm *easm, params []eulFuncParam) {
	for _, param := range params {
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

func (e *eulang) compileStatementIntoEasm(easm *easm, stmt eulStatement) {
	switch stmt.kind {
	case eulStmtKindExpr:
		expr := e.compileExprIntoEasm(easm, stmt.as.expr)
		for _, t := range expr.types {
			if t != eulTypeVoid {

				easm.pushInstruction(eulvm.Instruction{
					OpCode: eulvm.DROP,
				})
			}
		}
	case eulStmtKindIf:
		e.compileIfIntoEasm(easm, stmt.as.eif)
	case eulStmtKindVarAssign:
		e.compileVarAssignIntoEasm(easm, stmt.as.varAssign)
	case eulStmtKindMultiVarAssign:
		e.compileMultiVarAssignIntoEasm(easm, stmt.as.multiAssign)
	case eulStmtKindWhile:
		e.compileWhileIntoEasm(easm, stmt.as.while)
	case eulStmtKindVarDef:
		// TODO statements as vars should be compiling into stack storage for compile-time known variables
		e.compileVarDefIntoEasm(easm, stmt.as.vardef, storageKindStack)
	case eulStmtKindMapWrite:
		e.compileMapWriteIntoEasm(easm, stmt.as.mapWrite)
	case eulStmtKindReturn:
		e.compileReturnIntoEasm(easm, stmt.as.freturn)
	default:
		panic(fmt.Sprintf("stmt kind doesn't exist kind %d", stmt.kind))
	}
}

func (e *eulang) compileWhileIntoEasm(easm *easm, w eulWhile) {
	condExpr := e.compileExprIntoEasm(easm, w.condition)
	//TODO later make something like (checkCondExpression cause it seems like reusable)
	//TODO for now we don't have booleans
	if condExpr.types[0] != eulTypeBool {
		log.Fatalf("%s:%d:%d ERROR while condition expression type should be boolean, got %s",
			condExpr.loc.filepath, condExpr.loc.row, condExpr.loc.col, eulTypes[condExpr.types[0]])
	}

	easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.NOT,
	})

	jumpWhileAddr := easm.PushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPI,
	})
	e.pushNewScope(e.scope.expReturn)
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

func (e *eulang) compileMultiVarAssignIntoEasm(easm *easm, multiexp eulMultiAssign) {

	// 1. put all values on the stack | stack looks like [val1,val2,val3...]
	var assignTypes []eulType

	for i := len(multiexp.values) - 1; i >= 0; i-- {
		exp := multiexp.values[i]
		expr := e.compileExprIntoEasm(easm, exp)
		assignTypes = append(assignTypes, expr.types...)
	}

	if len(multiexp.names) != len(assignTypes) {
		log.Fatalf("%s:%d:%d ERROR expected '%d' values but got '%d'",
			multiexp.loc.filepath, multiexp.loc.row, multiexp.loc.col, len(multiexp.names), len(assignTypes))
	}

	// 2. put addr on the stack one by one with swap and mstore [val1, val2, val3, addr3] -> swap -> [val1, val2, addr3, val3] -> mstore [val1, val2] ...
	for i, name := range multiexp.names {
		vari := e.getCompiledVarByName(name)
		if vari.etype != assignTypes[i] {
			log.Fatalf("%s:%d:%d ERROR do not match types on the left and right side in expression. left side '%s' right '%s'",
				multiexp.loc.filepath, multiexp.loc.row, multiexp.loc.col, eulTypes[vari.etype], eulTypes[assignTypes[i]])
		}

		e.compileGetVarAddr(easm, vari)
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.SWAP,
			Operand: *uint256.NewInt(1),
		})
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.MSTORE256,
		})
	}

}

func (e *eulang) compileVarAssignIntoEasm(easm *easm, expr eulVarAssign) {
	vari := e.getCompiledVarByName(expr.name)

	if vari == nil {
		log.Fatalf("%s:%d:%d ERROR cannot assign not declared variable '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	e.compileGetVarAddr(easm, vari)

	// TODO maybe refactor its later
	if vari.etype == eulTypeBytes32 && expr.value.kind == eulExprKindStrLit {
		expr.value.kind = eulExprKindBytes32Lit
		expr.value.as.bytes32Lit = common.HexToHash(expr.value.as.strLit)
		if expr.value.as.bytes32Lit.Hex() != expr.value.as.strLit {
			log.Fatalf("%s:%d:%d ERROR cannot convert str literal '%s' to bytes32",
				expr.loc.filepath, expr.loc.row, expr.loc.col, expr.value.as.strLit)
		}
	}
	if vari.etype == eulTypeAddress && expr.value.kind == eulExprKindStrLit {
		expr.value.kind = eulExprKindAddressLit
		expr.value.as.addressLit = common.HexToAddress(expr.value.as.strLit)
		if !common.IsHexAddress(expr.value.as.strLit) {
			log.Fatalf("%s:%d:%d ERROR cannot convert str literal '%s' to address",
				expr.loc.filepath, expr.loc.row, expr.loc.col, expr.value.as.strLit)
		}
	}
	//if vari.etype == eulTyp

	compiledExpr := e.compileExprIntoEasm(easm, expr.value)

	if compiledExpr.types[0] != vari.etype {
		log.Fatalf("%s:%d:%d ERROR do not match types on the left and right side in expression '%s'. left side '%s' right '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name, eulTypes[vari.etype], eulTypes[compiledExpr.types[0]])
	}

	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.MSTORE256,
	})
}

func (e *eulang) compileMapWriteIntoEasm(easm *easm, mwrite eulMapWrite) {
	mdef, ok := e.maps[mwrite.name]
	if !ok {
		log.Fatalf("%s:%d:%d ERROR cannot write into undefined map '%s'",
			mwrite.loc.filepath, mwrite.loc.row, mwrite.loc.col, mwrite.name)
	}

	mapprefix := strToWords(fmt.Sprint(mdef.name, "."))
	if len(mapprefix) > 1 {

		log.Fatalf("%s:%d:%d ERROR map '%s' name is too long",
			mwrite.loc.filepath, mwrite.loc.row, mwrite.loc.col, mwrite.name)
	}

	key := e.compileExprIntoEasm(easm, mwrite.key)
	val := e.compileExprIntoEasm(easm, mwrite.value)

	if mdef.keyType != key.types[0] {
		log.Fatalf("%s:%d:%d ERROR map '%s' write key type doesnt match. expected '%s' but got '%s'",
			mwrite.loc.filepath, mwrite.loc.row, mwrite.loc.col, mwrite.name, eulTypes[mdef.keyType], eulTypes[key.types[0]])
	}
	if mdef.valType != val.types[0] {
		log.Fatalf("%s:%d:%d ERROR map '%s' write val type doesn't match. expected '%s' but got '%s'",
			mwrite.loc.filepath, mwrite.loc.row, mwrite.loc.col, mwrite.name, eulTypes[mdef.valType], eulTypes[val.types[0]])
	}

	// TODO Eulang later add map write for dynamic types (do we need it?)
	easm.PushInstruction(eulvm.Instruction{
		OpCode:  eulvm.MAPVSSTORE,
		Operand: mapprefix[0],
	})
}

func (e *eulang) compileBinaryOpIntoEasm(easm *easm, binOp binaryOp) eulType {
	//TODO in the future probably better to return compiled expression
	lhsCompiled := e.compileExprIntoEasm(easm, binOp.lhs)
	rhsCompiled := e.compileExprIntoEasm(easm, binOp.rhs)

	if len(lhsCompiled.types) != 1 || len(rhsCompiled.types) != 1 {
		log.Fatalf("%s:%d:%d ERROR binary operation can't return more than 1 value",
			binOp.loc.filepath, binOp.loc.row, binOp.loc.col)
	}

	if rhsCompiled.types[0] != lhsCompiled.types[0] {
		log.Fatalf("%s:%d:%d ERROR expression types on left and right side do not match '%s' != '%s'",
			binOp.loc.filepath, binOp.loc.row, binOp.loc.col, eulTypes[lhsCompiled.types[0]], eulTypes[rhsCompiled.types[0]])
	}

	t := lhsCompiled.types[0]

	bOp, ok := binaryOpByType[t][binOp.kind]

	if !ok {
		log.Fatalf("%s:%d:%d ERROR impossible binary operation for types '%s' and '%s' ",
			binOp.loc.filepath, binOp.loc.row, binOp.loc.col, eulTypes[lhsCompiled.types[0]], eulTypes[rhsCompiled.types[0]])
	}

	easm.pushInstruction(bOp.instruction)

	return bOp.returns
}

func (e *eulang) compileVarReadIntoEasm(easm *easm, expr varRead) eulType {
	cvar := e.getCompiledVarByName(expr.name)

	if cvar == nil {
		log.Fatalf("%s:%d:%d ERROR undefined var '%s'",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	e.compileGetVarAddr(easm, cvar)

	//TODO var reads and var write should be depending on type and storage
	switch cvar.storage {
	case storageKindStack:
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.MLOAD,
		})
	case storageKindStatic:
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.MLOAD,
		})
	case storageKindCalldata:
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.DATALOAD,
		})
	}

	return cvar.etype
}

func (e *eulang) compileMapReadIntoEasm(easm *easm, expr mapRead) eulType {
	// TODO for now all maps are in global scope
	mread, ok := e.maps[expr.name]
	if !ok {
		log.Fatalf("%s:%d:%d ERROR map read from undefined map '%s' ",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	mapprefix := strToWords(fmt.Sprint(mread.name, "."))
	if len(mapprefix) > 1 {
		log.Fatalf("%s:%d:%d ERROR map '%s' name is too long",
			expr.loc.filepath, expr.loc.row, expr.loc.col, expr.name)
	}

	compiledKey := e.compileExprIntoEasm(easm, expr.key)
	if compiledKey.types[0] != mread.keyType {
		log.Fatalf("%s:%d:%d ERROR map read from map key type missmatched. expected '%s', but got '%s' ",
			expr.loc.filepath, expr.loc.row, expr.loc.col, eulTypes[mread.keyType], eulTypes[compiledKey.types[0]])
	}

	//TODO for now map read available only for fixed types
	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.MAPVSSLOAD,
		Operand: mapprefix[0],
	})

	return mread.valType
}

func (e *eulang) compileIfIntoEasm(easm *easm, eif eulIf) {
	condExpr := e.compileExprIntoEasm(easm, eif.condition)
	if condExpr.types[0] != eulTypeBool {
		log.Fatalf("%s:%d:%d ERROR if condition expression type should be boolean, got %s",
			eif.loc.filepath, eif.loc.row, eif.loc.col, eulTypes[condExpr.types[0]])
	}

	easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.NOT,
	})

	jmpThenAddr := easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPI,
	})
	e.pushNewScope(e.scope.expReturn)
	e.compileBlockIntoEasm(easm, eif.ethen)
	e.popScope()
	jmpElseAddr := easm.pushInstruction(eulvm.Instruction{
		OpCode: eulvm.JUMPDEST,
	})
	elseAddr := easm.program.Size()
	e.pushNewScope(e.scope.expReturn)
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
			e.compileNativeWriteIntoEasm(easm, expr.as.funcCall)
			cExp.types = []eulType{eulTypeVoid}
		} else if expr.as.funcCall.name == "writef" {
			e.compileNativeWriteFIntoEasm(easm, expr.as.funcCall)
			cExp.types = []eulType{eulTypeVoid}
		} else {
			types := e.compileFuncCallIntoEasm(easm, expr.as.funcCall)
			cExp.types = types
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
		cExp.types = []eulType{eulTypei64}
	case eulExprKindBytes32Lit:
		w := new(uint256.Int)
		w.SetBytes32(expr.as.bytes32Lit.Bytes())
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *w,
		})
		cExp.types = []eulType{eulTypeBytes32}
	case eulExprKindAddressLit:
		w := new(uint256.Int)
		w.SetBytes(expr.as.addressLit.Bytes())
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *w,
		})
		cExp.types = []eulType{eulTypeAddress}
	case eulExprKindIntLit:
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *uint256.NewInt(uint64(expr.as.intLit)),
		})
		cExp.types = []eulType{eulTypei64}
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

		cExp.types = []eulType{eulTypeBool}
	case eulExprKindVarRead:
		cExp.types = []eulType{e.compileVarReadIntoEasm(easm, expr.as.varRead)}
	case eulExprKindBinaryOp:
		cExp.types = []eulType{e.compileBinaryOpIntoEasm(easm, *expr.as.binaryOp)}
	case eulExprKindMapRead:
		cExp.types = []eulType{e.compileMapReadIntoEasm(easm, *expr.as.mapRead)}
	default:
		panic("unsupported expression kind")
	}

	return cExp
}

func (e *eulang) compileNativeWriteFIntoEasm(easm *easm, funcCall eulFuncCall) {
	// TODO add check format string feats to types in funcCall
	// 1. Compile args
	for i := len(funcCall.args) - 1; i >= 0; i-- {
		arg := funcCall.args[i]
		e.compileExprIntoEasm(easm, arg.value)
	}

	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.OpCode(eulvm.NATIVE),
		Operand: *uint256.NewInt(eulvm.NativeWriteF),
	})
}

func (e *eulang) compileNativeWriteIntoEasm(easm *easm, funcall eulFuncCall) {
	e.compileExprIntoEasm(easm, funcall.args[0].value)
	easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.OpCode(eulvm.NATIVE),
		Operand: *uint256.NewInt(eulvm.NativeWrite),
	})
}

func (e *eulang) compileFuncCallIntoEasm(easm *easm, funcCall eulFuncCall) []eulType {
	//TODO eulang/issues/10 add deffered compiled function addresses resolving later

	header, ok := e.funcHeaders[funcCall.name]
	if !ok {
		log.Fatalf("%s:%d:%d ERROR calling undefined func '%s'",
			funcCall.loc.filepath, funcCall.loc.row, funcCall.loc.col, funcCall.name)
	}

	if header.modifier == eulModifierKindExternal {
		log.Fatalf("%s:%d:%d ERROR calling func with external modifier is forbidden",
			funcCall.loc.filepath, funcCall.loc.row, funcCall.loc.col)
	}

	//compile args
	{
		if len(funcCall.args) != len(header.params) {
			log.Fatalf("%s:%d:%d ERROR funcall arity missmatch. Expected '%d' arguments but got '%d' instead ",
				funcCall.loc.filepath, funcCall.loc.row, funcCall.loc.col, len(header.params), len(funcCall.args))
		}
		for i := len(funcCall.args) - 1; i >= 0; i-- {
			arg := funcCall.args[i]
			param := header.params[i]
			expr := e.compileExprIntoEasm(easm, arg.value)
			if param.typee != expr.types[0] {
				log.Fatalf("%s:%d:%d ERROR funcall type missmatch. Expected '%s' type but got '%s' instead ",
					expr.loc.filepath, expr.loc.row, expr.loc.col, eulTypes[param.typee], eulTypes[expr.types[0]])
			}
		}
	}
	//framez
	e.compilePushNewFrame(easm)
	cf, ok := e.funcs[funcCall.name]

	idx := easm.pushInstruction(eulvm.Instruction{
		OpCode:  eulvm.CALL,
		Operand: *uint256.NewInt(uint64(cf.addr)),
	})

	if !ok {
		e.unresolvedAddrIndexes[funcCall.name] = append(e.unresolvedAddrIndexes[funcCall.name], idx) // if func defenition is not compiled yet, we will resolve its address later
	}

	e.compilePopFrame(easm)

	return header.returns
}

func (e *eulang) pushNewScope(expReturn []eulType) {
	scope := eulScope{
		expReturn:    expReturn,
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
			Operand: *cv.addr,
		})
	case storageKindStack:
		e.compileReadFrameAddr(easm)
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *cv.addr,
		})
		easm.pushInstruction(eulvm.Instruction{
			OpCode: eulvm.SUB,
		})
	case storageKindCalldata:
		easm.pushInstruction(eulvm.Instruction{
			OpCode:  eulvm.PUSH,
			Operand: *cv.addr,
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
func (e *eulang) GenerateInput(method string, args []string) []byte {
	var input []byte
	meth := uint256.NewInt(uint64(e.funcs[method].addr)).Bytes32()
	input = append(input, meth[:]...)
	params := e.funcs[method].params
	if len(params) != len(args) {
		log.Fatalf("params and args amount doesn't match. Got '%d' want '%d'", len(args), len(params))
	}
	for i, param := range params {
		switch param.typee {
		case eulTypei64:
			argi64, err := strconv.Atoi(args[i])
			if err != nil {
				log.Fatalf("param arg types doesn't match. Cant convert '%v' to int", args[i])
			}
			arg := uint256.NewInt(uint64(argi64)).Bytes32()
			input = append(input, arg[:]...)
		case eulTypeBool:
			var arg [32]byte
			if args[i] == "true" {
				arg = uint256.NewInt(uint64(1)).Bytes32()

			} else if args[i] == "false" {
				arg = uint256.NewInt(uint64(0)).Bytes32()
			}

			input = append(input, arg[:]...)
		case eulTypeBytes32:
			harg := common.HexToHash(args[i])
			if harg.Hex() != args[i] {
				log.Fatalf("param arg types doesn't match. Cant convert '%v' to bytes32", args[i])
			}
			input = append(input, harg.Bytes()...)
		case eulTypeAddress:
			if !common.IsHexAddress(args[i]) {
				log.Fatalf("param arg types doesn't match. Cant convert '%v' to address", args[i])
			}
			harg := common.HexToAddress(args[i])
			arg := new(uint256.Int).SetBytes(harg.Bytes()).Bytes32()

			input = append(input, arg[:]...)
		default:
			panic("unrecognized eulang type in function call ")
			//eulTypeVoid
			//eulTypeBytes32
			//eulTypeAddress
		}

	}
	return input[:]
}
