package eulvm

import (
	"log"
	"testing"

	"github.com/holiman/uint256"
)

func Test_exec(t *testing.T) {
	testProg := []Instruction{
		makeInstruction(PUSH, 1),
		makeInstruction(PUSH, 10),
		makeInstruction(ADD, -1),
		makeInstruction(PRINT, -1),
		makeInstruction(STOP, -1),
	}

	euler := New(testProg)
	err := euler.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func Benchmark_exec(b *testing.B) {
	b.StopTimer()
	testProg := []Instruction{
		makeInstruction(PUSH, 1),
		makeInstruction(PUSH, 10),
		makeInstruction(ADD, -1),
		makeInstruction(STOP, -1),
	}

	euler := New(testProg)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		euler.Run()
		euler.Reset()
	}
}

func makeInstruction(op OpCode, operand int) Instruction {
	var ins Instruction
	ins.OpCode = op
	var integer = new(uint256.Int)
	ins.Operand = *integer.SetUint64(uint64(operand))
	return ins
}
