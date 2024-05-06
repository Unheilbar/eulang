package eulvm

import (
	"fmt"
)

// TODO probably stack should be with unsafe pointer + stack size
type Stack struct {
	data []Word
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]Word, 0, size),
	}
}

func (st *Stack) pop() (ret Word) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *Stack) peek() (ret Word) {
	ret = st.data[len(st.data)-1]
	return
}

func (st *Stack) push(d Word) {
	st.data = append(st.data, d) // TODO try pointer in a future. for now see how it works
}

func (st *Stack) dup(n int) {
	st.push(st.data[len(st.data)-n])
}

func (st *Stack) Dump() {
	fmt.Println("### stack ###")
	if len(st.data) > 0 {
		for i, val := range st.data {
			fmt.Printf("%-3d  %v\n", i, val)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("#############")
}
