package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Unheilbar/eulang/compiler"
)

func main() {
	for _, arg := range os.Args[1:] {
		code := compiler.Compile(arg)
		dump(fmt.Sprint(arg, ".bin"), code)
	}
}

func dump(path string, code []byte) {
	if err := os.WriteFile(path, code, 0666); err != nil {
		log.Fatal(err)
	}
}
