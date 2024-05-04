package euvm

type OpCode byte

// 0x0 arithmetic range
const (
	STOP OpCode = iota
	ADD
	SUB
	POP
	PUSH
	JUMPDEST
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
	INPUT
)

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
	case INPUT:
		return opInput
	case JUMPDEST:
		return opJumpdest
	case STOP:
		return opStop
	}

	panic("unkown opcode")
}

var OpCodesView = map[string]OpCode{
	"ADD":      ADD,
	"INPUT":    INPUT,
	"PRINT":    PRINT,
	"STOP":     STOP,
	"PUSH":     PUSH,
	"JUMPDEST": JUMPDEST,
}

var OpCodes = map[OpCode]string{
	ADD:      "ADD",
	INPUT:    "INPUT",
	PRINT:    "PRINT",
	STOP:     "STOP",
	PUSH:     "PUSH",
	JUMPDEST: "JUMPDEST",
}
