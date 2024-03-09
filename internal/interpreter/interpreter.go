package interpreter

import (
	"bytes"
	_ "embed"
	"errors"
	"math/rand"
	"slices"
	"time"

	"github.com/charmbracelet/log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	// In CHIP-8, the program starts at address 0x200.
	programStartAddress = 0x200
	// The display parameters for original CHIP-8
	DisplayWidth  = 64
	DisplayHeight = 32
	// How often the delay and sound timer are decremented (in Hz)
	timerFrequency = 60
)

var (
	lastUpdate           time.Time
	lastDelayTimerUpdate time.Time
	lastSoundTimerUpdate time.Time
	decrementInterval    = time.Second / timerFrequency
	SupportedModes       = []string{
		"CHIP-8 (default)",
		"COSMAC-VIP",
	}
)

//go:embed beep.mp3
var beepData []byte

// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#font
var font = [16][5]byte{
	{0xF0, 0x90, 0x90, 0x90, 0xF0}, // char0
	{0x20, 0x60, 0x20, 0x20, 0x70}, // char1
	{0xF0, 0x10, 0xF0, 0x80, 0xF0}, // char2
	{0xF0, 0x10, 0xF0, 0x10, 0xF0}, // char3
	{0x90, 0x90, 0xF0, 0x10, 0x10}, // char4
	{0xF0, 0x80, 0xF0, 0x10, 0xF0}, // char5
	{0xF0, 0x80, 0xF0, 0x90, 0xF0}, // char6
	{0xF0, 0x10, 0x20, 0x40, 0x40}, // char7
	{0xF0, 0x90, 0xF0, 0x90, 0xF0}, // char8
	{0xF0, 0x90, 0xF0, 0x10, 0xF0}, // char9
	{0xF0, 0x90, 0xF0, 0x90, 0x90}, // charA
	{0xE0, 0x90, 0xE0, 0x90, 0xE0}, // charB
	{0xF0, 0x80, 0x80, 0x80, 0xF0}, // charC
	{0xE0, 0x90, 0x90, 0x90, 0xE0}, // charD
	{0xF0, 0x80, 0xF0, 0x80, 0xF0}, // charE
	{0xF0, 0x80, 0xF0, 0x80, 0x80}, // charF
}

type CHIP8 struct {
	// Define 4k of RAM.
	memory [4096]byte

	// Current instruction in memory to execute.
	pc uint16

	// Index register, holding addresses in memory.
	I uint16

	// CPU registers.
	V [16]byte

	// Delay timer. Decremented at 60Hz until 0.
	delayTimer byte

	// Sound timer. Decremented at 60Hz until 0. Emits beep while not 0.
	soundTimer byte

	// The beep to play, when appropriate
	beep *audio.Player

	// The execution stack for subroutines
	stack Stack

	// The current display of the program
	display Display

	// Size of program being executed.
	programSize int

	// Represent the currently pressed keys
	pressedKeys []byte

	// If set, there pressedKeys that need to be processed
	dirtyKeys bool

	// Tweakable settings to use when running the interpreter
	Options CHIP8Options

	// Logger object to use
	Logger *log.Logger
}

type CHIP8Options struct {
	// So it can be seen on modern displays
	DisplayScaleFactor int `mapstructure:"display_scale_factor"`
	// Max cycle speed of CHIP-8 exec loop
	ThrottleSpeed int `mapstructure:"throttle_speed"`
	// Limit how many instruction_limits the program is run for. For debug purposes.
	InstructionLimit int `mapstructure:"instruction_limit"`
	// Enable quirks for COSMAC-like behavior
	CosmacQuirks COSMACQuirks `mapstructure:"cosmac-vip"`
}

type COSMACQuirks struct {
	// reset VF to 0 during instructions 8XY1, 8XY2, 8XY3
	ResetVF bool `mapstructure:"reset_vf"`
	// increment memory index I during FX55 and FX65
	IncrementI bool `mapstructure:"increment_i"`
}

