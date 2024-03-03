package chip8

import "errors"

type Stack []uint16

// Function to manage the stack.
func (s *Stack) Push(v uint16) {
	*s = append(*s, v)
}
func (s *Stack) Pop() (uint16, error) {
	l := len(*s)
	if l == 0 {
		return 0, errors.New("empty stack")
	}
	v := (*s)[l-1]
	*s = (*s)[:l-1]
	return v, nil
}
