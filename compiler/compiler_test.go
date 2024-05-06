package compiler

import (
	"testing"

	"github.com/Unheilbar/eulang/euvm"
	"github.com/stretchr/testify/assert"
)

func Test_Save(t *testing.T) {
	var filename = "test.o"

	program := []instruction{
		{Instruction: euvm.ADD, Operand: [32]byte{1, 2, 3}},
		{Instruction: euvm.SUB, Operand: [32]byte{1, 2, 3}},
	}

	err := dumpProgramIntoFile(filename, program)

	assert.NoError(t, err)

	prog, err := loadProgramFromFile(filename)

	assert.NoError(t, err)

	assert.Equal(t, program, prog)

}
