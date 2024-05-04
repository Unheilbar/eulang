package euvm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"github.com/holiman/uint256"
)

type Stack struct {
	data []uint256.Int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]uint256.Int, 0, size),
	}
}

func (st *Stack) Dump() {
	fmt.Println("### stack ###")
	if len(st.data) > 0 {
		for i, val := range st.data {
			fmt.Printf("%-3d  %v\n", i, val)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("#############")
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

func (st *Stack) dup(n int) {
	st.push(&st.data[len(st.data)-n])
}

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
	log.Fatal("sub is not implemented")
	return nil, nil
}

func opGt(pc *uint64, scope *ScopeContext) ([]byte, error) {
	fmt.Println("gt")
	return nil, nil
}

func opEq(pc *uint64, scope *ScopeContext) ([]byte, error) {
	x, y := scope.stack.pop(), scope.stack.peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opJumpdest(pc *uint64, scope *ScopeContext) ([]byte, error) {
	*pc = binary.LittleEndian.Uint64(scope.code[*pc : *pc+8])
	return nil, nil
}

var errStopToken = errors.New("reached stop op")

func opStop(pc *uint64, scope *ScopeContext) ([]byte, error) {
	scope.stack.Dump()
	fmt.Println("stop")
	return nil, errStopToken
}

// experimental for tests only
func opPrint(pc *uint64, scope *ScopeContext) ([]byte, error) {
	fmt.Println(scope.stack.peek().Uint64())
	return nil, nil
}

func opInput(pc *uint64, scope *ScopeContext) ([]byte, error) {
	var i int
	fmt.Scanf("%d", &i)
	var integer = new(uint256.Int)
	scope.stack.push(integer.SetUint64(uint64(i)))
	return nil, nil
}
