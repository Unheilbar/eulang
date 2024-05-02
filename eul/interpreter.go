package eul

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/holiman/uint256"
)

type OpCode byte

type Stack struct {
	data []uint256.Int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]uint256.Int, 0, size),
	}
}

func (st *Stack) pop() (ret uint256.Int) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *Stack) peek() (ret *uint256.Int) {
	ret = &st.data[len(st.data)-1]
	return
}

func (st *Stack) push(d *uint256.Int) {
	st.data = append(st.data, *d)
}

// 0x0 arithmetic range
const (
	STOP OpCode = iota
	ADD
	SUB
	POP
	PUSH
)

// 0x10 range - comparison ops.
const (
	LT OpCode = iota + 0x10
	GT
	EQ
)

// 0x20 - debug
const (
	PRINT OpCode = iota + 0x20
)

type Interpreter struct {
}

func (in *Interpreter) Run(code []byte) {
	var pc uint64
	s := NewStack(256)
	scope := &ScopeContext{
		stack: s,
		code:  code,
	}
	for {
		ex := getEx(OpCode(code[pc]))
		pc++
		_, err := ex(&pc, scope)

		if err != nil {
			return
		}
	}
}

type ScopeContext struct {
	stack *Stack
	code  []byte
}

type execution func(pc *uint64, scope *ScopeContext) ([]byte, error)

func getEx(opc OpCode) execution {
	switch opc {
	case POP:
		return opPop
	case PUSH:
		return opPush
	case ADD:
		return opAdd
	case SUB:
		return opSub
	case GT:
		return opGt
	case EQ:
		return opEq
	case PRINT:
		return opPrint
	case STOP:
		return opStop
	}

	panic("unkown opcode")
}

func opPop(pc *uint64, scope *ScopeContext) ([]byte, error) {
	scope.stack.pop()
	return nil, nil
}

func opPush(pc *uint64, scope *ScopeContext) ([]byte, error) {
	var integer = new(uint256.Int)
	scope.stack.push(integer.SetUint64(uint64(binary.LittleEndian.Uint64(scope.code[*pc : *pc+8]))))
	*pc += 8
	return nil, nil
}

func opAdd(pc *uint64, scope *ScopeContext) ([]byte, error) {
	x, y := scope.stack.pop(), scope.stack.peek()
	y.Add(&x, y)
	return nil, nil
}

func opSub(pc *uint64, scope *ScopeContext) ([]byte, error) {
	fmt.Println("sub")
	return nil, nil
}

func opGt(pc *uint64, scope *ScopeContext) ([]byte, error) {
	fmt.Println("gt")
	return nil, nil
}

func opEq(pc *uint64, scope *ScopeContext) ([]byte, error) {
	fmt.Println("eq")
	return nil, nil
}

var errStopToken = errors.New("reached stop op")

func opStop(pc *uint64, scope *ScopeContext) ([]byte, error) {
	fmt.Println("stop")
	return nil, errStopToken
}

func opPrint(pc *uint64, scope *ScopeContext) ([]byte, error) {
	fmt.Println(scope.stack.peek().Uint64())
	return nil, nil
}
