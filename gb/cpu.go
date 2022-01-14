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

// <----------------------------- CPU INSTRUCTIONS -----------------------------> //

type CPU struct {
	regs Registers
	mem  Memory
}

var stopped bool = false

// ADD - Add (w/ 8-bit address)
func (cpu *CPU) ADD(address *uint8, value uint8) {
	// Add the value to the accumulator and set the flags
	result := uint16(*address) + uint16(value)

	// set address to the result
	*address = uint8(result & 0xFF)

	cpu.regs.SetCarry((result & 0xff00) != 0)
	cpu.regs.SetZero(*address == 0)
	cpu.regs.SetHalfCarry(((*address & 0x0F) + (value & 0x0F)) > 0xF)
	cpu.regs.SetSubtract(false)

}

// ADD - Add (w/ 16-bit address)
func (cpu *CPU) ADD_16(address *uint16, value uint16) {
	// Add the value to the accumulator and set the flags
	result := uint32(*address + value)

	// set address to the result
	*address = uint16(result & 0xFFFF)

	cpu.regs.SetCarry((result & 0xFFFF0000) != 0)
	cpu.regs.SetZero(*address == 0)
	cpu.regs.SetHalfCarry(((*address & 0x0F) + (value & 0x0F)) > 0xF)
	cpu.regs.SetSubtract(false)
}

// ADC - Add with Carry
func (cpu *CPU) ADC(value uint8) {

	// add value of carry flag to value, accounting for overflow with uint16
	result := uint16(value) + uint16(cpu.regs.a) + uint16(cpu.regs.GetCarry())

	cpu.regs.SetZero(result == 0)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(((cpu.regs.a & 0x0F) + (value & 0x0F) + cpu.regs.GetCarry()) > 0xF)
	cpu.regs.SetCarry((result & 0xff00) != 0)

	// set the accumulator to the result
	cpu.regs.a = uint8(result & 0xFF)
}

// SUB - Subtract
func (cpu *CPU) SUB(value uint8) {

	cpu.regs.SetCarry(cpu.regs.a < value)
	cpu.regs.SetHalfCarry((cpu.regs.a & 0x0F) < (value & 0x0F))
	cpu.regs.SetSubtract(true)

	cpu.regs.a -= value
	cpu.regs.SetZero(cpu.regs.a == 0)

}

// SBC - Subtract with Carry
func (cpu *CPU) SBC(value uint8) {

	newValue := value + cpu.regs.GetCarry()

	cpu.regs.SetCarry(cpu.regs.a < newValue)
	cpu.regs.SetHalfCarry((cpu.regs.a & 0x0F) < (newValue & 0x0F))
	cpu.regs.SetSubtract(true)

	cpu.regs.a -= newValue
	cpu.regs.SetZero(cpu.regs.a == 0)

}

// AND - Logical AND
func (cpu *CPU) AND(value uint8) {
	cpu.regs.a &= value
	cpu.regs.SetZero(cpu.regs.a == 0)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(true)
	cpu.regs.SetCarry(false)
}

// OR - Logical OR
func (cpu *CPU) OR(value uint8) {
	cpu.regs.a |= value
	cpu.regs.SetZero(cpu.regs.a == 0)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)
	cpu.regs.SetCarry(false)

}

// XOR - Logical XOR
func (cpu *CPU) XOR(value uint8) {
	cpu.regs.a ^= value
	cpu.regs.SetZero(cpu.regs.a == 0)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)
	cpu.regs.SetCarry(false)

}

// CP - Compare
func (cpu *CPU) CP(value uint8) {
	cpu.regs.SetZero(cpu.regs.a == value)
	cpu.regs.SetCarry(cpu.regs.a < value)
	cpu.regs.SetHalfCarry((cpu.regs.a & 0x0F) < (value & 0x0F))
	cpu.regs.SetSubtract(true)
}

// INC - Increment
func (cpu *CPU) INC(value uint8) uint8 {
	cpu.regs.SetHalfCarry((value & 0x0F) == 0x0F)
	cpu.regs.SetSubtract(false)

	value++
	cpu.regs.SetZero(value == 0)

	return value
}

// DEC - Decrement
func (cpu *CPU) DEC(value uint8) uint8 {
	cpu.regs.SetHalfCarry((value & 0x0F) == 0)
	cpu.regs.SetSubtract(true)

	value--
	cpu.regs.SetZero(value == 0)

	return value
}

// <----------------------------- OPCODES + INSTRUCTIONS -----------------------------> //

// TODO: check all the + and - instructions

// 0x00 - NOP
func (cpu *CPU) NOP() {}

