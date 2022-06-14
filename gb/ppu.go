package gb

// import (
// 	"encoding/gob"
// 	"image"
// )

// The PPU, or pixel processing unit, is used to render the Gameboy screen and process graphics.

// The Gameboy’s screen resolution is 160x144 pixels.

/*
The Gameboy has 3 distinct video layers that are all made up of 8x8 pixel tiles.
- Background
- Window
- Sprites
*/

type PPU struct {
	mem     MemoryMap // memory map interface
	console *Console  // reference to parent console
}
