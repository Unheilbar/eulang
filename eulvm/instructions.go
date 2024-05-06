package eulvm

import "github.com/holiman/uint256"

type Word = uint256.Int

type Instruction struct {
	OpCode  OpCode
	Operand Word
}
