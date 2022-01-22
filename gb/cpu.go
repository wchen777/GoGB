package gb

import (
	"fmt"
)

// The Gameboy CPU is an 8-bit processor w/ a 16-bit address space.

// <----------------------------- TYPEDEFS -----------------------------> //

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

type OperandInfo struct {
	operand8  uint8
	operand16 uint16
}

type Instruction struct {
	name             string
	instuctionLength uint8 // number of bytes for the instruction
	execute          func(*OperandInfo)
}

type CPU struct {
	regs       Registers
	mem        Memory
	table      [256]Instruction
	ticksTable [256]uint8
	ticks      uint32
	stopped    bool
}

// <----------------------------- REGISTERS -----------------------------> //

/*

The Gameboy CPU has the following registers:

	- Eight 8-bit registers: A, B, C, D, E, F, H, L
	- Two 16-bit registers: SP (Stack Pointer), PC (Program Counter)

*/

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
func (cpu *CPU) NOP(stepInfo *OperandInfo) {}

// 0x01 - LD BC, d16 (d16 means 16 bit immediate value, operand will be from PC)
func (cpu *CPU) LD_BC_d16(stepInfo *OperandInfo) {
	cpu.regs.SetBC(stepInfo.operand16)
}

// 0x02 - LD (BC), A
func (cpu *CPU) LD_BC_A(stepInfo *OperandInfo) {
	// write at address bc the value of the accumulator

	cpu.mem.Write8(cpu.regs.GetBC(), cpu.regs.a)
}

// 0x03 - INC BC
func (cpu *CPU) INC_BC(stepInfo *OperandInfo) {
	NN := cpu.regs.GetBC()
	NN++
	cpu.regs.SetBC(NN)
}

// 0x04 - INC B
func (cpu *CPU) INC_B(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.INC(cpu.regs.b)
}

// 0x05 - DEC B
func (cpu *CPU) DEC_B(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.DEC(cpu.regs.b)
}

// 0x06 - LD B, d8
func (cpu *CPU) LD_B_d8(stepInfo *OperandInfo) {
	cpu.regs.b = stepInfo.operand8
}

// 0x07 - RLCA (rotate left through carry)
func (cpu *CPU) RLCA(stepInfo *OperandInfo) {
	cpu.regs.a = (cpu.regs.a << 1) | (cpu.regs.a >> 7)

	cpu.regs.SetZero(false)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)

	// set the carry flag to bit 0
	cpu.regs.SetCarry((cpu.regs.a & 0x01) != 0)

}

// 0x08 - LD (a16), SP (?)
func (cpu *CPU) LD_a16_SP(stepInfo *OperandInfo) {
	// write the stack pointer to the address
	cpu.mem.Write16(stepInfo.operand16, cpu.regs.sp)
}

// 0x09 - ADD HL, BC
func (cpu *CPU) ADD_HL_BC(stepInfo *OperandInfo) {
	// TODO: why is this erroring?
	// cpu.ADD_16(&cpu.regs.GetHL(), cpu.regs.GetBC())
}

// 0x0A - LD A, (BC)
func (cpu *CPU) LD_A_BC(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetBC())
}

// 0x0B - DEC BC
func (cpu *CPU) DEC_BC(stepInfo *OperandInfo) {
	NN := cpu.regs.GetBC()
	NN--
	cpu.regs.SetBC(NN)
}

// 0x0C - INC C
func (cpu *CPU) INC_C(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.INC(cpu.regs.c)
}

// 0x0D - DEC C
func (cpu *CPU) DEC_C(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.DEC(cpu.regs.c)
}

// 0x0E - LD C, d8
func (cpu *CPU) LD_C_d8(stepInfo *OperandInfo) {
	cpu.regs.c = stepInfo.operand8
}

// 0x0F - RRCA (rotate right through carry)
func (cpu *CPU) RRCA(stepInfo *OperandInfo) {
	// set the carry flag to bit 0
	cpu.regs.SetCarry((cpu.regs.a & 0x01) != 0)

	cpu.regs.a = (cpu.regs.a >> 1) | (cpu.regs.a << 7)

	cpu.regs.SetZero(false)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)

}

// 0x10 - STOP
func (cpu *CPU) STOP(stepInfo *OperandInfo) {
	cpu.stopped = true
}

// 0x11 - LD DE, d16 (d16 means 16 bit immediate value, operand will be from PC)
func (cpu *CPU) LD_DE_d16(stepInfo *OperandInfo) {
	cpu.regs.SetDE(stepInfo.operand16)
}

// 0x12 - LD (DE), A
func (cpu *CPU) LD_DE_A(stepInfo *OperandInfo) {
	// write at address bc the value of the accumulator
	cpu.mem.Write8(cpu.regs.GetDE(), cpu.regs.a)
}

// 0x13 - INC DE
func (cpu *CPU) INC_DE(stepInfo *OperandInfo) {
	NN := cpu.regs.GetDE()
	NN++
	cpu.regs.SetDE(NN)
}

// 0x14 - INC D
func (cpu *CPU) INC_D(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.INC(cpu.regs.d)
}

// 0x15 - DEC D
func (cpu *CPU) DEC_D(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.DEC(cpu.regs.d)
}

// 0x16 - LD D, d8
func (cpu *CPU) LD_D_d8(stepInfo *OperandInfo) {
	cpu.regs.d = stepInfo.operand8
}

// 0x17 - RLA (rotate left through carry)
func (cpu *CPU) RLA(stepInfo *OperandInfo) {

	// TODO: CHECK THIS

	// set the carry flag to bit 0
	cpu.regs.SetCarry((cpu.regs.a & 0x80) != 0)

	cpu.regs.a = (cpu.regs.a << 1) | (cpu.regs.a >> 7)

	cpu.regs.SetZero(false)
	cpu.regs.SetSubtract(false)
	cpu.regs.SetHalfCarry(false)

}

// 0x18 - JR r8 (r8 means 8 bit immediate value, operand will be from PC)
func (cpu *CPU) JR_r8(stepInfo *OperandInfo) {
	cpu.regs.pc += uint16(stepInfo.operand8)
}

// 0x19 - ADD HL, DE
func (cpu *CPU) ADD_HL_DE(stepInfo *OperandInfo) {
	// TODO: why is this erroring?
	// cpu.ADD_16(&cpu.regs.GetHL(), cpu.regs.GetDE())
}

