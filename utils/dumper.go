package utils

import (
	"encoding/binary"
	"io"
	"log"
	"os"

	"github.com/Unheilbar/eulang/eulvm"
)

//TODO it's just a prove of conception. euler later come up with a better package name

func DumpProgramIntoFile(filename string, program []eulvm.Instruction) (err error) {
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

func LoadProgramFromFile(filename string) ([]eulvm.Instruction, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal("can't close file", err)
		}
	}()

	var program []eulvm.Instruction

	for {
		instruction := eulvm.Instruction{}
		err := binary.Read(fi, binary.LittleEndian, &instruction)

		if err == io.EOF {
			break
		}

		program = append(program, instruction)
	}
	return program, err
}
