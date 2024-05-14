package eulvm

type OpCode byte

// 0x0 arithmetic range
const (
	STOP OpCode = iota
	ADD
	SUB
	POP
	PUSH
	DUP
	JUMPDEST
	JUMPI
	MSTORE8
	MSTORE256
	MLOAD
	MLOAD256
	DROP
)

// 0x10 range - comparison ops.
const (
	LT OpCode = iota + 0x10
	GT
	EQ
	NOT
)

// 0x20 - debug
const (
	PRINT OpCode = iota + 0x20
	INPUT
	WRITESTR
	NATIVE
	NOP
)

var OpCodesView = map[string]OpCode{
	"ADD":       ADD,
	"INPUT":     INPUT,
	"PRINT":     PRINT,
	"STOP":      STOP,
	"PUSH":      PUSH,
	"JUMPDEST":  JUMPDEST,
	"JUMPI":     JUMPI,
	"EQ":        EQ,
	"DUP":       DUP,
	"WRITESTR":  WRITESTR,
	"MSTORE8":   MSTORE8,
	"MSTORE256": MSTORE256,
	"MLOAD":     MLOAD,
	"MLOAD256":  MLOAD256,
	"NATIVE":    NATIVE,
	"NOT":       NOT,
	"LT":        LT,
	"DROP":      DROP,
}

var OpCodes = map[OpCode]string{
	ADD:       "ADD",
	INPUT:     "INPUT",
	PRINT:     "PRINT",
	STOP:      "STOP",
	PUSH:      "PUSH",
	JUMPDEST:  "JUMPDEST",
	JUMPI:     "JUMPI",
	EQ:        "EQ",
	DUP:       "DUP",
	WRITESTR:  "WRITESTR",
	MSTORE8:   "MSTORE8",
	MSTORE256: "MSTORE256",
	MLOAD:     "MLOAD",
	MLOAD256:  "MLOAD256",
	NATIVE:    "NATIVE",
	NOT:       "NOT",
	LT:        "LT",
	DROP:      "DROP",
}
