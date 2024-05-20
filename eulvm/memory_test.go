package eulvm

import (
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func Test__Memory(t *testing.T) {
	m := NewMemory()

	expVal := uint256.NewInt(10)

	m.Set32(0, *expVal)

	var actualVal = new(uint256.Int)

	actualVal.SetBytes32(m.store[0:32])

	assert.Equal(t, expVal, actualVal)
}