// cpu.ADD_16(cpu.regs.GetBC(), cpu.regs.GetHL())

// 0x1A - LD A, (DE)
func (cpu *CPU) LD_A_DE(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetDE())
}

// 0x1B - DEC DE
func (cpu *CPU) DEC_DE(stepInfo *OperandInfo) {
	NN := cpu.regs.GetDE()
	NN--
	cpu.regs.SetDE(NN)
}

// 0x1C - INC E
func (cpu *CPU) INC_E(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.INC(cpu.regs.e)
}

// 0x1D - DEC E
func (cpu *CPU) DEC_E(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.DEC(cpu.regs.e)
}

// 0x1E - LD E, d8
func (cpu *CPU) LD_E_d8(stepInfo *OperandInfo) {
	cpu.regs.e = stepInfo.operand8
}

// 0x1F - RRA (rotate right through carry)
func (cpu *CPU) RRA(stepInfo *OperandInfo) {
	// TODO: check this
}

// 0x20 - JR NZ, r8 (r8 means 8 bit immediate value, operand will be from PC)
func (cpu *CPU) JR_NZ_r8(stepInfo *OperandInfo) {
	// TODO: check this
	if cpu.regs.GetZero() == 0 {
		cpu.regs.pc += uint16(stepInfo.operand8)
	}
}

// 0x21 - LD HL, d16 (d16 means 16 bit immediate value, operand will be from PC)
func (cpu *CPU) LD_HL_d16(stepInfo *OperandInfo) {
	cpu.regs.SetHL(stepInfo.operand16)
}

// 0x22 - LD (HL+), A
func (cpu *CPU) LDi_HLp_A(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.a)
	cpu.regs.SetHL(cpu.regs.GetHL() + 1)
}

// 0x23 - INC HL
func (cpu *CPU) INC_HL(stepInfo *OperandInfo) {
	NN := cpu.regs.GetHL()
	NN++
	cpu.regs.SetHL(NN)
}

// 0x24 - INC H
func (cpu *CPU) INC_H(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.INC(cpu.regs.h)
}

// 0x25 - DEC H
func (cpu *CPU) DEC_H(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.DEC(cpu.regs.h)
}

// 0x26 - LD H, d8
func (cpu *CPU) LD_H_d8(stepInfo *OperandInfo) {
	cpu.regs.h = stepInfo.operand8
}

// 0x27 - DAA (decimal adjust accumulator)
func (cpu *CPU) DAA(stepInfo *OperandInfo) {
	// TODO: this

}

// 0x28 - JR Z, r8
func (cpu *CPU) JR_Z_r8(stepInfo *OperandInfo) {

}

// 0x29 - ADD HL, HL
func (cpu *CPU) ADD_HL_HL(stepInfo *OperandInfo) {

}

// 0x2A - LD A, (HL+)
func (cpu *CPU) LDi_A_HLp(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetHL())
	cpu.regs.SetHL(cpu.regs.GetHL() + 1)
}

// 0x2B - DEC HL
func (cpu *CPU) DEC_HL(stepInfo *OperandInfo) {
	NN := cpu.regs.GetHL()
	NN--
	cpu.regs.SetHL(NN)
}

// 0x2C - INC L
func (cpu *CPU) INC_L(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.INC(cpu.regs.l)
}

// 0x2D - DEC L
func (cpu *CPU) DEC_L(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.DEC(cpu.regs.l)
}

// 0x2E - LD L, d8
func (cpu *CPU) LD_L_d8(stepInfo *OperandInfo) {
	cpu.regs.l = stepInfo.operand8
}

// 0x2F - CPL (complement accumulator)
func (cpu *CPU) CPL(stepInfo *OperandInfo) {
	// TODO: this
}

// 0x30 - JR NC, r8
func (cpu *CPU) JR_NC_r8(stepInfo *OperandInfo) {
	// TODO: this
}

// 0x31 - LD SP, d16
func (cpu *CPU) LD_SP_d16(stepInfo *OperandInfo) {
	cpu.regs.sp = stepInfo.operand16
}

// 0x32 - LD (HL-), A
func (cpu *CPU) LD_HLm_A(stepInfo *OperandInfo) {
	// write at address bc the value of the accumulator
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.a)
	cpu.regs.SetHL(cpu.regs.GetHL() - 1)
}

// 0x33 - INC SP
func (cpu *CPU) INC_SP(stepInfo *OperandInfo) {
	cpu.regs.sp++
}

// 0x34 - INC (HL+)
func (cpu *CPU) INC_HLp(stepInfo *OperandInfo) {
	// set hl to be the increment of the value of the address at hl
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.INC(cpu.mem.Read8(cpu.regs.GetHL())))
}

// 0x35 - DEC (HL+)
func (cpu *CPU) DEC_HLp(stepInfo *OperandInfo) {
	// set hl to be the decrement of the value of the address at hl
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.DEC(cpu.mem.Read8(cpu.regs.GetHL())))
}

// 0x36 - LD (HL+), d8
func (cpu *CPU) LD_HLp_d8(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), stepInfo.operand8)
}

// 0x37 - SCF (set carry flag)
func (cpu *CPU) SCF(stepInfo *OperandInfo) {
	cpu.regs.SetCarry(true)
	cpu.regs.SetZero(false)
	cpu.regs.SetHalfCarry(false)
}

// 0x38 - JR C, r8
func (cpu *CPU) JR_C_r8(stepInfo *OperandInfo) {
	// TODO: this
}

// 0x39 - ADD HL, SP
func (cpu *CPU) ADD_HL_SP(stepInfo *OperandInfo) {
	// TODO: use add functions?
}

// 0x3A - LD A, (HL-)
func (cpu *CPU) LD_A_HLm(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetHL())
	cpu.regs.SetHL(cpu.regs.GetHL() - 1)
}

// 0x3B - DEC SP
func (cpu *CPU) DEC_SP(stepInfo *OperandInfo) {
	cpu.regs.sp--
}

// 0x3C - INC A
func (cpu *CPU) INC_A(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.INC(cpu.regs.a)
}

// 0x3D - DEC A
func (cpu *CPU) DEC_A(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.DEC(cpu.regs.a)
}

