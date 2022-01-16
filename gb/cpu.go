package gb

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
func (cpu *CPU) LD_HLp_A(stepInfo *OperandInfo) {
	cpu.mem.Write8(cpu.regs.GetHL(), cpu.regs.a)
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
func (cpu *CPU) LD_A_HLp(stepInfo *OperandInfo) {
	cpu.regs.a = cpu.mem.Read8(cpu.regs.GetHL())
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
		{"LD (HL+), A", 1, cpu.LD_HLp_A},   // 0x22
		{"INC HL", 1, cpu.INC_HL},          // 0x23
		{"INC H", 1, cpu.INC_H},            // 0x24
		{"DEC H", 1, cpu.DEC_H},            // 0x25
		{"LD H, d8", 2, cpu.LD_H_d8},       // 0x26
		{"DAA", 1, cpu.DAA},                // 0x27
		{"JR Z, r8", 2, cpu.JR_Z_r8},       // 0x28
		{"ADD HL, HL", 1, cpu.ADD_HL_HL},   // 0x29
		{"LD A, (HL+)", 1, cpu.LD_A_HLp},   // 0x2A
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
}

var CLOCK_SPEED uint32 = 4194304
var FRAME_RATE uint32 = 60
var CYCLES_PER_FRAME uint32 = CLOCK_SPEED / FRAME_RATE
