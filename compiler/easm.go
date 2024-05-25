package compiler

import (
	"log"

	"github.com/Unheilbar/eulang/eulvm"
	"github.com/Unheilbar/eulang/utils"
	"github.com/holiman/uint256"
)

// easm translates expressions from ast to assembly instructions
type easm struct {
	program eulvm.Program
	memory  *eulvm.Memory
}

func NewEasm() *easm {
	return &easm{
		memory: eulvm.NewMemory(),
	}
}

func (e *easm) translateSource(filepath string, program []eulvm.Program) {
	panic("easm translate source not implemented")
}

// returns the address of the word in the memory
func (e *easm) pushStringToMemory(str string) eulvm.Word {
	return e.pushByteArrToMemory([]byte(str))
}

func (e *easm) pushByteArrToMemory(arr []byte) eulvm.Word {
	memSize := e.memory.Size()
	words := arrToWords(arr)

	if int(memSize)+len(words)*32 > eulvm.MemoryCapacity {
		log.Fatal("memory limit exceeded. Increase memory limit in virtual machine")
	}

	result := *uint256.NewInt(uint64(memSize))

	for _, word := range words {
		e.memory.Set32(uint64(e.memory.Size()), word)
	}

	return result
}

func (e *easm) pushWordToMemory(w eulvm.Word) eulvm.Word {
	memSize := e.memory.Size()
	result := *uint256.NewInt(uint64(memSize))

	e.memory.Set32(uint64(e.memory.Size()), w)

	return result
}

func (e *easm) pushBufferToMemory(buffer []byte) eulvm.Word {
	result := uint256.NewInt(uint64(e.memory.Size()))
	e.memory.Set(e.memory.Size(), uint64(len(buffer)), buffer)

	return *result
}

// returns instruction address
func (e *easm) pushInstruction(i eulvm.Instruction) int {
	//TODO euler do we need program capacity?
	return e.program.PushInstruction(i)
}

// TODO later shouldn't be public
func (e *easm) PushInstruction(i eulvm.Instruction) int {
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
		var b [32]byte
		copy(b[:], []byte(str))
		word.SetBytes32(b[:])
		words = append(words, word)
	}

	return words
}

func arrToWords(arr []byte) []eulvm.Word {
	words := make([]eulvm.Word, 0)
	offset := 0
	step := 32
	for offset+step < len(arr) {
		var word eulvm.Word
		word.SetBytes(arr[offset : offset+step])
		offset += step
		words = append(words, word)
	}
	if offset < len(arr) {
		var word eulvm.Word
		var b [32]byte
		copy(b[:], arr[offset:])
		word.SetBytes32(b[:])
		words = append(words, word)
	}

	return words
}

func (e *easm) dumpProgramToFile(filepath string) {
	e.program.PreallocMemory = e.memory.Store()
	utils.DumpProgramIntoFile(filepath, e.program)
}

func (e *easm) GetProgram() eulvm.Program {
	e.program.PreallocMemory = e.memory.Store()
	return e.program
}