// 0x01 - LD BC, d16 (d16 means 16 bit immediate value, operand will be from PC)
func (cpu *CPU) LD_BC_d16(operand uint16) {
	cpu.regs.SetBC(operand)
}

// 0x02 - LD (BC), A
func (cpu *CPU) LD_BC_A() {
	// write at address bc the value of the accumulator

	cpu.mem.Write8(cpu.regs.GetBC(), cpu.regs.a)
}

// 0x03 - INC BC
func (cpu *CPU) INC_BC() {
	NN := cpu.regs.GetBC()
	NN++
	cpu.regs.SetBC(NN)
}

// 0x04 - INC B
func (cpu *CPU) INC_B() {
	cpu.regs.b = cpu.INC(cpu.regs.b)
}

// 0x05 - DEC B
func (cpu *CPU) DEC_B() {
	cpu.regs.b = cpu.DEC(cpu.regs.b)
}

// 0x06 - LD B, d8
func (cpu *CPU) LD_B_d8(operand uint8) {
	cpu.regs.b = operand
}

// 0x07 - RLCA (rotate left through carry)
func (cpu *CPU) RLCA() {
	cpu.regs.a = (cpu.regs.a << 1) | (cpu.regs.a >> 7)

	cpu.regs.SetZero(false)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)

	// set the carry flag to bit 0
	cpu.regs.SetCarry((cpu.regs.a & 0x01) != 0)

}

// 0x08 - LD (a16), SP (?)
func (cpu *CPU) LD_a16_SP(operand uint16) {
	// write the stack pointer to the address
	cpu.mem.Write16(operand, cpu.regs.sp)
}

// 0x09 - ADD HL, BC
func (cpu *CPU) ADD_HL_BC() {
	// TODO: why is this erroring?
	// cpu.ADD_16(&cpu.regs.GetHL(), cpu.regs.GetBC())
}

// 0x0A - LD A, (BC)
func (cpu *CPU) LD_A_BC() {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetBC())
}

// 0x0B - DEC BC
func (cpu *CPU) DEC_BC() {
	NN := cpu.regs.GetBC()
	NN--
	cpu.regs.SetBC(NN)
}

// 0x0C - INC C
func (cpu *CPU) INC_C() {
	cpu.regs.c = cpu.INC(cpu.regs.c)
}

// 0x0D - DEC C
func (cpu *CPU) DEC_C() {
	cpu.regs.c = cpu.DEC(cpu.regs.c)
}

// 0x0E - LD C, d8
func (cpu *CPU) LD_C_d8(operand uint8) {
	cpu.regs.c = operand
}

// 0x0F - RRCA (rotate right through carry)
func (cpu *CPU) RRCA() {
	// set the carry flag to bit 0
	cpu.regs.SetCarry((cpu.regs.a & 0x01) != 0)

	cpu.regs.a = (cpu.regs.a >> 1) | (cpu.regs.a << 7)

	cpu.regs.SetZero(false)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)

}

// 0x10 - STOP
func (cpu *CPU) STOP() {
	stopped := true
}

// 0x11 - LD DE, d16 (d16 means 16 bit immediate value, operand will be from PC)
func (cpu *CPU) LD_DE_d16(operand uint16) {
	cpu.regs.SetDE(operand)
}

// 0x12 - LD (DE), A
func (cpu *CPU) LD_DE_A() {
	// write at address bc the value of the accumulator
	cpu.mem.Write8(cpu.regs.GetDE(), cpu.regs.a)
}

// 0x13 - INC DE
func (cpu *CPU) INC_DE() {
	NN := cpu.regs.GetDE()
	NN++
	cpu.regs.SetDE(NN)
}

// 0x14 - INC D
func (cpu *CPU) INC_D() {
	cpu.regs.d = cpu.INC(cpu.regs.d)
}

// 0x15 - DEC D
func (cpu *CPU) DEC_D() {
	cpu.regs.d = cpu.DEC(cpu.regs.d)
}

// 0x16 - LD D, d8
func (cpu *CPU) LD_D_d8(operand uint8) {
	cpu.regs.d = operand
}

// 0x17 - RLA (rotate left through carry)
func (cpu *CPU) RLA() {

	// TODO: CHECK THIS

	// set the carry flag to bit 0
	cpu.regs.SetCarry((cpu.regs.a & 0x80) != 0)

	cpu.regs.a = (cpu.regs.a << 1) | (cpu.regs.a >> 7)

	cpu.regs.SetZero(false)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)

}

// 0x18 - JR r8 (r8 means 8 bit immediate value, operand will be from PC)
func (cpu *CPU) JR_r8(operand uint8) {
	cpu.regs.pc += uint16(operand)
}

