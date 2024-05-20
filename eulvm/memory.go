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
	if offset+32 > m.size {
		m.size = offset + 32
	}
}

// Set sets offset + size to value
func (m *Memory) Set(offset, size uint64, value []byte) {
	// It's possible the offset is greater than 0 and size equals 0. This is because
	// the calcMemSize (common.go) could potentially return 0 when size is zero (NO-OP)
	if size > 0 {

		//TODO write proper check for memory capacity
		// length of store may never be less than offset + size.
		// The store should be resized PRIOR to setting the memory
		if offset+size > uint64(len(m.store)) {
			panic("invalid memory: store empty")
		}
		copy(m.store[offset:offset+size], value)
	}

	if offset+size > m.size {
		m.size = offset + size
	}
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
