package eulvm

import "fmt"

// TODO eulang program is a prototype for contract in euler
type Program struct {
	Instrutions []Instruction

	PreallocMemory []byte
}

func NewProgram(instrs []Instruction, preallocMemory []byte) Program {
	return Program{
		Instrutions:    instrs,
		PreallocMemory: preallocMemory,
	}
}

func (p *Program) PushInstruction(i Instruction) int {
	//TODO do we need program capacity here?
	p.Instrutions = append(p.Instrutions, i)
	return len(p.Instrutions) - 1
}

func (p *Program) Size() int {
	return len(p.Instrutions)
}

func (p *Program) Dump() {
	fmt.Println("=====instructions dump=====")
	for i, inst := range p.Instrutions {
		fmt.Println(i, OpCodes[inst.OpCode], inst.Operand)
	}
	fmt.Println("===end instructions dump===")
}
