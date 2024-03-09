# CHIP-8
A Go implementation of a [CHIP-8 interpreter](https://www.wikiwand.com/en/CHIP-8).

## Installation
Download the latest [release](https://github.com/braheezy/chip-8/releases) for your platform.

## Usage
Launch with your ROM:

    chip8 <chip-8 file>

Log instructions as they are processed (Warning! produces lots of messages):

    chip8 -debug <chip-8 file>

While the program passes all test ROMs from [Timendus' Test Suite](https://github.com/Timendus/chip8-test-suite), YMMV with random ROMs you pull from the Internet.

## Resources
- http://www.emulator101.com/introduction-to-chip-8.html
- https://tobiasvl.github.io/blog/write-a-chip-8-emulator/
- https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set
- https://github.com/Timendus/chip8-test-suite
- http://devernay.free.fr/hacks/chip8/C8TECH10.HTM
