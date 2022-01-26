package gb

/*

GameBoy Memory Areas

$FFFF	Interrupt Enable Flag
$FF80-$FFFE	Zero Page - 127 bytes
$FF00-$FF7F	Hardware I/O Registers
$FEA0-$FEFF	Unusable Memory
$FE00-$FE9F	OAM - Object Attribute Memory
$E000-$FDFF	Echo RAM - Reserved, Do Not Use
$D000-$DFFF	Internal RAM - Bank 1-7 (switchable - CGB only)
$C000-$CFFF	Internal RAM - Bank 0 (fixed)
$A000-$BFFF	Cartridge RAM (If Available)
$9C00-$9FFF	BG Map Data 2
$9800-$9BFF	BG Map Data 1
$8000-$97FF	Character RAM
$4000-$7FFF	Cartridge ROM - Switchable Banks 1-xx
$0150-$3FFF	Cartridge ROM - Bank 0 (fixed)
$0100-$014F	Cartridge Header Area
$0000-$00FF	Restart and Interrupt Vectors

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
