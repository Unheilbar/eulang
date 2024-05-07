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
