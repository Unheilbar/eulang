package eul

import (
	"encoding/binary"
	"testing"
)

func Test_Run(t *testing.T) {
	interpreter := Interpreter{}
	a := make([]byte, 8)
	b := make([]byte, 8)

	binary.LittleEndian.PutUint64(a, 25)
	binary.LittleEndian.PutUint64(b, 32)
	code := []byte{}
	code = append(code, byte(PUSH))
	code = append(code, a...)
	code = append(code, byte(PUSH))
	code = append(code, b...)
	code = append(code, byte(ADD))
	code = append(code, byte(PRINT))
	code = append(code, byte(STOP))
	interpreter.Run(code)
}
