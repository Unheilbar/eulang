package utils

import (
	"encoding/gob"
	"log"
	"os"

	"github.com/Unheilbar/eulang/eulvm"
)

//TODO it's just a prove of conception. eulang later come up with a better package name

func DumpProgramIntoFile(filename string, program eulvm.Program) (err error) {
	fi, err := os.Create(filename)
	if err != nil {
		return err
	}
	gob.Register(program)
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal("can't close file", err)
		}
	}()

	err = gob.NewEncoder(fi).Encode(program)
	if err != nil {
		log.Fatalf("can't encode program into file %s err %s", filename, err)
	}

	return nil
}

func LoadProgramFromFile(filename string) (eulvm.Program, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return eulvm.Program{}, err
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatal("can't close file", err)
		}
	}()

	var program eulvm.Program

	err = gob.NewDecoder(fi).Decode(&program)

	if err != nil {
		log.Fatalf("can't decode file %s err %s", filename, err)
	}

	return program, nil
}
