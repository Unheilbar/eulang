package euvm

import (
	"testing"
)

func Test_Run(t *testing.T) {
	interpreter := Interpreter{}
	code := []byte{}
	code = append(code, byte(INPUT))
	code = append(code, byte(INPUT))
	code = append(code, byte(ADD))
	code = append(code, byte(PRINT))
	code = append(code, byte(STOP))
	interpreter.Run(code)
}