// 0x3E - LD A, d8
func (cpu *CPU) LD_A_d8(stepInfo *OperandInfo) {
	cpu.regs.a = stepInfo.operand8
}

// 0x3F - CCF (complement carry flag)
func (cpu *CPU) CCF(stepInfo *OperandInfo) {
	// TODO: ??
}

// 0x40 - LD B, B
func (cpu *CPU) LD_B_B(stepInfo *OperandInfo) {
	// NOP
}

// 0x41 - LD B, C
func (cpu *CPU) LD_B_C(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.regs.c
}

// 0x42 - LD B, D
func (cpu *CPU) LD_B_D(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.regs.d
}

// 0x43 - LD B, E
func (cpu *CPU) LD_B_E(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.regs.e
}

// 0x44 - LD B, H
func (cpu *CPU) LD_B_H(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.regs.h
}

// 0x45 - LD B, L
func (cpu *CPU) LD_B_L(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.regs.l
}

// 0x46 - LD B, (HL+)
func (cpu *CPU) LD_B_HLp(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x47 - LD B, A
func (cpu *CPU) LD_B_A(stepInfo *OperandInfo) {
	cpu.regs.b = cpu.regs.a
}

// 0x48 - LD C, B
func (cpu *CPU) LD_C_B(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.regs.b
}

// 0x49 - LD C, C
func (cpu *CPU) LD_C_C(stepInfo *OperandInfo) {
	// NOP
}

// 0x4A - LD C, D
func (cpu *CPU) LD_C_D(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.regs.d
}

// 0x4B - LD C, E
func (cpu *CPU) LD_C_E(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.regs.e
}

// 0x4C - LD C, H
func (cpu *CPU) LD_C_H(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.regs.h
}

// 0x4D - LD C, L
func (cpu *CPU) LD_C_L(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.regs.l
}

// 0x4E - LD C, (HL+)
func (cpu *CPU) LD_C_HLp(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x4F - LD C, A
func (cpu *CPU) LD_C_A(stepInfo *OperandInfo) {
	cpu.regs.c = cpu.regs.a
}

// 0x50 - LD D, B
func (cpu *CPU) LD_D_B(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.regs.b
}

// 0x51 - LD D, C
func (cpu *CPU) LD_D_C(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.regs.c
}

// 0x52 - LD D, D
func (cpu *CPU) LD_D_D(stepInfo *OperandInfo) {
	// NOP
}

// 0x53 - LD D, E
func (cpu *CPU) LD_D_E(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.regs.e
}

// 0x54 - LD D, H
func (cpu *CPU) LD_D_H(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.regs.h
}

// 0x55 - LD D, L
func (cpu *CPU) LD_D_L(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.regs.l
}

// 0x56 - LD D, (HL+)
func (cpu *CPU) LD_D_HLp(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x57 - LD D, A
func (cpu *CPU) LD_D_A(stepInfo *OperandInfo) {
	cpu.regs.d = cpu.regs.a
}

// 0x58 - LD E, B
func (cpu *CPU) LD_E_B(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.regs.b
}

// 0x59 - LD E, C
func (cpu *CPU) LD_E_C(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.regs.c
}

// 0x5A - LD E, D
func (cpu *CPU) LD_E_D(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.regs.d
}

// 0x5B - LD E, E
func (cpu *CPU) LD_E_E(stepInfo *OperandInfo) {
	// NOP
}

// 0x5C - LD E, H
func (cpu *CPU) LD_E_H(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.regs.h
}

// 0x5D - LD E, L
func (cpu *CPU) LD_E_L(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.regs.l
}

// 0x5E - LD E, (HL+)
func (cpu *CPU) LD_E_HLp(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x5F - LD E, A
func (cpu *CPU) LD_E_A(stepInfo *OperandInfo) {
	cpu.regs.e = cpu.regs.a
}

// 0x60 - LD H, B
func (cpu *CPU) LD_H_B(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.regs.b
}

// 0x61 - LD H, C
func (cpu *CPU) LD_H_C(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.regs.c
}

// 0x62 - LD H, D
func (cpu *CPU) LD_H_D(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.regs.d
}

// 0x63 - LD H, E
func (cpu *CPU) LD_H_E(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.regs.e
}

// 0x64 - LD H, H
func (cpu *CPU) LD_H_H(stepInfo *OperandInfo) {
	// NOP
}

// 0x65 - LD H, L
func (cpu *CPU) LD_H_L(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.regs.l
}

// 0x66 - LD H, (HL+)
func (cpu *CPU) LD_H_HLp(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x67 - LD H, A
func (cpu *CPU) LD_H_A(stepInfo *OperandInfo) {
	cpu.regs.h = cpu.regs.a
}

// 0x68 - LD L, B
func (cpu *CPU) LD_L_B(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.regs.b
}

// 0x69 - LD L, C
func (cpu *CPU) LD_L_C(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.regs.c
}

// 0x6A - LD L, D
func (cpu *CPU) LD_L_D(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.regs.d
}

// 0x6B - LD L, E
func (cpu *CPU) LD_L_E(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.regs.e
}

// 0x6C - LD L, H
func (cpu *CPU) LD_L_H(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.regs.h
}

// 0x6D - LD L, L
func (cpu *CPU) LD_L_L(stepInfo *OperandInfo) {
	// NOP
}

// 0x6E - LD L, (HL+)
func (cpu *CPU) LD_L_HLp(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x6F - LD L, A
func (cpu *CPU) LD_L_A(stepInfo *OperandInfo) {
	cpu.regs.l = cpu.regs.a
}

// 0x70 - LD (HL+), B
func (cpu *CPU) LD_HLp_B(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.b)
}

// 0x71 - LD (HL+), C
func (cpu *CPU) LD_HLp_C(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.c)
}

// 0x72 - LD (HL+), D
func (cpu *CPU) LD_HLp_D(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.d)
}

// 0x73 - LD (HL+), E
func (cpu *CPU) LD_HLp_E(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.e)
}

// 0x74 - LD (HL+), H
func (cpu *CPU) LD_HLp_H(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.h)
}

// 0x75 - LD (HL+), L
func (cpu *CPU) LD_HLp_L(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.l)
}

