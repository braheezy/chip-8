package main

import (
	"fmt"
	"testing"
)

func TestNibbles(t *testing.T) {
	tests := []struct {
		instruction Instruction
		start       int
		end         int
		expected    uint16
	}{
		{Instruction(0xABCD), 2, 2, 0xB},
		{Instruction(0xABCD), 3, 3, 0xC},
		{Instruction(0xABCD), 4, 4, 0xD},
		{Instruction(0xABCD), 2, 3, 0xBC},
		{Instruction(0xABCD), 3, 4, 0xCD},
		{Instruction(0xABCD), 2, 4, 0xBCD},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Instruction %X [%d-%d]", test.instruction, test.start, test.end), func(t *testing.T) {
			result := test.instruction.nibbles(test.start, test.end)
			if result != test.expected {
				t.Errorf("Expected: %X, Got: %X", test.expected, result)
			}
		})
	}
}
