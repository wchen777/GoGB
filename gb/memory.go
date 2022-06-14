package gb

/*

GameBoy Memory Areas

0000 – 3FFF ROM0 Non-switchable ROM Bank.
4000 – 7FFF ROMX Switchable ROM bank.

8000 – 9FFF VRAM Video RAM, switchable (0-1) in GBC mode.

A000 – BFFF SRAM External RAM in cartridge, often battery buffered.
C000 – CFFF WRAM0 Work RAM.
D000 – DFFF WRAMX Work RAM, switchable (1-7) in GBC mode

E000 – FDFF ECHO Description of the behaviour below.

FE00 – FE9F OAM (Object Attribute Table) Sprite information table.
FEA0 – FEFF UNUSED Description of the behaviour below.
FF00 – FF7F I/O Registers I/O registers are mapped here.
FF80 – FFFE HRAM Internal CPU RAM

FFFF - IE Register Interrupt enable flags.


*/

// 64kb memory map

// type Memory struct {
// 	cart [0x8000]uint8
// 	sram [0x2000]uint8
// 	vram [0x2000]uint8
// 	wram [0x2000]uint8
// 	oam  [0x100]uint8
// 	hram [0x80]uint8
// 	io   [0x100]uint8
// }

type MemoryMap struct {
	console *Console // not sure if we need this?
	cart    [0x8000]uint8
	vram    [0x2000]uint8
	sram    [0x2000]uint8
	wram    [0x2000]uint8
	oam     [0x100]uint8
	hram    [0x80]uint8
	io      [0x100]uint8
}

const ROM_END = 0x8000
const VRAM_END = 0xA000
const SRAM_END = 0xC000
const WRAM_END = 0xD000
const ECHO_END = 0xFE00
const OAM_END = 0xFEA0
const UNUSED_END = 0xFF00
const IO_END = 0xFF80
const HRAM_END = 0xFFFF

// Reads and Writes, take in any 16-bit address and delegate to the correct memory area

// Write an 8-bit value to the address
func (mem *MemoryMap) Write8(address uint16, value uint8) {
	switch {
	case address < ROM_END:
		// cart
	case address < VRAM_END:
		// vram
	case address < SRAM_END:
		// sram
	case address < WRAM_END:
		// wram
	case address < ECHO_END:
		// echo
	case address < OAM_END:
		// oam
	case address < UNUSED_END:
		// unused
	case address < IO_END:
		// io
	case address < HRAM_END:
	// hram
	case address == 0xFFFF:
		// interrupt flag
	default:
		panic("Invalid memory address")
	}
}

// Read an 8-bit value from the address
func (mem *MemoryMap) Read8(address uint16) uint8 {
	switch {
	case address < ROM_END:
		// cart
		return mem.cart[address]
	case address < VRAM_END:
		// vram
		return mem.vram[address-ROM_END]
	case address < SRAM_END:
		// sram
		return mem.sram[address-VRAM_END]
	case address < WRAM_END:
		// wram
		return mem.wram[address-SRAM_END]
	case address < ECHO_END:
		// echo
		return mem.wram[address-WRAM_END]
	case address < OAM_END:
		// oam
		return mem.oam[address-ECHO_END]
	case address < UNUSED_END:
		// unused
		// what to do here?
		return 0
	case address < IO_END:
		// io
		return mem.io[address-UNUSED_END]
	case address < HRAM_END:
		// hram
		return mem.hram[address-IO_END]
	case address == 0xFFFF:
		// interrupt flag
		return 0 // TODO: interrupts?
	default:
		panic("Invalid memory address")
	}
}

// Write a 16-bit value to the address
func (mem *MemoryMap) Write16(address uint16, value uint16) {

}

// Read a 16-bit value from the address
func (mem *MemoryMap) Read16(address uint16) uint16 {
	// read twice, at address and address+1
	// return the double word value
	// use Read8 to read the low and high bytes
	return 0
}

//
func (mem *MemoryMap) WriteToStack16(value uint16, sp *uint16) {
	// decrement sp by 2
	*sp -= 2

	// call write 16 to sp
}
