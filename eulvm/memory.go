package eulvm

import (
	"fmt"

	"github.com/holiman/uint256"
)

const MemoryCapacity = 100 * 1024 // memory limit to avoid extra allocations during execution

type Memory struct {
	store [MemoryCapacity]byte //
	size  uint64
}

func NewMemory() *Memory {
	return &Memory{}
}

func NewMemoryWithPrealloc(prealloc []byte) *Memory {
	m := &Memory{}
	copy(m.store[:], prealloc)
	m.size = uint64(len(prealloc))
	return m
}

func (m *Memory) Set32(offset uint64, val uint256.Int) {
	// length of store may never be less than offset + size.
	// The store should be resized PRIOR to setting the memory
	if offset+32 > uint64(MemoryCapacity) {
		panic("invalid memory: store empty")
	}
	// Zero the memory area
	copy(m.store[offset:offset+32], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	// Fill in relevant bits
	val.WriteToSlice(m.store[offset:])
	m.size += 32
}

func (m *Memory) Size() uint64 {
	return m.size
}

func (m *Memory) Store() []byte {
	res := make([]byte, m.size)
	copy(res[:], m.store[:m.size])
	return res
}

func (m *Memory) Dump() {
	fmt.Println("allocated size:", m.size)
	fmt.Println("===memory dump===")
	fmt.Println(m.store[:m.size])
	fmt.Println("=end memory dump=")
}
