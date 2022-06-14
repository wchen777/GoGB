package gb

/*

GameBoy Memory Areas

$FFFF	Interrupt Enable Flag
$0000 – 3FFF ROM0 Non-switchable ROM Bank.
$4000 – 7FFF ROMX Switchable ROM bank.

$8000 – 9FFF VRAM Video RAM, switchable (0-1) in GBC mode.

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

type Memory struct {
	cart [0x8000]uint8
	sram [0x2000]uint8
	vram [0x2000]uint8
	wram [0x2000]uint8
	oam  [0x100]uint8
	hram [0x80]uint8
	io   [0x100]uint8
}

// Reads and Writes

// Write an 8-bit value to the address
func (mem *Memory) Write8(address uint16, value uint8) {

}

// Read an 8-bit value from the address
func (mem *Memory) Read8(address uint16) uint8 {
	return 0
}

// Write a 16-bit value to the address
func (mem *Memory) Write16(address uint16, value uint16) {

}

// Read a 16-bit value from the address
func (mem *Memory) Read16(address uint16) uint16 {
	return 0
}

//
func (mem *Memory) WriteToStack16(value uint16, sp *uint16) {
	// decrement sp by 2
	*sp -= 2

	// call write 16 to sp
}