// 0x76 - HALT
func (cpu *CPU) HALT(stepInfo *OperandInfo) {
	// TODO: this
	// halt execution until an interrupt occurs, use interrupt information to determine if an interrupt is pending
	// else increment pc
}

// 0x77 - LD (HL+), A
func (cpu *CPU) LD_HL_A(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.a)
}

// 0x78 - LD A, B
func (cpu *CPU) LD_A_B(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.regs.b
}

// 0x79 - LD A, C
func (cpu *CPU) LD_A_C(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.regs.c
}

// 0x7A - LD A, D
func (cpu *CPU) LD_A_D(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.regs.d
}

// 0x7B - LD A, E
func (cpu *CPU) LD_A_E(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.regs.e
}

// 0x7C - LD A, H
func (cpu *CPU) LD_A_H(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.regs.h
}

// 0x7D - LD A, L
func (cpu *CPU) LD_A_L(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.regs.l
}

// 0x7E - LD A, (HL+)
func (cpu *CPU) LD_A_HLp(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetHL())
}

// 0x7F - LD A, A
func (cpu *CPU) LD_A_A(stepInfo *OperandInfo) {
	// NOP
}

// 0x80 - ADD A, B
func (cpu *CPU) ADD_A_B(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.regs.b)
}

// 0x81 - ADD A, C
func (cpu *CPU) ADD_A_C(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.regs.c)
}

// 0x82 - ADD A, D
func (cpu *CPU) ADD_A_D(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.regs.d)
}

// 0x83 - ADD A, E
func (cpu *CPU) ADD_A_E(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.regs.e)
}

// 0x84 - ADD A, H
func (cpu *CPU) ADD_A_H(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.regs.h)
}

// 0x85 - ADD A, L
func (cpu *CPU) ADD_A_L(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.regs.l)
}

// 0x86 - ADD A, (HL+)
func (cpu *CPU) ADD_A_HL(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.mem.Read8(cpu.regs.GetHL()))
}

// 0x87 - ADD A, A
func (cpu *CPU) ADD_A_A(stepInfo *OperandInfo) {
	cpu.ADD(&cpu.regs.a, cpu.regs.a)
}

// 0x88 - ADC A, B
func (cpu *CPU) ADC_A_B(stepInfo *OperandInfo) {
	cpu.ADC(cpu.regs.b)
}

// 0x89 - ADC A, C
func (cpu *CPU) ADC_A_C(stepInfo *OperandInfo) {
	cpu.ADC(cpu.regs.c)
}

// 0x8A - ADC A, D
func (cpu *CPU) ADC_A_D(stepInfo *OperandInfo) {
	cpu.ADC(cpu.regs.d)
}

// 0x8B - ADC A, E
func (cpu *CPU) ADC_A_E(stepInfo *OperandInfo) {
	cpu.ADC(cpu.regs.e)
}

// 0x8C - ADC A, H
func (cpu *CPU) ADC_A_H(stepInfo *OperandInfo) {
	cpu.ADC(cpu.regs.h)
}

// 0x8D - ADC A, L
func (cpu *CPU) ADC_A_L(stepInfo *OperandInfo) {
	cpu.ADC(cpu.regs.l)
}

// 0x8E - ADC A, (HL)
func (cpu *CPU) ADC_A_HL(stepInfo *OperandInfo) {
	cpu.ADC(cpu.mem.Read8(cpu.regs.GetHL()))
}

// 0x8F - ADC A, A
func (cpu *CPU) ADC_A_A(stepInfo *OperandInfo) {
	cpu.ADC(cpu.regs.a)
}

// 0x90 - SUB B
func (cpu *CPU) SUB_B(stepInfo *OperandInfo) {
	cpu.SUB(cpu.regs.b)
}

// 0x91 - SUB C
func (cpu *CPU) SUB_C(stepInfo *OperandInfo) {
	cpu.SUB(cpu.regs.c)
}

// 0x92 - SUB D
func (cpu *CPU) SUB_D(stepInfo *OperandInfo) {
	cpu.SUB(cpu.regs.d)
}

// 0x93 - SUB E
func (cpu *CPU) SUB_E(stepInfo *OperandInfo) {
	cpu.SUB(cpu.regs.e)
}

// 0x94 - SUB H
func (cpu *CPU) SUB_H(stepInfo *OperandInfo) {
	cpu.SUB(cpu.regs.h)
}

// 0x95 - SUB L
func (cpu *CPU) SUB_L(stepInfo *OperandInfo) {
	cpu.SUB(cpu.regs.l)
}

// 0x96 - SUB (HL+)
func (cpu *CPU) SUB_HL(stepInfo *OperandInfo) {
	cpu.SUB(cpu.mem.Read8(cpu.regs.GetHL()))
}

// 0x97 - SUB A
func (cpu *CPU) SUB_A(stepInfo *OperandInfo) {
	cpu.SUB(cpu.regs.a)
}

// 0x98 - SBC A, B
func (cpu *CPU) SBC_A_B(stepInfo *OperandInfo) {
	cpu.SBC(cpu.regs.b)
}

// 0x99 - SBC A, C
func (cpu *CPU) SBC_A_C(stepInfo *OperandInfo) {
	cpu.SBC(cpu.regs.c)
}

// 0x9A - SBC A, D
func (cpu *CPU) SBC_A_D(stepInfo *OperandInfo) {
	cpu.SBC(cpu.regs.d)
}

// 0x9B - SBC A, E
func (cpu *CPU) SBC_A_E(stepInfo *OperandInfo) {
	cpu.SBC(cpu.regs.e)
}

// 0x9C - SBC A, H
func (cpu *CPU) SBC_A_H(stepInfo *OperandInfo) {
	cpu.SBC(cpu.regs.h)
}

// 0x9D - SBC A, L
func (cpu *CPU) SBC_A_L(stepInfo *OperandInfo) {
	cpu.SBC(cpu.regs.l)
}

// 0x9E - SBC A, (HL)
func (cpu *CPU) SBC_A_HL(stepInfo *OperandInfo) {
	cpu.SBC(cpu.mem.Read8(cpu.regs.GetHL()))
}

// 0x9F - SBC A, A
func (cpu *CPU) SBC_A_A(stepInfo *OperandInfo) {
	cpu.SBC(cpu.regs.a)
}

