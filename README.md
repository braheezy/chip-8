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

### Configuration
Various aspects of the interpreter can be tweaked in these ways, listed by precedence:
1. Setting the appropriate Environment Variable.
1. Creating a `config.toml` file in the same directory that `chip8` is being executed from

Running `chip8 --write-config` will attempt to dump all values and their default values into the default configuration location

This table summarizes the existing configuration values and how to set them.

| Configuration | Default | TOML | Environment |
|---------------|---------|------|-------------|
| Change the display scale factor.<br>**1** uses the original 64x32 pixel display. | 10 | `display_scale_factor` | `CHIP8_DISPLAY_SCALE_FACTOR` |
| Delay the rate the interpreter processes instructions | 0 | `throttle_speed` | `CHIP8_THROTTLE_SPEED` |
| Stop execution after this many instructions are executed | 0 | `cycle_limit` | `CHIP8_CYCLE_LIMIT` |

### Run Modes and Quirks
Timendus provides this succinct description of what Quirks are:
> CHIP-8, SUPER-CHIP and XO-CHIP have subtle differences in the way they interpret the bytecode. We often call these differences quirks...This is one of the hardest parts to "get right" and often a reason why "some games work, but some don't".

All quirks belong to some other variation of CHIP-8. They can be set individually in `config.toml`. To use the different generations of CHIP-8, run with the appropriate flag set. This is equivalent to enabling all the quirks for that chipset:

    # Enable all COSMAC VIP works
    chip8 --cosmac <ROM>

#### COSMAC VIP ####
The following quirks are grouped under `cosmac-vip` section in the configuration file.

| Configuration Value | Description |
|---------------------|-------------|
| `reset_vf`          | The AND, OR and XOR opcodes (`8xy1`, `8xy2`, and `8xy3`) reset the flags register (`VF`) to zero
| `increment_i`       | Increment the memory index while `Fx55` and `Fx56` operate

## Resources
- http://www.emulator101.com/introduction-to-chip-8.html
- https://tobiasvl.github.io/blog/write-a-chip-8-emulator/
- https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set
- https://github.com/Timendus/chip8-test-suite
- http://devernay.free.fr/hacks/chip8/C8TECH10.HTM