// 0x19 - ADD HL, DE
func (cpu *CPU) ADD_HL_DE() {
	// TODO: why is this erroring?
	// cpu.ADD_16(&cpu.regs.GetHL(), cpu.regs.GetDE())
}

// cpu.ADD_16(cpu.regs.GetBC(), cpu.regs.GetHL())

// 0x1A - LD A, (DE)
func (cpu *CPU) LD_A_DE() {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetDE())
}

// 0x1B - DEC DE
func (cpu *CPU) DEC_DE() {
	NN := cpu.regs.GetDE()
	NN--
	cpu.regs.SetDE(NN)
}

// 0x1C - INC E
func (cpu *CPU) INC_E() {
	cpu.regs.e = cpu.INC(cpu.regs.e)
}

// 0x1D - DEC E
func (cpu *CPU) DEC_E() {
	cpu.regs.e = cpu.DEC(cpu.regs.e)
}

// 0x1E - LD E, d8
func (cpu *CPU) LD_E_d8(operand uint8) {
	cpu.regs.e = operand
}

// 0x1F - RRA (rotate right through carry)
func (cpu *CPU) RRA() {
	// TODO: check this
}

// 0x20 - JR NZ, r8 (r8 means 8 bit immediate value, operand will be from PC)
func (cpu *CPU) JR_NZ_r8(operand uint8) {
	// TODO: check this
	if cpu.regs.GetZero() == 0 {
		cpu.regs.pc += uint16(operand)
	}
}

// 0x21 - LD HL, d16 (d16 means 16 bit immediate value, operand will be from PC)
func (cpu *CPU) LD_HL_d16(operand uint16) {
	cpu.regs.SetHL(operand)
}

// 0x22 - LD (HL+), A
func (cpu *CPU) LD_HLp_A() {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.a)
}

// 0x23 - INC HL
func (cpu *CPU) INC_HL() {
	NN := cpu.regs.GetHL()
	NN++
	cpu.regs.SetHL(NN)
}

// 0x24 - INC H
func (cpu *CPU) INC_H() {
	cpu.regs.h = cpu.INC(cpu.regs.h)
}

// 0x25 - DEC H
func (cpu *CPU) DEC_H() {
	cpu.regs.h = cpu.DEC(cpu.regs.h)
}

// 0x26 - LD H, d8
func (cpu *CPU) LD_H_d8(operand uint8) {
	cpu.regs.h = operand
}

// 0x27 - DAA (decimal adjust accumulator)
func (cpu *CPU) DAA() {
	// TODO: this

}

// 0x28 - JR Z, r8
func (cpu *CPU) JR_Z_r8(operand uint8) {

}

// 0x29 - ADD HL, HL
func (cpu *CPU) ADD_HL_HL() {

}

// 0x2A - LD A, (HL+)
func (cpu *CPU) LD_A_HLp() {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x2B - DEC HL
func (cpu *CPU) DEC_HL() {
	NN := cpu.regs.GetHL()
	NN--
	cpu.regs.SetHL(NN)
}

// 0x2C - INC L
func (cpu *CPU) INC_L() {
	cpu.regs.l = cpu.INC(cpu.regs.l)
}

// 0x2D - DEC L
func (cpu *CPU) DEC_L() {
	cpu.regs.l = cpu.DEC(cpu.regs.l)
}

// 0x2E - LD L, d8
func (cpu *CPU) LD_L_d8(operand uint8) {
	cpu.regs.l = operand
}

// 0x2F - CPL (complement accumulator)
func (cpu *CPU) CPL() {
	// TODO: this
}

// 0x30 - JR NC, r8
func (cpu *CPU) JR_NC_r8(operand uint8) {
	// TODO: this
}

// 0x31 - LD SP, d16
func (cpu *CPU) LD_SP_d16(operand uint16) {
	cpu.regs.sp = operand
}

// 0x32 - LD (HL-), A
func (cpu *CPU) LD_HLm_A() {
	// write at address bc the value of the accumulator
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.a)
}

// 0x33 - INC SP
func (cpu *CPU) INC_SP() {
	cpu.regs.sp++
}

// 0x34 - INC (HL+)
func (cpu *CPU) INC_HLp() {
	// set hl to be the increment of the value of the address at hl
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.INC(cpu.mem.Read8(cpu.regs.GetHL())))
}

// 0x35 - DEC (HL+)
func (cpu *CPU) DEC_HLp() {
	// set hl to be the decrement of the value of the address at hl
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.DEC(cpu.mem.Read8(cpu.regs.GetHL())))
}

// 0x36 - LD (HL+), d8
func (cpu *CPU) LD_HLp_d8(operand uint8) {
	cpu.mem.Write8(cpu.regs.GetHL(), operand)
}

// 0x37 - SCF (set carry flag)

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