// 0xA0 - AND B
func (cpu *CPU) AND_B(stepInfo *OperandInfo) {
	cpu.AND(cpu.regs.b)
}

// 0xA1 - AND C
func (cpu *CPU) AND_C(stepInfo *OperandInfo) {
	cpu.AND(cpu.regs.c)
}

// 0xA2 - AND D
func (cpu *CPU) AND_D(stepInfo *OperandInfo) {
	cpu.AND(cpu.regs.d)
}

// 0xA3 - AND E
func (cpu *CPU) AND_E(stepInfo *OperandInfo) {
	cpu.AND(cpu.regs.e)
}

// 0xA4 - AND H
func (cpu *CPU) AND_H(stepInfo *OperandInfo) {
	cpu.AND(cpu.regs.h)
}

// 0xA5 - AND L
func (cpu *CPU) AND_L(stepInfo *OperandInfo) {
	cpu.AND(cpu.regs.l)
}

// 0xA6 - AND (HL)
func (cpu *CPU) AND_HL(stepInfo *OperandInfo) {
	cpu.AND(cpu.mem.Read8(cpu.regs.GetHL()))
}

// 0xA7 - AND A
func (cpu *CPU) AND_A(stepInfo *OperandInfo) {
	cpu.AND(cpu.regs.a)
}

// 0xA8 - XOR B
func (cpu *CPU) XOR_B(stepInfo *OperandInfo) {
	cpu.XOR(cpu.regs.b)
}

// 0xA9 - XOR C
func (cpu *CPU) XOR_C(stepInfo *OperandInfo) {
	cpu.XOR(cpu.regs.c)
}

// 0xAA - XOR D
func (cpu *CPU) XOR_D(stepInfo *OperandInfo) {
	cpu.XOR(cpu.regs.d)
}

// 0xAB - XOR E
func (cpu *CPU) XOR_E(stepInfo *OperandInfo) {
	cpu.XOR(cpu.regs.e)
}

// 0xAC - XOR H
func (cpu *CPU) XOR_H(stepInfo *OperandInfo) {
	cpu.XOR(cpu.regs.h)
}

// 0xAD - XOR L
func (cpu *CPU) XOR_L(stepInfo *OperandInfo) {
	cpu.XOR(cpu.regs.l)
}

// 0xAE - XOR (HL)
func (cpu *CPU) XOR_HL(stepInfo *OperandInfo) {
	cpu.XOR(cpu.mem.Read8(cpu.regs.GetHL()))
}

// 0xAF - XOR A
func (cpu *CPU) XOR_A(stepInfo *OperandInfo) {
	cpu.XOR(cpu.regs.a)
}

// 0xB0 - OR B
func (cpu *CPU) OR_B(stepInfo *OperandInfo) {
	cpu.OR(cpu.regs.b)
}

// 0xB1 - OR C
func (cpu *CPU) OR_C(stepInfo *OperandInfo) {
	cpu.OR(cpu.regs.c)
}

// 0xB2 - OR D
func (cpu *CPU) OR_D(stepInfo *OperandInfo) {
	cpu.OR(cpu.regs.d)
}

// 0xB3 - OR E
func (cpu *CPU) OR_E(stepInfo *OperandInfo) {
	cpu.OR(cpu.regs.e)
}

// 0xB4 - OR H
func (cpu *CPU) OR_H(stepInfo *OperandInfo) {
	cpu.OR(cpu.regs.h)
}

// 0xB5 - OR L
func (cpu *CPU) OR_L(stepInfo *OperandInfo) {
	cpu.OR(cpu.regs.l)
}

// 0xB6 - OR (HL)
func (cpu *CPU) OR_HL(stepInfo *OperandInfo) {
	cpu.OR(cpu.mem.Read8(cpu.regs.GetHL()))
}

// 0xB7 - OR A
func (cpu *CPU) OR_A(stepInfo *OperandInfo) {
	cpu.OR(cpu.regs.a)
}

// 0xB8 - CP B
func (cpu *CPU) CP_B(stepInfo *OperandInfo) {
	cpu.CP(cpu.regs.b)
}

// 0xB9 - CP C
func (cpu *CPU) CP_C(stepInfo *OperandInfo) {
	cpu.CP(cpu.regs.c)
}

// 0xBA - CP D
func (cpu *CPU) CP_D(stepInfo *OperandInfo) {
	cpu.CP(cpu.regs.d)
}

// 0xBB - CP E
func (cpu *CPU) CP_E(stepInfo *OperandInfo) {
	cpu.CP(cpu.regs.e)
}

// 0xBC - CP H
func (cpu *CPU) CP_H(stepInfo *OperandInfo) {
	cpu.CP(cpu.regs.h)
}

// 0xBD - CP L
func (cpu *CPU) CP_L(stepInfo *OperandInfo) {
	cpu.CP(cpu.regs.l)
}

// 0xBE - CP (HL)
func (cpu *CPU) CP_HL(stepInfo *OperandInfo) {
	cpu.CP(cpu.mem.Read8(cpu.regs.GetHL()))
}

//  0xBF - CP A
func (cpu *CPU) CP_A(stepInfo *OperandInfo) {
	cpu.CP(cpu.regs.a)
}

// 0xC0 - RET NZ
func (cpu *CPU) RET_NZ(stepInfo *OperandInfo) {
	// TODO: check this
	if cpu.regs.GetZero() == 0 {
		cpu.regs.pc = cpu.mem.Read16(cpu.regs.sp)
		cpu.regs.sp += 2
	}

	cpu.regs.pc++
}

// 0xC1 - POP BC
func (cpu *CPU) POP_BC(stepInfo *OperandInfo) {
	// TODO: check this
	cpu.regs.SetBC(cpu.mem.Read16(cpu.regs.sp))
	cpu.regs.sp += 2
	cpu.regs.pc++
}

// 0xC2 - JP NZ,nn
func (cpu *CPU) JP_NZ_NN(stepInfo *OperandInfo) {
	// TODO: check this
	if cpu.regs.GetZero() == 0 {
		cpu.regs.pc = stepInfo.operand16
	} else {
		cpu.regs.pc += 3
	}
}

// 0xC3 - JP nn
func (cpu *CPU) JP_NN(stepInfo *OperandInfo) {
	// TODO: check this
	cpu.regs.pc = stepInfo.operand16
}

