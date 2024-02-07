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

	// The execution stack for subroutines
	stack Stack

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

func (ch8 *CHIP8) readNextInstruction() Instruction {
	// Read next instruction from memory.
	first := ch8.memory[ch8.pc]
	second := ch8.memory[ch8.pc+1]
	ch8.pc += 2

	currentCycle++
	return Instruction(uint16(first)<<8 | uint16(second))
}

// Instruction represents a 16-bit instruction
type Instruction uint16

// Nibbles returns a slice of nibbles from start to end (inclusive) in little-endian order
func (i Instruction) nibbles(start, end int) uint16 {
	if start < 0 || end > 4 || start > end {
		// Invalid range
		return 0
	}

	// Shift right to align the desired nibbles to the rightmost positions
	i >>= uint(4 * (3 - end))

	// Mask out the unwanted nibbles
	mask := uint16((1 << uint(4*(end-start+1))) - 1)
	return uint16(i) & mask
}

func (ch8 *CHIP8) stepInterpreter() {
	instruction := ch8.readNextInstruction()

	firstNibble := instruction.nibbles(0, 0)

	switch firstNibble {
	case 0x0:
		secondNibble := instruction.nibbles(1, 1)
		if secondNibble == 0x0 {
			lastByte := instruction.nibbles(3, 3)
			if lastByte == 0xE0 {
				// CLS: Clear the display
				debug("[%04X] CLS\n", instruction)
				ch8.display.clear()
			} else if lastByte == 0xEE {
				// RET: Return from a subroutine.
				// Get the PC from the stack an update accordingly
				var err error
				debug("[%04X] RET\n", instruction)
				ch8.pc, err = ch8.stack.Pop()
				if err != nil {
					panic(err)
				}
			}
		} else {
			// SYS addr: Jump to a machine code routine.
			debug("[%04X] SYS addr not supported!\n", instruction)
		}
	case 0x1:
		// 1NNN: Jump to address NNN.
		value := instruction.nibbles(1, 3)
		debug("[%04X] Setting pc to %03X\n", instruction, value)
		ch8.pc = value

	case 0x2:
		// 2NNN: Execute subroutine starting at address NNN
		// Push the current PC to the stack, then set the PC to NNN.
		value := instruction.nibbles(1, 3)
		ch8.stack.Push(ch8.pc)
		debug("[%04X] Setting pc to %03X\n", instruction, value)
		ch8.pc = value

	case 0x3:
		// 3XNN: Skip the next instruction if VX equals NN.
		register := instruction.nibbles(1, 1)
		value := instruction.nibbles(2, 3)
		debug("[%04X] Skipping next instruction if V%X == %X\n", instruction, register, value)
		if ch8.V[register] == byte(value) {
			ch8.pc += 2
		}

	case 0x4:
		// 4XNN: Skip the next instruction if VX does not equal NN.
		register := instruction.nibbles(1, 1)
		value := instruction.nibbles(2, 3)
		debug("[%04X] Skipping next instruction if V%X != %X\n", instruction, register, value)
		if ch8.V[register] != byte(value) {
			ch8.pc += 2
		}

	case 0x5:
		// 5XY0: Skip the next instruction if VX equals VY.
		registerX := instruction.nibbles(1, 1)
		registerY := instruction.nibbles(2, 2)
		debug("[%04X] Skipping next instruction if V%X == V%X\n", instruction, registerX, registerY)
		if ch8.V[registerX] == ch8.V[registerY] {
			ch8.pc += 2
		}

	case 0x6:
		// 6XNN: Store number NN in register VX.
		register := instruction.nibbles(1, 1)
		value := instruction.nibbles(2, 3)
		debug("[%04X] Loading %X into register %d\n", instruction, value, register)
		ch8.V[register] = byte(value)

	case 0x7:
		// 7XNN: Add the value NN to register VX
		register := instruction.nibbles(1, 1)
		value := instruction.nibbles(2, 3)
		debug("[%04X] Adding %X to contents of register %d\n", instruction, value, register)
		ch8.V[register] += byte(value)

	case 0x9:
		// 9XY0: Skip the next instruction if VX does not equal VY.
		registerX := instruction.nibbles(1, 1)
		registerY := instruction.nibbles(2, 2)
		debug("[%04X] Skipping next instruction if V%X != V%X\n", instruction, registerX, registerY)
		if ch8.V[registerX] != ch8.V[registerY] {
			ch8.pc += 2
		}

	case 0xA:
		// ANNN: Set I to the address NNN.
		value := instruction.nibbles(1, 3)
		debug("[%04X] Loading %03X into I\n", instruction, value)
		ch8.I = value

	case 0xD:
		// DXYN" Draw a sprite at position VX, VY with N bytes of sprite data starting at the address stored in I
		// Set VF to 01 if any set pixels are changed to unset, and 00 otherwise

		// 1. Determine the X, Y values of where to start drawing.
		xReg := instruction.nibbles(1, 1)
		drawX := ch8.V[xReg] % displayWidth
		yReg := instruction.nibbles(2, 2)
		drawY := ch8.V[yReg] % displayHeight

		// 2. Set VF to 0
		ch8.V[0xF] = 0

		// 3. Determine how much sprite data to read
		//    This is how many contiguous blocks of memory, read from I, to draw.
		spriteHeight := instruction.nibbles(3, 3)
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
