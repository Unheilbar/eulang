package main

import (
	"testing"
)

func Benchmark__DemoTransfer(b *testing.B) {
	bs := prepareBench()
	for i := 0; i < b.N; i++ {
		bs.vm.Run(bs.transferCall)
		bs.vm.Reset() //reset stack/memory
	}
	expectedAmount := b.N
	bs.check(expectedAmount)
	expectedAmount += b.N
}
