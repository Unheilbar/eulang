package utils

import (
	"os"
	"testing"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func Test_DumpLoad(t *testing.T) {
	var filename = "test.o"

	instructions := []eulvm.Instruction{
		{OpCode: eulvm.ADD, Operand: *uint256.NewInt(10)},
		{OpCode: eulvm.SUB, Operand: *uint256.NewInt(10)},
	}

	program := eulvm.NewProgram(instructions, []byte{1, 2, 3})

	err := DumpProgramIntoFile(filename, program)

	assert.NoError(t, err)

	prog, err := LoadProgramFromFile(filename)

	assert.NoError(t, err)

	assert.Equal(t, program, prog)

	os.Remove(filename)
}
