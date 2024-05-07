package eulvm

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
	NOP
)

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
