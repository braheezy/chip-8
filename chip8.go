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

	// Index register, holding addresses in memory.
	I uint16

	// CPU registers.
	V [16]byte

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

	currentCycle++
	return uint16(first)<<8 | uint16(second)
}

func (ch8 *CHIP8) stepInterpreter() {
	instruction := ch8.readNextInstruction()

	firstNibble := instruction >> 12

	switch firstNibble {
	case 0x0:
		secondNibble := (instruction >> 8) & 0xF
		if secondNibble == 0x0 {
			lastByte := instruction & 0x00FF
			if lastByte == 0xE0 {
				// CLS: Clear the display
				debug("[%04X] CLS\n", instruction)
				ch8.display.clear()
			} else if lastByte == 0xEE {
				// RET: Return from a subroutine.
				println("RET found (00EE), but not implemented yet")
			}
		} else {
			// SYS addr: Jump to a machine code routine.
			debug("[%04X] SYS addr not supported!\n", instruction)
		}
	case 0x1:
		// 1NNN: Jump to address NNN.
		value := instruction & 0xFFF
		debug("[%04X] Setting pc to %03X\n", instruction, value)
		ch8.pc = value

	case 0x6:
		// 6XNN: Store number NN in register VX.
		register := (instruction >> 8) & 0xF
		value := instruction & 0xFF
		debug("[%04X] Loading %X into register %d\n", instruction, value, register)
		ch8.V[register] = byte(value)

	case 0x7:
		// 7XNN: Add the value NN to register VX
		register := (instruction >> 8) & 0xF
		value := instruction & 0xFF
		debug("[%04X] Adding %X to contents of register %d\n", instruction, value, register)
		ch8.V[register] += byte(value)

	case 0xA:
		// ANNN: Set I to the address NNN.
		value := instruction & 0xFFF
		debug("[%04X] Loading %03X into I\n", instruction, value)
		ch8.I = value

	case 0xD:
		// DXYN" Draw a sprite at position VX, VY with N bytes of sprite data starting at the address stored in I
		// Set VF to 01 if any set pixels are changed to unset, and 00 otherwise

		// 1. Determine the X, Y values of where to start drawing.
		xReg := (instruction >> 8) & 0xF
		drawX := ch8.V[xReg] % displayWidth
		yReg := (instruction >> 4) & 0xF
		drawY := ch8.V[yReg] % displayHeight

		// 2. Set VF to 0
		ch8.V[0xF] = 0

		// 3. Determine how much sprite data to read
		//    This is how many contiguous blocks of memory, read from I, to draw.
		spriteHeight := instruction & 0x0F
		debug("[%04X] Drawing sprite from %v in memory at (%X, %X)\n", instruction, ch8.I, drawX, drawY)
		for y := uint16(0); y < spriteHeight; y++ {
			// Each byte in the sprite data is a line of 8 pixels.
			line := ch8.memory[ch8.I+y]
			// Check each bit in the line to see if we need to draw.
			for x := 0; x < 8; x++ {
				pixel := (line >> (7 - x)) & 1
				if pixel != 0 {
					// Pixel is not off, draw it.
					ch8.display.content[drawX+byte(x)][drawY+uint8(y)] ^= 1
				} else {
					// Reset this pixel
					if ch8.display.content[drawX+byte(x)][drawY+uint8(y)] != 0 {
						// This pixel was set, was turn on VF flag.
						ch8.V[0xF] = 1
					}
				}
			}
		}

	default:
		warn("[%04X] Unsupported instruction!\n", instruction)
	}
}