func (cq *COSMACQuirks) EnableAll() {
	cq.ResetVF = true
	cq.IncrementI = true
}

// NewDefaultApp creates a new App with default options.
func NewCHIP8(program *[]byte) *CHIP8 {
	chip8 := &CHIP8{
		pc:      programStartAddress,
		Options: DefaultCHIP8Options(),
	}

	chip8.programSize = len(*program) + programStartAddress
	// Load program into memory.
	copy(chip8.memory[programStartAddress:], *program)

	chip8.display.onColor = Pine.RGBA()
	chip8.display.offColor = Gold.RGBA()

	// Load font into memory
	// From 0x000 to 0x1FF
	for i := 0; i < 16; i++ {
		for j := 0; j < 5; j++ {
			chip8.memory[i*5+j] = font[i][j]
		}
	}

	// Load sound file
	data, err := mp3.DecodeWithSampleRate(44100, bytes.NewReader(beepData))
	if err != nil {
		panic(err)
	}
	chip8.beep, err = audio.NewContext(44100).NewPlayer(data)
	if err != nil {
		panic(err)
	}

	return chip8
}

func DefaultCHIP8Options() CHIP8Options {
	return CHIP8Options{
		DisplayScaleFactor: 1,
		ThrottleSpeed:      0,
		InstructionLimit:   -1,
		CosmacQuirks:       COSMACQuirks{},
	}
}

