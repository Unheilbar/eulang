package eulvm

type OpCode byte

// 0x0 arithmetic range
const (
	STOP OpCode = iota
	ADD
	SUB
	POP
	PUSH
	SWAP
	DUP
	JUMPDEST
	JUMPI
	MSTORE8
	MSTORE256
	MLOAD
	MLOAD256
	DROP
	RET
	CALL
	CALLDATA
	DATALOAD
)

// 0x10 range - comparison ops.
const (
	LT OpCode = iota + 0x20
	GT
	EQ
	NOT
	NEQ
	AND
	OR
)

// 0x20 - debug
const (
	PRINT OpCode = iota + 0x30
	INPUT
	WRITESTR
	NATIVE
	NOP
)

// 0x40 -state
const (
	VSSTORE    OpCode = iota + 0x40 // store var to version storage
	VSLOAD                          // load var from version storage
	MAPVSSTORE                      // store map key into version storage
	MAPVSSLOAD                      // load map key from version storage
)

var OpCodesView = map[string]OpCode{
	"ADD":        ADD,
	"INPUT":      INPUT,
	"PRINT":      PRINT,
	"STOP":       STOP,
	"PUSH":       PUSH,
	"JUMPDEST":   JUMPDEST,
	"JUMPI":      JUMPI,
	"EQ":         EQ,
	"DUP":        DUP,
	"WRITESTR":   WRITESTR,
	"MSTORE8":    MSTORE8,
	"MSTORE256":  MSTORE256,
	"MLOAD":      MLOAD,
	"MLOAD256":   MLOAD256,
	"NATIVE":     NATIVE,
	"NOT":        NOT,
	"LT":         LT,
	"GT":         GT,
	"DROP":       DROP,
	"RET":        RET,
	"CALLDATA":   CALLDATA,
	"DATALOAD":   DATALOAD,
	"SWAP":       SWAP,
	"SUB":        SUB,
	"NEQ":        NEQ,
	"AND":        AND,
	"OR":         OR,
	"VSSTORE":    VSSTORE,
	"VSLOAD":     VSLOAD,
	"MAPVSSTORE": MAPVSSTORE,
	"MAPVSSLOAD": MAPVSSLOAD,
}

var OpCodes = map[OpCode]string{
	ADD:        "ADD",
	INPUT:      "INPUT",
	PRINT:      "PRINT",
	STOP:       "STOP",
	PUSH:       "PUSH",
	JUMPDEST:   "JUMPDEST",
	JUMPI:      "JUMPI",
	EQ:         "EQ",
	DUP:        "DUP",
	WRITESTR:   "WRITESTR",
	MSTORE8:    "MSTORE8",
	MSTORE256:  "MSTORE256",
	MLOAD:      "MLOAD",
	MLOAD256:   "MLOAD256",
	NATIVE:     "NATIVE",
	NOT:        "NOT",
	LT:         "LT",
	DROP:       "DROP",
	CALL:       "CALL",
	CALLDATA:   "CALLDATA",
	DATALOAD:   "DATALOAD",
	RET:        "RET",
	SWAP:       "SWAP",
	SUB:        "SUB",
	NEQ:        "NEQ",
	GT:         "GT",
	AND:        "AND",
	OR:         "OR",
	VSSTORE:    "VSTORE",
	VSLOAD:     "VLOAD",
	MAPVSSTORE: "MAPVSSTORE",
	MAPVSSLOAD: "MAPVSSLOAD",
}

func checkOpCodes() {}