func (cpu *CPU) UNKNOWN(stepInfo *OperandInfo) {
	fmt.Printf("Unknown opcode!")
}

// <----------------------------- EXECUTION -----------------------------> //

func (cpu *CPU) CreateTable() {
	cpu.table = [256]Instruction{
		{"NOP", 0, cpu.NOP},                // 0x00
		{"LD BC, d16", 3, cpu.LD_BC_d16},   // 0x01
		{"LD (BC), A", 1, cpu.LD_BC_A},     // 0x02
		{"INC BC", 1, cpu.INC_BC},          // 0x03
		{"INC B", 1, cpu.INC_B},            // 0x04
		{"DEC B", 1, cpu.DEC_B},            // 0x05
		{"LD B, d8", 2, cpu.LD_B_d8},       // 0x06
		{"RLCA", 1, cpu.RLCA},              // 0x07
		{"LD (a16), SP", 3, cpu.LD_a16_SP}, // 0x08
		{"ADD HL, BC", 1, cpu.ADD_HL_BC},   // 0x09
		{"LD A, (BC)", 1, cpu.LD_A_BC},     // 0x0A
		{"DEC BC", 1, cpu.DEC_BC},          // 0x0B
		{"INC C", 1, cpu.INC_C},            // 0x0C
		{"DEC C", 1, cpu.DEC_C},            // 0x0D
		{"LD C, d8", 2, cpu.LD_C_d8},       // 0x0E
		{"RRCA", 1, cpu.RRCA},              // 0x0F
		{"STOP", 1, cpu.STOP},              // 0x10
		{"LD DE, d16", 3, cpu.LD_DE_d16},   // 0x11
		{"LD (DE), A", 1, cpu.LD_DE_A},     // 0x12
		{"INC DE", 1, cpu.INC_DE},          // 0x13
		{"INC D", 1, cpu.INC_D},            // 0x14
		{"DEC D", 1, cpu.DEC_D},            // 0x15
		{"LD D, d8", 2, cpu.LD_D_d8},       // 0x16
		{"RLA", 1, cpu.RLA},                // 0x17
		{"JR r8", 2, cpu.JR_r8},            // 0x18
		{"ADD HL, DE", 1, cpu.ADD_HL_DE},   // 0x19
		{"LD A, (DE)", 1, cpu.LD_A_DE},     // 0x1A
		{"DEC DE", 1, cpu.DEC_DE},          // 0x1B
		{"INC E", 1, cpu.INC_E},            // 0x1C
		{"DEC E", 1, cpu.DEC_E},            // 0x1D
		{"LD E, d8", 2, cpu.LD_E_d8},       // 0x1E
		{"RRA", 1, cpu.RRA},                // 0x1F
		{"JR NZ, r8", 2, cpu.JR_NZ_r8},     // 0x20
		{"LD HL, d16", 3, cpu.LD_HL_d16},   // 0x21
		{"LD (HL+), A", 1, cpu.LDi_HLp_A},  // 0x22
		{"INC HL", 1, cpu.INC_HL},          // 0x23
		{"INC H", 1, cpu.INC_H},            // 0x24
		{"DEC H", 1, cpu.DEC_H},            // 0x25
		{"LD H, d8", 2, cpu.LD_H_d8},       // 0x26
		{"DAA", 1, cpu.DAA},                // 0x27
		{"JR Z, r8", 2, cpu.JR_Z_r8},       // 0x28
		{"ADD HL, HL", 1, cpu.ADD_HL_HL},   // 0x29
		{"LD A, (HL+)", 1, cpu.LDi_A_HLp},  // 0x2A
		{"DEC HL", 1, cpu.DEC_HL},          // 0x2B
		{"INC L", 1, cpu.INC_L},            // 0x2C
		{"DEC L", 1, cpu.DEC_L},            // 0x2D
		{"LD L, d8", 2, cpu.LD_L_d8},       // 0x2E
		{"CPL", 1, cpu.CPL},                // 0x2F
		{"JR NC, r8", 2, cpu.JR_NC_r8},     // 0x30
		{"LD SP, d16", 3, cpu.LD_SP_d16},   // 0x31
		{"LD (HL-), A", 1, cpu.LD_HLm_A},   // 0x32
		{"INC SP", 1, cpu.INC_SP},          // 0x33
		{"INC (HL+)", 1, cpu.INC_HLp},      // 0x34
		{"DEC (HL)", 1, cpu.DEC_HLp},       // 0x35
		{"LD (HL), d8", 2, cpu.LD_HLp_d8},  // 0x36
		{"SCF", 1, cpu.SCF},                // 0x37
		{"JR C, r8", 2, cpu.JR_C_r8},       // 0x38
		{"ADD HL, SP", 1, cpu.ADD_HL_SP},   // 0x39
		{"LD A, (HL-)", 1, cpu.LD_A_HLm},   // 0x3A
		{"DEC SP", 1, cpu.DEC_SP},          // 0x3B
		{"INC A", 1, cpu.INC_A},            // 0x3C
		{"DEC A", 1, cpu.DEC_A},            // 0x3D
		{"LD A, d8", 2, cpu.LD_A_d8},       // 0x3E
		{"CCF", 1, cpu.CCF},                // 0x3F
		{"LD B, B", 1, cpu.LD_B_B},         // 0x40
		{"LD B, C", 1, cpu.LD_B_C},         // 0x41
		{"LD B, D", 1, cpu.LD_B_D},         // 0x42
		{"LD B, E", 1, cpu.LD_B_E},         // 0x43
		{"LD B, H", 1, cpu.LD_B_H},         // 0x44
		{"LD B, L", 1, cpu.LD_B_L},         // 0x45
		{"LD B, (HL+)", 1, cpu.LD_B_HLp},   // 0x46
		{"LD B, A", 1, cpu.LD_B_A},         // 0x47
		{"LD C, B", 1, cpu.LD_C_B},         // 0x48
		{"LD C, C", 1, cpu.LD_C_C},         // 0x49
		{"LD C, D", 1, cpu.LD_C_D},         // 0x4A
		{"LD C, E", 1, cpu.LD_C_E},         // 0x4B
		{"LD C, H", 1, cpu.LD_C_H},         // 0x4C
		{"LD C, L", 1, cpu.LD_C_L},         // 0x4D
		{"LD C, (HL+)", 1, cpu.LD_C_HLp},   // 0x4E
		{"LD C, A", 1, cpu.LD_C_A},         // 0x4F
		{"LD D, B", 1, cpu.LD_D_B},         // 0x50
		{"LD D, C", 1, cpu.LD_D_C},         // 0x51
		{"LD D, D", 1, cpu.LD_D_D},         // 0x52
		{"LD D, E", 1, cpu.LD_D_E},         // 0x53
		{"LD D, H", 1, cpu.LD_D_H},         // 0x54
		{"LD D, L", 1, cpu.LD_D_L},         // 0x55
		{"LD D, (HL+)", 1, cpu.LD_D_HLp},   // 0x56
		{"LD D, A", 1, cpu.LD_D_A},         // 0x57
		{"LD E, B", 1, cpu.LD_E_B},         // 0x58
		{"LD E, C", 1, cpu.LD_E_C},         // 0x59
		{"LD E, D", 1, cpu.LD_E_D},         // 0x5A
		{"LD E, E", 1, cpu.LD_E_E},         // 0x5B
		{"LD E, H", 1, cpu.LD_E_H},         // 0x5C
		{"LD E, L", 1, cpu.LD_E_L},         // 0x5D
		{"LD E, (HL+)", 1, cpu.LD_E_HLp},   // 0x5E
		{"LD E, A", 1, cpu.LD_E_A},         // 0x5F
		{"LD H, B", 1, cpu.LD_H_B},         // 0x60
		{"LD H, C", 1, cpu.LD_H_C},         // 0x61
		{"LD H, D", 1, cpu.LD_H_D},         // 0x62
		{"LD H, E", 1, cpu.LD_H_E},         // 0x63
		{"LD H, H", 1, cpu.LD_H_H},         // 0x64
		{"LD H, L", 1, cpu.LD_H_L},         // 0x65
		{"LD H, (HL+)", 1, cpu.LD_H_HLp},   // 0x66
		{"LD H, A", 1, cpu.LD_H_A},         // 0x67
		{"LD L, B", 1, cpu.LD_L_B},         // 0x68
		{"LD L, C", 1, cpu.LD_L_C},         // 0x69
		{"LD L, D", 1, cpu.LD_L_D},         // 0x6A
		{"LD L, E", 1, cpu.LD_L_E},         // 0x6B
		{"LD L, H", 1, cpu.LD_L_H},         // 0x6C
		{"LD L, L", 1, cpu.LD_L_L},         // 0x6D
		{"LD L, (HL+)", 1, cpu.LD_L_HLp},   // 0x6E
		{"LD L, A", 1, cpu.LD_L_A},         // 0x6F
		{"LD (HL+), B", 1, cpu.LD_HLp_B},   // 0x70
		{"LD (HL+), C", 1, cpu.LD_HLp_C},   // 0x71
		{"LD (HL+), D", 1, cpu.LD_HLp_D},   // 0x72
		{"LD (HL+), E", 1, cpu.LD_HLp_E},   // 0x73
		{"LD (HL+), H", 1, cpu.LD_HLp_H},   // 0x74
		{"LD (HL+), L", 1, cpu.LD_HLp_L},   // 0x75
		{"HALT", 1, cpu.HALT},              // 0x76
		{"LD (HL), A", 1, cpu.LD_HL_A},     // 0x77
		{"LD A, B", 1, cpu.LD_A_B},         // 0x78
		{"LD A, C", 1, cpu.LD_A_C},         // 0x79
		{"LD A, D", 1, cpu.LD_A_D},         // 0x7A
		{"LD A, E", 1, cpu.LD_A_E},         // 0x7B
		{"LD A, H", 1, cpu.LD_A_H},         // 0x7C
		{"LD A, L", 1, cpu.LD_A_L},         // 0x7D
		{"LD A, (HL+)", 1, cpu.LD_A_HLp},   // 0x7E
		{"LD A, A", 1, cpu.LD_A_A},         // 0x7F
		{"ADD A, B", 1, cpu.ADD_A_B},       // 0x80
		{"ADD A, C", 1, cpu.ADD_A_C},       // 0x81
		{"ADD A, D", 1, cpu.ADD_A_D},       // 0x82
		{"ADD A, E", 1, cpu.ADD_A_E},       // 0x83
		{"ADD A, H", 1, cpu.ADD_A_H},       // 0x84
		{"ADD A, L", 1, cpu.ADD_A_L},       // 0x85
		{"ADD A, (HL)", 1, cpu.ADD_A_HL},   // 0x86
		{"ADD A, A", 1, cpu.ADD_A_A},       // 0x87
		{"ADC A, B", 1, cpu.ADC_A_B},       // 0x88
		{"ADC A, C", 1, cpu.ADC_A_C},       // 0x89
		{"ADC A, D", 1, cpu.ADC_A_D},       // 0x8A
		{"ADC A, E", 1, cpu.ADC_A_E},       // 0x8B
		{"ADC A, H", 1, cpu.ADC_A_H},       // 0x8C
		{"ADC A, L", 1, cpu.ADC_A_L},       // 0x8D
		{"ADC A, (HL)", 1, cpu.ADC_A_HL},   // 0x8E
		{"ADC A, A", 1, cpu.ADC_A_A},       // 0x8F
		{"SUB B", 1, cpu.SUB_B},            // 0x90
		{"SUB C", 1, cpu.SUB_C},            // 0x91
		{"SUB D", 1, cpu.SUB_D},            // 0x92
		{"SUB E", 1, cpu.SUB_E},            // 0x93
		{"SUB H", 1, cpu.SUB_H},            // 0x94
		{"SUB L", 1, cpu.SUB_L},            // 0x95
		{"SUB (HL)", 1, cpu.SUB_HL},        // 0x96
		{"SUB A", 1, cpu.SUB_A},            // 0x97
		{"SBC A, B", 1, cpu.SBC_A_B},       // 0x98
		{"SBC A, C", 1, cpu.SBC_A_C},       // 0x99
		{"SBC A, D", 1, cpu.SBC_A_D},       // 0x9A
		{"SBC A, E", 1, cpu.SBC_A_E},       // 0x9B
		{"SBC A, H", 1, cpu.SBC_A_H},       // 0x9C
		{"SBC A, L", 1, cpu.SBC_A_L},       // 0x9D
		{"SBC A, (HL)", 1, cpu.SBC_A_HL},   // 0x9E
		{"SBC A, A", 1, cpu.SBC_A_A},       // 0x9F
		{"AND B", 1, cpu.AND_B},            // 0xA0
		{"AND C", 1, cpu.AND_C},            // 0xA1
		{"AND D", 1, cpu.AND_D},            // 0xA2
		{"AND E", 1, cpu.AND_E},            // 0xA3
		{"AND H", 1, cpu.AND_H},            // 0xA4
		{"AND L", 1, cpu.AND_L},            // 0xA5
		{"AND (HL)", 1, cpu.AND_HL},        // 0xA6
		{"AND A", 1, cpu.AND_A},            // 0xA7
		{"XOR B", 1, cpu.XOR_B},            // 0xA8
		{"XOR C", 1, cpu.XOR_C},            // 0xA9
		{"XOR D", 1, cpu.XOR_D},            // 0xAA
		{"XOR E", 1, cpu.XOR_E},            // 0xAB
		{"XOR H", 1, cpu.XOR_H},            // 0xAC
		{"XOR L", 1, cpu.XOR_L},            // 0xAD
		{"XOR (HL)", 1, cpu.XOR_HL},        // 0xAE
		{"XOR A", 1, cpu.XOR_A},            // 0xAF
		{"OR B", 1, cpu.OR_B},              // 0xB0
		{"OR C", 1, cpu.OR_C},              // 0xB1
		{"OR D", 1, cpu.OR_D},              // 0xB2
		{"OR E", 1, cpu.OR_E},              // 0xB3
		{"OR H", 1, cpu.OR_H},              // 0xB4
		{"OR L", 1, cpu.OR_L},              // 0xB5
		{"OR (HL)", 1, cpu.OR_HL},          // 0xB6
		{"OR A", 1, cpu.OR_A},              // 0xB7
		{"CP B", 1, cpu.CP_B},              // 0xB8
		{"CP C", 1, cpu.CP_C},              // 0xB9
		{"CP D", 1, cpu.CP_D},              // 0xBA
		{"CP E", 1, cpu.CP_E},              // 0xBB
		{"CP H", 1, cpu.CP_H},              // 0xBC
		{"CP L", 1, cpu.CP_L},              // 0xBD
		{"CP (HL)", 1, cpu.CP_HL},          // 0xBE
		{"CP A", 1, cpu.CP_A},              // 0xBF
		// {"RET NZ", 1, cpu.RET_NZ},          // 0xC0
		// {"POP BC", 1, cpu.POP_BC},          // 0xC1
		// {"JP NZ, nn", 3, cpu.JP_NZ_nn},     // 0xC2

	}
}

