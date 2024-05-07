package eulvm

import "github.com/holiman/uint256"

// EULER probably it's better to implement something like a C union for a word.
// In order to increase our perfomance we need to use a stronger type system than that
// TODO test later something like
/*type Word2 struct {
	int64 int64
	uint256.Int
	...
}*/
type Word = uint256.Int

type Instruction struct {
	OpCode  OpCode
	Operand Word
}
