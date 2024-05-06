package compiler

import (
	"encoding/binary"
	"io"
	"log"
	"os"

	"github.com/Unheilbar/eulang/euvm"
)

const labelSfx = ":"
const labelLen = 8

type instruction struct {
	Instruction euvm.OpCode
	Operand     [32]byte
}

func dumpProgramIntoFile(filename string, program []instruction) (err error) {
	fi, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal("can't close file", err)
		}
	}()

	err = binary.Write(fi, binary.LittleEndian, program)
	if err != nil {
		log.Fatal("cant write to file", err)
	}

	return err
}

func loadProgramFromFile(filename string) ([]instruction, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal("can't close file", err)
		}
	}()

	var program []instruction

	for {
		instruction := instruction{}
		err := binary.Read(fi, binary.LittleEndian, &instruction)

		if err == io.EOF {
			break
		}

		program = append(program, instruction)
	}
	return program, err
}