func (cpu *CPU) CreateTicks(opcode uint8) {
	cpu.ticksTable = [256]uint8{
		2, 6, 4, 4, 2, 2, 4, 4, 10, 4, 4, 4, 2, 2, 4, 4, // 0x0_
		2, 6, 4, 4, 2, 2, 4, 4, 4, 4, 4, 4, 2, 2, 4, 4, // 0x1_
		0, 6, 4, 4, 2, 2, 4, 2, 0, 4, 4, 4, 2, 2, 4, 2, // 0x2_
		4, 6, 4, 4, 6, 6, 6, 2, 0, 4, 4, 4, 2, 2, 4, 2, // 0x3_
		2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0x4_
		2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0x5_
		2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0x6_
		4, 4, 4, 4, 4, 4, 2, 4, 2, 2, 2, 2, 2, 2, 4, 2, // 0x7_
		2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0x8_
		2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0x9_
		2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0xa_
		2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0xb_
		0, 6, 0, 6, 0, 8, 4, 8, 0, 2, 0, 0, 0, 6, 4, 8, // 0xc_
		0, 6, 0, 0, 0, 8, 4, 8, 0, 8, 0, 0, 0, 0, 4, 8, // 0xd_
		6, 6, 4, 0, 0, 8, 4, 8, 8, 2, 8, 0, 0, 0, 4, 8, // 0xe_
		6, 6, 4, 2, 0, 8, 4, 8, 6, 4, 8, 2, 0, 0, 4, 8, // 0xf_
	}
}

