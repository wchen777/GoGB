package gb

// The Gameboy CPU is an 8-bit processor w/ a 16-bit address space.

// <----------------------------- REGISTERS -----------------------------> //

/*

The Gameboy CPU has the following registers:

	- Eight 8-bit registers: A, B, C, D, E, F, H, L
	- Two 16-bit registers: SP (Stack Pointer), PC (Program Counter)

*/

type Registers struct {
	a  uint8  // Accumulator
	b  uint8  // Register B
	c  uint8  // Register C
	d  uint8  // Register D
	e  uint8  // Register E
	h  uint8  // Register H
	l  uint8  // Register L
	f  uint8  // Flags
	pc uint16 // Program Counter
	sp uint16 // Stack Pointer
}

/*
	While the CPU only has 8 bit registers,
	there are instructions that allow the game to read and write 16 bits
	(i.e. 2 bytes) at the same time
*/

func (r *Registers) GetBC() uint16 {
	return (uint16(r.b) << 8) | uint16(r.c)
}

func (r *Registers) GetDE() uint16 {
	return (uint16(r.d) << 8) | uint16(r.e)
}

func (r *Registers) GetHL() uint16 {
	return (uint16(r.h) << 8) | uint16(r.l)
}

func (r *Registers) SetBC(value uint16) {
	r.b = uint8((value & 0xFF00) >> 8)
	r.c = uint8(value & 0xFF)
}

func (r *Registers) SetDE(value uint16) {
	r.d = uint8((value & 0xFF00) >> 8)
	r.e = uint8(value & 0xFF)
}

func (r *Registers) SetHL(value uint16) {
	r.h = uint8((value & 0xFF00) >> 8)
	r.l = uint8(value & 0xFF)
}

/*
	Flags register:

		- Bit 7: Zero flag
		- Bit 6: Subtract flag
		- Bit 5: Half carry flag
		- Bit 4: Carry flag
*/

// returns 1 or 0 depending on the value of the flags
func (r *Registers) GetZero() uint8 {
	return (r.f & 0x80) >> 7
}

func (r *Registers) GetSubtract() uint8 {
	return (r.f & 0x40) >> 6
}

func (r *Registers) GetHalfCarry() uint8 {
	return (r.f & 0x20) >> 5
}

func (r *Registers) GetCarry() uint8 {
	return (r.f & 0x10) >> 4
}

// sets fkags to 0 or 1 depending on the value of the bool
func (r *Registers) SetZero(value bool) {
	if value {
		r.f |= 0x80
	} else {
		r.f &= 0x7F
	}
}

func (r *Registers) SetSubtract(value bool) {
	if value {
		r.f |= 0x40
	} else {
		r.f &= 0xBF
	}
}

func (r *Registers) SetHalfCarry(value bool) {
	if value {
		r.f |= 0x20
	} else {
		r.f &= 0xDF
	}
}

func (r *Registers) SetCarry(value bool) {
	if value {
		r.f |= 0x10
	} else {
		r.f &= 0xEF
	}
}

// <----------------------------- INSTRUCTIONS -----------------------------> //

type CPU struct {
	regs Registers
}

// ADC - Add with Carry
func (cpu *CPU) ADC(value uint8) {

	// add value of carry flag to value, accounting for overflow with uint16
	result := uint16(value + cpu.regs.a + cpu.regs.GetCarry())

	cpu.regs.SetZero(cpu.regs.a == value)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(((value & 0xF) + (cpu.regs.a & 0xF)) > 0xF)
	cpu.regs.SetCarry((result & 0xff00) != 0)

	// set the accumulator to the result
	cpu.regs.a = uint8(result & 0xFF)
}

// <----------------------------- EXECUTION -----------------------------> //

// Step uses the program counter to read an instruction from memory and executes it
func (cpu *CPU) Step() int {
	// Use the program counter to read the instruction byte from memory.

	// Translate the byte to one of the instances of the Instruction enum

	// If we can successfully translate the instruction call our execute method else panic which now returns the next program counter

	// Set this next program counter on our CPU

	return 0
}

var CLOCK_SPEED uint32 = 4194304
var FRAME_RATE uint32 = 60
var CYCLES_PER_FRAME uint32 = CLOCK_SPEED / FRAME_RATE
