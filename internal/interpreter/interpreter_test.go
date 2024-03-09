package interpreter

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
		{Instruction(0xABCD), 0, 0, 0xA},
		{Instruction(0xABCD), 1, 1, 0xB},
		{Instruction(0xABCD), 2, 2, 0xC},
		{Instruction(0xABCD), 3, 3, 0xD},
		{Instruction(0xABCD), 1, 2, 0xBC},
		{Instruction(0xABCD), 2, 3, 0xCD},
		{Instruction(0xABCD), 1, 3, 0xBCD},
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
