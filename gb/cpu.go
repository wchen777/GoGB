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

func (r *Registers) Get_BC() uint16 {
	return (uint16(r.b) << 8) | uint16(r.c)
}

func (r *Registers) Get_DE() uint16 {
	return (uint16(r.d) << 8) | uint16(r.e)
}

func (r *Registers) Get_HL() uint16 {
	return (uint16(r.h) << 8) | uint16(r.l)
}

func (r *Registers) Set_BC(value uint16) {
	r.b = uint8((value & 0xFF00) >> 8)
	r.c = uint8(value & 0xFF)
}

func (r *Registers) Set_DE(value uint16) {
	r.d = uint8((value & 0xFF00) >> 8)
	r.e = uint8(value & 0xFF)
}

func (r *Registers) Set_HL(value uint16) {
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

func (r *Registers) Get_Zero_Flag() bool {
	return ((r.f & 0x80) >> 7) == 1
}

func (r *Registers) Get_Subtract_Flag() bool {
	return ((r.f & 0x40) >> 6) == 1
}

func (r *Registers) Get_Half_Carry_Flag() bool {
	return ((r.f & 0x20) >> 5) == 1
}

func (r *Registers) Get_Carry_Flag() bool {
	return ((r.f & 0x10) >> 4) == 1
}

func (r *Registers) Set_Zero_Flag(value bool) {
	if value {
		r.f |= 0x80
	} else {
		r.f &= 0x7F
	}
}

func (r *Registers) Set_Subtract_Flag(value bool) {
	if value {
		r.f |= 0x40
	} else {
		r.f &= 0xBF
	}
}

func (r *Registers) Set_Half_Carry_Flag(value bool) {
	if value {
		r.f |= 0x20
	} else {
		r.f &= 0xDF
	}
}

func (r *Registers) Set_Carry_Flag(value bool) {
	if value {
		r.f |= 0x10
	} else {
		r.f &= 0xEF
	}
}

// <----------------------------- INSTRUCTIONS -----------------------------> //

type CPU struct {
	reg Registers
}

var CLOCK_SPEED uint32 = 4194304
var FRAME_RATE uint32 = 60
var CYCLES_PER_FRAME uint32 = CLOCK_SPEED / FRAME_RATE
