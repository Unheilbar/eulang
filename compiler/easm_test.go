package compiler

import (
	"fmt"
	"testing"

	"github.com/Unheilbar/eulang/eulvm"
)

func Test_CompileEasmFromFile(t *testing.T) {
	program := CompileEasmFromFile("../examples/summator.easm", "")

	for _, inst := range program {
		fmt.Println(eulvm.OpCodes[inst.OpCode], inst.Operand.Uint64())
	}
}
