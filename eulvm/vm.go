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
}

const limit = 1024

func New(prog []Instruction) *EulVM {
	return &EulVM{
		program: prog,
	}
}

func (e *EulVM) Run() error {
	for i := 0; i < limit; i++ {
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
)

var stopToken = errors.New("program stopped")

func executeNext(e *EulVM) error {
	if e.ip >= len(e.program) {
		return errIllegalCall
	}

	inst := e.program[e.ip]
	//fmt.Println("call ->", OpCodes[inst.OpCode])
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
		// EULER! for debug only
		num := e.stack[e.stackSize]
		fmt.Println(num.Uint64())
		e.ip++
		return nil
	case NOP:
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
	for i := 0; i < e.stackSize; i++ {
		fmt.Println(e.stack[i])
	}
	fmt.Println("-----dump-----")
}