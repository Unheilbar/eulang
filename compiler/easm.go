package compiler

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/holiman/uint256"
)

// easm represents a translation context of the compiler
type easm struct {
	program eulvm.Program
	memory  *eulvm.Memory
}

func newEasm() *easm {
	return &easm{
		memory: eulvm.NewMemory(),
	}
}

func (e *easm) translateSource(filepath string, program []eulvm.Program) {
	panic("easm translate source not implemented")
}

// returns the address of the word in the memory
func (e *easm) pushStringToMemory(str string) eulvm.Word {
	var result eulvm.Word

	memSize := e.memory.Size()
	words := strToWords(str)

	if int(memSize)+len(words)*32 > eulvm.MemoryCapacity {
		log.Fatal("memory limit exceeded. Increase memory limit in virtual machine")
	}

	result = *uint256.NewInt(uint64(memSize))

	for _, word := range words {
		e.memory.Set32(uint64(e.memory.Size()), word)
	}

	return result
}

// returns instruction address
func (e *easm) pushInstruction(i eulvm.Instruction) int {
	//TODO euler do we need program capacity?
	return e.program.PushInstruction(i)
}

func strToWords(str string) []eulvm.Word {
	words := make([]eulvm.Word, 0)
	offset := 0
	step := 32
	for offset+step < len(str) {
		var word eulvm.Word
		word.SetBytes([]byte(str[offset:step]))
		offset += step
		words = append(words, word)
	}
	if offset < len(str) {
		var word eulvm.Word
		word.SetBytes([]byte(str[offset:]))
		words = append(words, word)
	}

	return words
}

func (e *easm) dumpProgramToFile(filepath string) {

}

// [DEPRECATED]
const labelSfx = ":"

func CompileEasmFromFile(filename string, outputpath string) []eulvm.Instruction {
	var labels = make(map[string]int, 0)
	var unresolvedInst = make(map[int]string, 0)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("can't open source file", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	instructions := make([]eulvm.Instruction, 0)

	for scanner.Scan() {
		var inst eulvm.Instruction
		line := strings.TrimSpace(scanner.Text())
		if strings.HasSuffix(line, labelSfx) {
			labels[strings.TrimRight(line, labelSfx)] = len(instructions)
			continue
		}

		spline := strings.Split(line, " ")
		opc, ok := eulvm.OpCodesView[spline[0]]
		if !ok {
			log.Fatal("err unkown upcode ", spline[0])
		}
		inst.OpCode = opc
		if len(spline) > 1 {
			if strings.Contains(spline[0], "JUMP") {
				idx, err := strconv.Atoi(spline[1])
				inst.Operand = *uint256.NewInt(uint64(idx))
				if err != nil {
					unresolvedInst[len(instructions)] = spline[1]
				}
			} else {
				//TODO later implement func getOperand(opCode OpCode, operand string). Because operands representation depends on the opcodes
				op, err := strconv.Atoi(spline[1])
				if err != nil {
					log.Fatal("illegal operand for opcode", inst.OpCode)
				}

				inst.Operand = *uint256.NewInt(uint64(op))
			}
		}
		instructions = append(instructions, inst)
	}

	for idx, label := range unresolvedInst {
		jumpIdx, ok := labels[label]
		if !ok {
			log.Fatalf("label %s can't be resolved", label)
		}

		instructions[idx].Operand = *uint256.NewInt(uint64(jumpIdx))
	}

	return instructions
}