func (ch8 *CHIP8) readNextInstruction() Instruction {
	// Read next instruction from memory.
	first := ch8.memory[ch8.pc]
	second := ch8.memory[ch8.pc+1]
	ch8.pc += 2

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

	exec := true

	for exec {

		// Throttle the CPU based on a configurable delay
		if ch8.Options.ThrottleSpeed*10 > 0 {
			elapsedTime := time.Since(lastUpdate)
			delay := time.Second / time.Duration(ch8.Options.ThrottleSpeed*10)
			if elapsedTime < delay {
				time.Sleep(delay - elapsedTime)
			}
			lastUpdate = time.Now()
		}

		if ch8.Options.InstructionLimit != -1 && int(ch8.pc)/2 >= ch8.Options.InstructionLimit {
			break
		}

		instruction := ch8.readNextInstruction()
		ch8.Logger.Debugf("[%04X] %04X", ch8.pc-2, instruction)

		firstNibble := instruction.nibbles(0, 0)

		switch firstNibble {
		case 0x0:
			secondNibble := instruction.nibbles(1, 1)
			if secondNibble == 0x0 {
				lastByte := instruction.nibbles(3, 3)
				if lastByte == 0x0 {
					// 00E0: Clear the display
					ch8.Logger.Debugf("[%04X] CLS", instruction)
					ch8.display.clear()
				} else if lastByte == 0xE {
					// 00EE: Return from a subroutine.
					// Get the PC from the stack an update accordingly
					var err error
					ch8.Logger.Debugf("[%04X] RET", instruction)
					ch8.pc, err = ch8.stack.Pop()
					if err != nil {
						panic(err)
					}
				} else {
					ch8.Logger.Warnf("[%04X] Unsupported instruction!", instruction)
				}
			} else {
				// 0NNN: Jump to a machine code routine.
				ch8.Logger.Warnf("[%04X] Machine code execution not supported!", instruction)
			}

		case 0x1:
			// 1NNN: Jump to address NNN.
			value := instruction.nibbles(1, 3)
			ch8.Logger.Debugf("[%04X] Setting pc to %03X", instruction, value)
			ch8.pc = value
			exec = false

		case 0x2:
			// 2NNN: Execute subroutine starting at address NNN
			// Push the current PC to the stack, then set the PC to NNN.
			value := instruction.nibbles(1, 3)
			ch8.stack.Push(ch8.pc)
			ch8.Logger.Debugf("[%04X] Setting pc to %03X", instruction, value)
			ch8.pc = value

		case 0x3:
			// 3XNN: Skip the next instruction if VX equals NN.
			register := instruction.nibbles(1, 1)
			value := instruction.nibbles(2, 3)
			ch8.Logger.Debugf("[%04X] Skipping next instruction if %X == %X", instruction, ch8.V[register], value)
			if ch8.V[register] == byte(value) {
				ch8.pc += 2
			}

		case 0x4:
			// 4XNN: Skip the next instruction if VX does not equal NN.
			register := instruction.nibbles(1, 1)
			value := instruction.nibbles(2, 3)
			ch8.Logger.Debugf("[%04X] Skipping next instruction if V%X != %X", instruction, register, value)
			if ch8.V[register] != byte(value) {
				ch8.pc += 2
			}

		case 0x5:
			// 5XY0: Skip the next instruction if VX equals VY.
			registerX := instruction.nibbles(1, 1)
			registerY := instruction.nibbles(2, 2)
			ch8.Logger.Debugf("[%04X] Skipping next instruction if V%X == V%X", instruction, registerX, registerY)
			if ch8.V[registerX] == ch8.V[registerY] {
				ch8.pc += 2
			}

		case 0x6:
			// 6XNN: Store number NN in register VX.
			register := instruction.nibbles(1, 1)
			value := instruction.nibbles(2, 3)
			ch8.Logger.Debugf("[%04X] Loading %X into V%d", instruction, value, register)
			ch8.V[register] = byte(value)

		case 0x7:
			// 7XNN: Add the value NN to register VX
			register := instruction.nibbles(1, 1)
			value := instruction.nibbles(2, 3)
			ch8.Logger.Debugf("[%04X] Adding %X to contents of V%d", instruction, value, register)
			ch8.V[register] += byte(value)

		case 0x8:
			// Arithmetic and logical operators
			registerX := instruction.nibbles(1, 1)
			registerY := instruction.nibbles(2, 2)
			lastNibble := instruction.nibbles(3, 3)
			switch lastNibble {
			case 0x0:
				// 8XY0: Store the value of register VY in register VX.
				ch8.Logger.Debugf("[%04X] Loading contents of V%d into V%d", instruction, registerY, registerX)
				ch8.V[registerX] = ch8.V[registerY]

			case 0x1:
				// 8XY1: Set VX to VX OR VY.
				ch8.Logger.Debugf("[%04X] Loading (V%d | V%d) into V%d", instruction, registerY, registerX, registerX)
				ch8.V[registerX] |= ch8.V[registerY]
				if ch8.Options.CosmacQuirks.ResetVF {
					ch8.V[0xF] = 0
				}

			case 0x2:
				// 8XY2: Set VX to VX AND VY.
				ch8.Logger.Debugf("[%04X] Loading (V%d & V%d) into V%d", instruction, registerY, registerX, registerX)
				ch8.V[registerX] &= ch8.V[registerY]
				if ch8.Options.CosmacQuirks.ResetVF {
					ch8.V[0xF] = 0
				}

			case 0x3:
				// 8XY3: Set VX to VX XOR VY.
				ch8.Logger.Debugf("[%04X] Loading (V%d XOR V%d) into V%d", instruction, registerY, registerX, registerX)
				ch8.V[registerX] ^= ch8.V[registerY]
				if ch8.Options.CosmacQuirks.ResetVF {
					ch8.V[0xF] = 0
				}

			case 0x4:
				// 8XY4: Add the value of register VY to register VX
				// Set VF to 01 if a carry occurs
				// Set VF to 00 if a carry does not occur
				ch8.Logger.Debugf("[%04X] Adding V%d to V%d and storing into V%d", instruction, registerY, registerX, registerX)
				carry := uint16(ch8.V[registerX])+uint16(ch8.V[registerY]) > 0xFF
				ch8.V[registerX] += ch8.V[registerY]
				ch8.V[0xF] = 0
				if carry {
					ch8.V[0xF] = 1
				}

			case 0x5:
				// 8XY5: Subtract the value of register VY from register VX
				// Set VF to 00 if a borrow occurs
				// Set VF to 01 if a borrow does not occur
				ch8.Logger.Debugf("[%04X] Subtracting V%d from V%d and storing into V%d", instruction, registerY, registerX, registerX)
				borrow := ch8.V[registerY] > ch8.V[registerX]
				ch8.V[registerX] -= ch8.V[registerY]
				ch8.V[0xF] = 1
				if borrow {
					ch8.V[0xF] = 0
				}

			case 0x6:
				// 8XY6: Store the value of register VY shifted right one bit in register VX
				// Set VF to the least significant bit prior to the shift.
				// TODO: COSMAC VIP: Set VX to value in VY first
				ch8.Logger.Debugf("[%04X] Shifting V%d right and storing into V%d", instruction, registerY, registerX)
				value := ch8.V[registerY]
				ch8.V[registerX] = value >> 1
				ch8.V[0xF] = value & 0x1

			case 0x7:
				// 8XY7: Subtract the value of register VX from register VY
				// Set VF to 00 if a borrow occurs
				// Set VF to 01 if a borrow does not occur
				ch8.Logger.Debugf("[%04X] Subtracting V%d from V%d and storing into V%d", instruction, registerX, registerY, registerX)
				borrow := ch8.V[registerX] > ch8.V[registerY]
				ch8.V[registerX] = ch8.V[registerY] - ch8.V[registerX]
				ch8.V[0xF] = 1
				if borrow {
					ch8.V[0xF] = 0
				}

			case 0xE:
				// 8XYE: Store the value of register VY shifted left one bit in register VX
				// Set VF to the least significant bit prior to the shift.
				// TODO: COSMAC VIP: Set VX to value in VY first
				ch8.Logger.Debugf("[%04X] Shifting V%d right and storing into V%d", instruction, registerY, registerX)
				value := ch8.V[registerY]
				ch8.V[registerX] = value << 1
				ch8.V[0xF] = value >> 7
			}

		case 0x9:
			// 9XY0: Skip the next instruction if VX does not equal VY.
			registerX := instruction.nibbles(1, 1)
			registerY := instruction.nibbles(2, 2)
			ch8.Logger.Debugf("[%04X] Skipping next instruction if V%X != V%X", instruction, registerX, registerY)
			if ch8.V[registerX] != ch8.V[registerY] {
				ch8.pc += 2
			}

		case 0xA:
			// ANNN: Set I to the address NNN.
			value := instruction.nibbles(1, 3)
			ch8.Logger.Debugf("[%04X] Loading %03X into I", instruction, value)
			ch8.I = value

		case 0xB:
			// BNNN: Jump to the address NNN plus V0.
			// TODO: CHIP-48 and SUPER-CHIP: Jump to BXNN
			value := instruction.nibbles(1, 3)
			ch8.Logger.Debugf("[%04X] Setting pc to %03X + V0", instruction, value)
			ch8.pc = value + uint16(ch8.V[0])

		case 0xC:
			// CXNN: Set VX to a random number AND NN.
			value := instruction.nibbles(2, 3)
			registerX := instruction.nibbles(1, 1)
			randomNumber := rand.Intn(256)
			ch8.Logger.Debugf("[%04X] Setting V%X to (%d AND %X)", instruction, registerX, randomNumber, value)
			ch8.V[registerX] = byte(randomNumber & int(value))

		case 0xD:
			// DXYN: Draw a sprite at position VX, VY with N bytes of sprite data starting at the address stored in I
			// Set VF to 01 if any set pixels are changed to unset, and 00 otherwise

			// 1. Determine the X, Y values of where to start drawing.
			xReg := instruction.nibbles(1, 1)
			drawX := ch8.V[xReg] % DisplayWidth
			yReg := instruction.nibbles(2, 2)
			drawY := ch8.V[yReg] % DisplayHeight

			// 2. Set VF to 0
			ch8.V[0xF] = 0

			// 3. Determine how much sprite data to read
			//    This is how many contiguous blocks of memory, read from I, to draw.
			spriteHeight := instruction.nibbles(3, 3)
			ch8.Logger.Debugf("[%04X] Drawing %d-sized sprite at (%d, %d)", instruction, spriteHeight, drawX, drawY)
			for y := uint16(0); y < spriteHeight; y++ {
				// Each byte in the sprite data is a line of 8 pixels.
				line := ch8.memory[ch8.I+y]
				yLoc := drawY + byte(y)
				for x := 0; x < 8 && drawX+byte(x) < DisplayWidth && yLoc < DisplayHeight; x++ {
					xLoc := drawX + byte(x)
					pixel := (line >> (7 - x)) & 1

					if pixel != 0 {
						currentPixel := ch8.display.content[xLoc][yLoc]
						ch8.display.content[xLoc][yLoc] ^= 1

						if currentPixel != 0 {
							// Pixel was set, turn on VF flag.
							ch8.V[0xF] = 1
						}
					}
				}
			}
			exec = false

		case 0xE:
			lastHalf := instruction.nibbles(2, 3)
			registerX := instruction.nibbles(1, 1)
			switch lastHalf {
			case 0x9E:
				// EX9E: Skip the next instruction if the key stored in VX is pressed
				hexKey := ch8.V[registerX]
				for _, pressedKey := range ch8.pressedKeys {
					if pressedKey == hexKey {
						ch8.Logger.Debugf("[%04X] Skipping next instruction b/c %X key is pressed", instruction, hexKey)
						ch8.pc += 2
						ch8.dirtyKeys = false
						exec = false
						break
					}
				}

			case 0xA1:
				// EXA1: Skip the next instruction if the key stored in VX is not pressed
				if len(ch8.pressedKeys) == 0 {
					ch8.Logger.Debugf("[%04X] Skipping next instruction b/c no keys are pressed", instruction)
					ch8.pc += 2
				} else {
					hexKey := ch8.V[registerX]
					// Some key is pressed, is it the one we care about?
					if !slices.Contains(ch8.pressedKeys, hexKey) {
						ch8.Logger.Debugf("[%04X] Skipping next instruction b/c %X key is not pressed", instruction, hexKey)
						ch8.pc += 2
					}
				}
				ch8.dirtyKeys = false
				exec = false
			}

		case 0xF:
			lastHalf := instruction.nibbles(2, 3)
			registerX := instruction.nibbles(1, 1)
			switch lastHalf {
			case 0x07:
				// FX07: Set VX to the value of the delay timer.
				ch8.Logger.Debugf("[%04X] Loading contents of delay timer into V%d", instruction, registerX)
				ch8.V[registerX] = ch8.delayTimer

			case 0x15:
				// FX15: Set the delay timer to the value of register VX
				ch8.Logger.Debugf("[%04X] Setting delay timer to contents of V%d", instruction, registerX)
				ch8.delayTimer = ch8.V[registerX]
				lastDelayTimerUpdate = time.Now()

			case 0x18:
				// FX18: Set the sound timer to the value of register VX
				ch8.Logger.Debugf("[%04X] Setting sound timer to %d", instruction, ch8.V[registerX])
				ch8.soundTimer = ch8.V[registerX]
				lastSoundTimerUpdate = time.Now()

			case 0x1E:
				// FX1E: Add the value of register VX to register I
				// TODO: set VF to 1 if I “overflows” from 0FFF to above 1000 for Amiga quirk
				ch8.Logger.Debugf("[%04X] Adding contents of V%d to I", instruction, registerX)
				ch8.I += uint16(ch8.V[registerX])

			case 0x0A:
				// FX0A: Wait for key press, put hex value in VX
				if len(ch8.pressedKeys) > 0 {
					hexKey := ch8.pressedKeys[0]
					if inpututil.IsKeyJustReleased(ebiten.Key(hexKey)) {
						ch8.Logger.Debugf("[%04X] Waited for key press and storing %X value in V%d", instruction, hexKey, registerX)
						ch8.V[registerX] = hexKey
					}
				} else {
					ch8.pc -= 2
				}
				exec = false

			case 0x29:
				// FX29: Set I to the location of the sprite for the character in register VX
				// TODO: An 8-bit register can hold two hexadecimal numbers, but this would only point to one character. The original COSMAC VIP interpreter just took the last nibble of VX and used that as the character.
				ch8.Logger.Debugf("[%04X] Setting I to memory address of font character in V%d", instruction, registerX)
				ch8.I = uint16(ch8.V[registerX]) * 5

			case 0x33:
				// FX33: Store the binary-coded decimal equivalent of the value stored in register VX at addresses I, I + 1, and I + 2
				ch8.Logger.Debugf("[%04X] Storing BCD of V%d at memory addresses I, I + 1, and I + 2", instruction, registerX)
				ch8.memory[ch8.I] = ch8.V[registerX] / 100
				ch8.memory[ch8.I+1] = (ch8.V[registerX] / 10) % 10
				ch8.memory[ch8.I+2] = ch8.V[registerX] % 10

			case 0x55:
				// FX55: Store registers V0 through VX in memory starting at address I
				ch8.Logger.Debugf("[%04X] Storing V0 through V%d at memory address I", instruction, registerX)
				for i := uint16(0); i <= uint16(registerX); i++ {
					ch8.memory[ch8.I+i] = ch8.V[i]
				}
				if ch8.Options.CosmacQuirks.IncrementI {
					// COSMAC VIP incremented the I register while it worked. Each time it stored or loaded one register, it incremented I. After the instruction was finished, I would be set to the new value I + X + 1.
					ch8.I = registerX + 1
				}

			case 0x65:
				// FX65: Read registers V0 through VX from memory starting at address I
				ch8.Logger.Debugf("[%04X] Reading V0 through V%d from memory address I", instruction, registerX)
				for i := uint16(0); i <= uint16(registerX); i++ {
					ch8.V[i] = ch8.memory[ch8.I+i]
				}
				if ch8.Options.CosmacQuirks.IncrementI {
					// COSMAC VIP incremented the I register while it worked. Each time it stored or loaded one register, it incremented I. After the instruction was finished, I would be set to the new value I + X + 1.
					ch8.I = registerX + 1
				}

			default:
				ch8.Logger.Warnf("[%04X] Unsupported instruction!", instruction)
			}

		default:
			ch8.Logger.Warnf("[%04X] Unsupported instruction!", instruction)
		}
	}
}

// Convert keypad key to hex value
func keyToHex(key ebiten.Key) (byte, error) {
	var hexValue byte
	switch key {
	case ebiten.KeyX:
		hexValue = 0x0
	case ebiten.Key1:
		hexValue = 0x1
	case ebiten.Key2:
		hexValue = 0x2
	case ebiten.Key3:
		hexValue = 0x3
	case ebiten.Key4:
		hexValue = 0xC
	case ebiten.KeyQ:
		hexValue = 0x4
	case ebiten.KeyW:
		hexValue = 0x5
	case ebiten.KeyE:
		hexValue = 0x6
	case ebiten.KeyR:
		hexValue = 0xD
	case ebiten.KeyA:
		hexValue = 0x7
	case ebiten.KeyS:
		hexValue = 0x8
	case ebiten.KeyD:
		hexValue = 0x9
	case ebiten.KeyF:
		hexValue = 0xE
	case ebiten.KeyZ:
		hexValue = 0xA
	case ebiten.KeyC:
		hexValue = 0xB
	case ebiten.KeyV:
		hexValue = 0xF
	default:
		return 0, errors.New("unsupported key")
	}
	return hexValue, nil
}
