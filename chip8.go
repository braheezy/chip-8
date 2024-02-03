package main

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

	// // CPU registers.
	// v [16]byte

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

func (ch8 *CHIP8) readNextInstruction() (uint8, uint8) {
	// Read next instruction from memory.
	firstNib := ch8.memory[ch8.pc] >> 4
	secondNib := ch8.memory[ch8.pc] & 0x0F
	ch8.pc += 2
	return firstNib, secondNib
}
