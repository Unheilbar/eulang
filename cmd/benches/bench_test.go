package main

import (
	"testing"
)

func Benchmark__DemoTransfer(b *testing.B) {
	b.StopTimer()

	bs := prepareBench()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		bs.vm.Run(bs.transferCall)
		bs.vm.Reset() //reset stack/memory
	}
}