// Step uses the program counter to read an instruction from memory and executes it
func (cpu *CPU) Step() {

	// opcode for a specific instruction
	var opcode uint8

	if cpu.stopped {
		return
	}

	// Use the program counter to read the instruction byte from memory.
	opcode = cpu.mem.Read8(cpu.regs.pc)

	// Increment the program counter
	cpu.regs.pc++

	// Translate the byte to an instruction
	instruction := cpu.table[opcode]

	// If we can successfully translate the instruction, call our execute method
	// else panic which now returns the next program counter

	// check if the instruction is valid/not undefined
	// if instruction == (Instruction{}) {
	// 	return
	// }

	switch instruction.instuctionLength {
	case 0:
	case 1:
		instruction.execute(&OperandInfo{})

	case 2:
		operand := cpu.mem.Read8(cpu.regs.pc)
		cpu.regs.pc += uint16(operand)
		instruction.execute(&OperandInfo{operand8: operand})

	case 3:
		operand := cpu.mem.Read16(cpu.regs.pc)
		cpu.regs.pc += operand
		instruction.execute(&OperandInfo{operand16: operand})

	default:
		panic("Invalid instruction length")
	}

	// set ticks using ticks table
	cpu.ticks += uint32(cpu.ticksTable[opcode])

}

// Reset sets the CPU to a default state
func (cpu *CPU) Reset() {

	// TODO: create set af register function for this
	// cpu.regs.af = 0x01B0

	cpu.regs.SetBC(0x0013)
	cpu.regs.SetDE(0x00D8)
	cpu.regs.SetHL(0x014D)

	cpu.regs.sp = 0xFFFE
	cpu.regs.pc = 0x0100

	cpu.stopped = false
	cpu.ticks = 0

}

var CLOCK_SPEED uint32 = 4194304
var FRAME_RATE uint32 = 60
var CYCLES_PER_FRAME uint32 = CLOCK_SPEED / FRAME_RATE
