package compiler

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Unheilbar/eulang/euvm"
)

const labelSfx = ":"
const labelLen = 8

type instruction struct {
	instruction euvm.OpCode
	operand     []byte
	label       string
}

func Compile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("can't open source file", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	counter := 0
	instructions := make([]instruction, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasSuffix(line, labelSfx) {
			labels[strings.TrimRight(line, labelSfx)] = counter
			continue
		}

		spline := strings.Split(line, " ")
		var i instruction
		opc, ok := euvm.OpCodesView[spline[0]]
		if !ok {
			log.Fatal("err unkown upcode ", spline[0])
		}
		i.instruction = opc
		if len(spline) > 1 {
			if strings.Contains(spline[0], "JUMP") {
				i.label = spline[1]
				counter += labelLen
			} else {
				i.operand = operandToBytes(spline[1])
			}
		}
		instructions = append(instructions, i)
		counter = counter + 1 + len(i.operand)
	}
	var code []byte
	for _, ins := range instructions {
		code = append(code, byte(ins.instruction))
		if ins.label != "" && ins.operand == nil {
			ins.operand = operandToBytes(fmt.Sprint(labels[ins.label]))
		}

		code = append(code, ins.operand...)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("failed scan file", err)
	}

	return code
}

// labels keep index of every label we found in the compiler
var labels = make(map[string]int, 0)

func translateAsm(asm []string) []byte {
	code := make([]byte, 0)
	for _, line := range asm {
		var inst instruction
		sl := strings.Split(line, " ")
		opc, ok := euvm.OpCodesView[sl[0]]
		if !ok {
			log.Fatal("err unkown upcode ", sl[0])
		}
		inst.instruction = opc
		code = append(code, byte(opc))
		if len(sl) > 1 {
			// replace label when we found it
			operand := sl[1]
			if val, ok := labels[sl[1]]; ok {
				operand = fmt.Sprint(val)
			}

			code = append(code, operandToBytes(operand)...)
		}
	}
	return code
}

func translateAsmLine(line string) []byte {
	code := make([]byte, 0)
	sl := strings.Split(line, " ")
	opc, ok := euvm.OpCodesView[sl[0]]
	if !ok {
		log.Fatal("err unkown upcode ", sl[0])
	}

	code = append(code, byte(opc))
	if len(sl) > 1 {
		// replace label when we found it
		operand := sl[1]
		if val, ok := labels[sl[1]]; ok {
			operand = fmt.Sprint(val)
		}

		code = append(code, operandToBytes(operand)...)
	}
	return code
}

func operandToBytes(op string) (ret []byte) {
	i, err := strconv.Atoi(op)

	if err != nil {
		log.Fatal("invalid operand", op)
	}
	ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(i))
	return
}
