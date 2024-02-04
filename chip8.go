package main

import (
	"fmt"
)

const (
	// In CHIP-8, the program starts at address 0x200.
	programStartAddress = 0x200
)

type CHIP8 struct {
	// Define 4k of RAM.
	memory [4096]byte

	// Current instruction in memory to execute.
	pc uint16

	// // Index register, holding addresses in memory.
	// i uint16

	// CPU registers.
	v [16]byte

	// // Delay timer. Decremented at 60Hz until 0.
	// dt byte

	// // Sound timer. Decremented at 60Hz until 0. Emits beep while not 0.
	// st byte

	// // The execution stack for subroutines
	// stack Stack

	// The current display of the program
	display Display

	// Size of program being executed.
	programSize int
}

func NewCHIP8() *CHIP8 {
	chip8 := &CHIP8{
		pc: programStartAddress,
	}
	chip8.display.onColor = Pine.RGBA()
	chip8.display.offColor = Gold.RGBA()
	return chip8
}

func (ch8 *CHIP8) readNextInstruction() uint16 {
	// Read next instruction from memory.
	first := ch8.memory[ch8.pc]
	second := ch8.memory[ch8.pc+1]
	ch8.pc += 2
	return uint16(first)<<8 | uint16(second)
}

func (ch8 *CHIP8) process() {
	// Execute instructions
	for int(ch8.pc) < ch8.programSize {
		instruction := ch8.readNextInstruction()

		firstNibble := instruction >> 12

		switch firstNibble {
		case 0x0:
			secondNibble := instruction >> 8
			if secondNibble == 0x0 {
				lastByte := instruction & 0x00FF
				if lastByte == 0xE0 {
					// CLS: Clear the display
					println("CLS found (00E0)")
					ch8.display.clear()
				} else if lastByte == 0xEE {
					// RET: Return from a subroutine.
					println("RET found (00EE), but not implemented yet")
				}
			} else {
				// SYS addr: Jump to a machine code routine.
				fmt.Printf("SYS addr not supported!: %04X\n", instruction)
			}
		case 0x6:
			// LD Vx, byte
			// Store number NN in register VX.
			value := (instruction >> 4) & 0x0F
			register := instruction & 0x0F
			fmt.Printf("Loading %d into register %d\n", value, register)
			ch8.v[register] = byte(value)

		default:
			fmt.Printf("Unsupported instruction: %04X\n", instruction)
		}
	}
}
