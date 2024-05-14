package eulvm

import (
	"errors"
	"fmt"

	"github.com/holiman/uint256"
)

const stackCapacity = 1024

type EulVM struct {
	program []Instruction //TODO make unsafe pointer to avoid program size check?
	ip      int

	stack     [stackCapacity]Word
	stackSize int

	memory *Memory
}

const ProgramLimit = 1024

func New(prog Program) *EulVM {
	var m *Memory
	if len(prog.PreallocMemory) != 0 {
		m = NewMemoryWithPrealloc(prog.PreallocMemory)
	} else {
		m = NewMemory()
	}
	return &EulVM{
		program: prog.Instrutions,
		memory:  m,
	}
}

func (e *EulVM) Run() error {
	for i := 0; i < ProgramLimit; i++ {
		err := executeNext(e)
		if err != nil {
			if err == stopToken {
				return nil
			}
			return err
		}
	}
	return errProgramLimitExceeded
}

var (
	errIllegalCall          = errors.New("illegal program call")
	errProgramLimitExceeded = errors.New("program limit cycle exceeded")
	errInvalidOpCodeCalled  = errors.New("opcode doesn't exist")
	errInvalidMemoryAccess  = errors.New("program accessed memory beyond memory capacity")
	errUnknownNative        = errors.New("native function doesn't exists")
)

var stopToken = errors.New("program stopped")

func executeNext(e *EulVM) error {
	if e.ip >= len(e.program) {
		return errIllegalCall
	}

	inst := e.program[e.ip]

	//fmt.Println(e.ip, "-->call", OpCodes[inst.OpCode], "operand", inst.Operand.Uint64())
	switch inst.OpCode {
	case PUSH:
		e.stackSize++
		e.stack[e.stackSize] = inst.Operand
		e.ip++
		return nil
	case DUP:
		e.stackSize++
		e.stack[e.stackSize] = e.stack[e.stackSize-1]
		e.ip++
		return nil
	case ADD:
		//
		e.stack[e.stackSize-1].Add(
			&(e.stack[e.stackSize]),
			&(e.stack[e.stackSize-1]),
		)
		e.stackSize--
		e.ip++
		return nil
	case EQ:
		//
		if e.stack[e.stackSize].Eq(&e.stack[e.stackSize-1]) {
			e.stack[e.stackSize-1].SetOne()
		} else {
			e.stack[e.stackSize-1].Clear()
		}
		e.stackSize--
		e.ip++
		return nil
	case JUMPDEST:
		// TODO validate pointer for jump instructions (or maybe it's already done?)
		e.ip = int(inst.Operand.Uint64())
		return nil
	case JUMPI:
		cond := e.stack[e.stackSize]
		e.stackSize--
		if !cond.IsZero() {
			e.ip = int(inst.Operand.Uint64())
			return nil
		}
		e.ip++
		return nil
	case INPUT:
		//EULER!! for debug only
		var i int
		fmt.Scanf("%d", &i)
		e.stackSize++
		e.stack[e.stackSize] = *uint256.NewInt(uint64(i))
		e.ip++
		return nil
	case PRINT:
		// EULER! for debug only [deprecated use native]
		num := e.stack[e.stackSize]
		fmt.Println(num.Uint64())
		e.ip++
		return nil
	case WRITESTR:
		// EULER! for debug only [deprecated use native]
		size := e.stack[e.stackSize].Uint64()
		offset := e.stack[e.stackSize-1].Uint64()
		fmt.Println(string(e.memory.store[offset:size]))
		e.ip++
		e.stackSize -= 2
		return nil
	// TODO in a future case MSTORE8:
	case NATIVE:
		e.ip++
		return e.execNative(inst.Operand.Uint64())
	case MSTORE256:
		offset := e.stack[e.stackSize-1].Uint64()
		val := e.stack[e.stackSize]
		e.memory.Set32(offset, val)
		e.stackSize -= 2
		e.ip++
		return nil
	case MLOAD:
		addr := e.stack[e.stackSize].Uint64()
		if addr > MemoryCapacity {
			return errInvalidMemoryAccess
		}

		e.stack[e.stackSize].SetBytes(e.memory.store[addr:32])
		e.ip++
		return nil
	case LT:
		x := e.stack[e.stackSize]
		e.stackSize--
		if x.Lt(&e.stack[e.stackSize]) {
			e.stack[e.stackSize].SetOne()
		} else {
			e.stack[e.stackSize].Clear()
		}
		e.ip++
		return nil
	case NOP:
		e.ip++
		return nil
	case NOT:
		if e.stack[e.stackSize].IsZero() {
			e.stack[e.stackSize][0] = 1
		} else {
			e.stack[e.stackSize].Clear()
		}
		e.ip++
		return nil
	case DROP:
		e.stackSize--
		e.ip++
		return nil
	case STOP:
		return stopToken
	}

	return errInvalidOpCodeCalled

}

func (e *EulVM) Reset() {
	e.ip = 0
	e.stackSize = 0
}

func (e *EulVM) Dump() {
	fmt.Println("-----stack-----")
	for i := 0; i <= e.stackSize; i++ {
		fmt.Println(e.stack[i])
	}
	fmt.Println("-----dump-----")
	e.memory.Dump()
}

// euler native functions
const (
	NativeWrite uint64 = iota + 1
)

func (e *EulVM) execNative(id uint64) error {
	switch id {
	case NativeWrite:
		size := e.stack[e.stackSize]
		addr := e.stack[e.stackSize-1]
		e.stackSize -= 2
		fmt.Println(string(e.memory.store[addr.Uint64() : addr.Uint64()+size.Uint64()]))
		return nil
	}

	return errUnknownNative
}
