package gb

// The Console puts all the Gameboy parts together.

type Console struct {
	cpu *CPU // Gameboy CPU
	ppu *PPU // Gameboy PPU
	apu *APU // Gameboy APU

}

func (c *Console) Step() int {
	return 0
}

func (c *Console) Save() {

}

func (c *Console) Load() {

}
