package eulvm

import (
	"errors"
	"fmt"
	"hash"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

const StackCapacity = 33

type keccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// input can be accessed by Operations to set program entry point
type EulVM struct {
	program []Instruction //TODO make unsafe pointer to avoid program size check?

	input []byte

	state map[common.Hash]common.Hash // TODO later use actual stateDB as storage backend. map can be used for temporary map storage inside of smart contract

	ip int

	stack     [StackCapacity]Word
	stackSize int
	memory    *Memory

	hasher       keccakState // Keccak256 hasher instance shared across opcodes
	hasherBuf    common.Hash // Keccak256 hasher result array shared aross opcodes
	mapKeyBuffer [64]byte

	debug bool
}

const ExecutionLimit = 1024

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
		state:   make(map[common.Hash]common.Hash),
		hasher:  sha3.NewLegacyKeccak256().(keccakState),
	}
}

func (e *EulVM) WithDebug() *EulVM {
	e.debug = true
	return e
}

func (e *EulVM) Run(input []byte) error {
	e.input = input

	for i := 0; i < ExecutionLimit; i++ {
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

var debugCounter int
var breakPoint int

func executeNext(e *EulVM) error {
	if e.ip >= len(e.program) {
		return errIllegalCall
	}

	inst := e.program[e.ip]
	if e.debug {
		debugCounter++
		if debugCounter < breakPoint {
			goto exec
		}
		var command string
		var operand int
		fmt.Scanf("%s %d", &command, &operand)
		command = strings.TrimSpace(command)
		switch command {
		case "help":
			fmt.Println(
				` debug mode for evm commands:
			  stack - dump current stack state
			  memory - dump current memory state
			  next_op or ''- show next command for execution
			  break - go to break point of debuger
			`)
			debugCounter--
			return nil
		case "stack":
			e.Dump()
			debugCounter--
			return nil
		case "memory":
			e.memory.Dump()
			debugCounter--
			return nil
		case "", "next_op":
		case "break":
			debugCounter--
			breakPoint = operand
		default:
			fmt.Println("use help to get commands info")
			return nil
		}
		fmt.Println("debug point", debugCounter, "ip:", e.ip, "-->call:",
			OpCodes[inst.OpCode],
			"operand:", inst.Operand.Uint64())
	}
exec:
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
	case NEQ:
		if !e.stack[e.stackSize].Eq(&e.stack[e.stackSize-1]) {
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

		e.stack[e.stackSize].SetBytes(e.memory.store[addr : addr+32])
		e.ip++
		return nil
	case VSSTORE:
		val := e.stack[e.stackSize]
		key := e.stack[e.stackSize-1]
		e.stackSize -= 2
		e.state[key.Bytes32()] = val.Bytes32()
		e.ip++
		return nil
	case VSLOAD:
		key := e.stack[e.stackSize].Bytes32()
		e.stack[e.stackSize].SetBytes(e.state[key].Bytes())
		e.ip++
		return nil
	case MAPVSSTORE:
		val := e.stack[e.stackSize]
		key := e.stack[e.stackSize-1].Bytes()
		copy(e.mapKeyBuffer[:32], key)
		copy(e.mapKeyBuffer[32:], inst.Operand.Bytes())

		e.hasher.Reset()
		e.hasher.Write(e.mapKeyBuffer[:])
		e.hasher.Read(e.hasherBuf[:])

		e.state[e.hasherBuf] = val.Bytes32()

		e.stackSize -= 2
		e.ip++
		return nil
	case MAPVSSLOAD:
		key := e.stack[e.stackSize].Bytes()
		copy(e.mapKeyBuffer[:32], key)
		copy(e.mapKeyBuffer[32:], inst.Operand.Bytes())

		e.hasher.Reset()
		e.hasher.Write(e.mapKeyBuffer[:])
		e.hasher.Read(e.hasherBuf[:])

		e.stack[e.stackSize].SetBytes(e.state[e.hasherBuf].Bytes())
		e.ip++
		return nil
	case LT:
		x := e.stack[e.stackSize-1]
		y := e.stack[e.stackSize]
		e.stackSize--
		if x.Lt(&y) {
			e.stack[e.stackSize].SetOne()
		} else {
			e.stack[e.stackSize].Clear()
		}
		e.ip++
		return nil
	case GT:
		x := e.stack[e.stackSize-1]
		y := e.stack[e.stackSize]
		e.stackSize--
		if x.Gt(&y) {
			e.stack[e.stackSize].SetOne()
		} else {
			e.stack[e.stackSize].Clear()
		}
		e.ip++
		return nil
	case SUB:
		x := e.stack[e.stackSize]
		y := e.stack[e.stackSize-1]
		e.stackSize--
		e.stack[e.stackSize] = *y.Sub(&y, &x)
		e.ip++
		return nil
	case AND:
		x := e.stack[e.stackSize]
		y := e.stack[e.stackSize-1]
		e.stackSize--
		e.stack[e.stackSize] = *y.And(&y, &x)
		e.ip++
		return nil
	case OR:
		x := e.stack[e.stackSize]
		y := e.stack[e.stackSize-1]
		e.stackSize--
		e.stack[e.stackSize] = *y.Or(&y, &x)
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
	case CALL:
		e.stackSize += 1
		e.stack[e.stackSize] = *uint256.NewInt(uint64(e.ip + 1)) //set return address of the call
		e.ip = int(inst.Operand.Uint64())                        //ip jumps to function
		return nil
	case RET:
		e.ip = int(e.stack[e.stackSize].Uint64())
		e.stackSize--
		return nil
	case CALLDATA:
		//TODO later implement load of call parameters
		var addr uint256.Int
		addr.SetBytes(e.input[:32])
		e.stackSize++
		e.stack[e.stackSize] = *uint256.NewInt(uint64(e.ip + 1)) // ip of return statement is next instruction

		e.ip = int(addr.Uint64()) // set instruction pointer to entry function
		return nil
	case DATALOAD:
		//TODO boundary cheks
		from := e.stack[e.stackSize].Uint64()
		val := e.input[from : from+WordLength.Uint64()]
		e.stack[e.stackSize].SetBytes(val)
		e.ip++
		return nil
	case SWAP:
		//TODO add stack overflow check
		a := e.stackSize
		b := e.stackSize - int(inst.Operand.Uint64())
		e.stack[a], e.stack[b] = e.stack[b], e.stack[a]
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
	fmt.Println("stack size:", e.stackSize)
	fmt.Println("-----stack-----")
	for i := 0; i <= e.stackSize; i++ {
		fmt.Println(e.stack[i])
	}
	fmt.Println("-----dump-----")
}

// euler native functions
const (
	NativeWrite uint64 = iota + 1
	NativeWriteF
)

func (e *EulVM) popInt() int {
	ret := e.stack[e.stackSize].Uint64()
	e.stackSize--
	return int(ret)
}

func (e *EulVM) popHash() common.Hash {
	ret := e.stack[e.stackSize]
	e.stackSize--
	return common.BytesToHash(ret.Bytes())
}

func (e *EulVM) popStr() string {
	size := e.stack[e.stackSize]
	addr := e.stack[e.stackSize-1]
	e.stackSize -= 2
	return string(e.memory.store[addr.Uint64() : addr.Uint64()+size.Uint64()])
}

func (e *EulVM) popAddr() common.Address {
	ret := e.stack[e.stackSize]
	e.stackSize--
	return common.BytesToAddress(ret.Bytes())
}

func (e *EulVM) execNative(id uint64) error {
	switch id {
	case NativeWrite:
		size := e.stack[e.stackSize]
		addr := e.stack[e.stackSize-1]
		e.stackSize -= 2
		fmt.Print(string(e.memory.store[addr.Uint64() : addr.Uint64()+size.Uint64()]))
		return nil
	case NativeWriteF:
		// TODO add stack overflow check
		var args []interface{}

		frmtStr := e.popStr()
		clone := strings.Clone(frmtStr)
		for clone != chopFrom(clone, isPercent) {
			clone = chopFrom(clone, isPercent)
			if strings.HasPrefix(clone, "%d") {
				args = append(args, e.popInt())
				clone = strings.TrimPrefix(clone, "%d")
			} else if strings.HasPrefix(clone, "%s") {
				args = append(args, e.popStr())
				clone = strings.TrimPrefix(clone, "%s")
				// NOTE add here future formattings
			} else if strings.HasPrefix(clone, "%v") {
				args = append(args, e.popHash())
				clone = strings.TrimPrefix(clone, "%v")
			} else if strings.HasPrefix(clone, "%x") {
				args = append(args, e.popAddr())
				clone = strings.TrimPrefix(clone, "%x")
			} else {
				clone = strings.TrimPrefix(clone, "%")
			}
		}

		fmt.Printf(frmtStr, args...)
		return nil
	}

	return errUnknownNative
}

func isPercent(r rune) bool {
	return r != '%'
}

func chopFrom(s string, until func(r rune) bool) string {
	for i, r := range s {
		if !until(r) {
			return s[i:]
		}
	}
	return s
}
