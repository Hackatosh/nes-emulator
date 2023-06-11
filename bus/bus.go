package bus

import (
	"encoding/binary"
	"fmt"
)

const CPU_RAM_START uint16 = 0x0000
const CPU_RAM_MIRRORS_END uint16 = 0x1FFF
const PPU_REGISTERS_START uint16 = 0x2000
const PPU_REGISTERS_MIRRORS_END uint16 = 0x3FFF
const PRG_ROM_START uint16 = 0x8000
const PRG_ROM_END uint16 = 0xFFFF

type Bus struct {
	rom    *Rom
	memory [0xffff]uint8
	// More info on memory structure here : https://www.nesdev.org/wiki/CPU_memory_map
}

// Memory helpers

func (bus *Bus) readPrgROM(address uint16) uint8 {
	var unmirroredAddress = address - 0x8000
	// Unmirroring if prgRom is of 16 KiB (we map 32 KiB addressing space)
	if len(bus.rom.prgRom) == 0x4000 && address >= 0x4000 {
		unmirroredAddress = address % 0x4000
	}
	return bus.rom.prgRom[unmirroredAddress]
}

func (bus *Bus) MemoryRead(address uint16) uint8 {
	var unmirroredAddress uint16
	switch {
	case CPU_RAM_START <= address && address <= CPU_RAM_MIRRORS_END:
		unmirroredAddress = address & 0b00000111_11111111
		return bus.memory[unmirroredAddress]
	case PPU_REGISTERS_START <= address && address <= PPU_REGISTERS_MIRRORS_END:
		unmirroredAddress = address & 0b00100000_00000111
		return bus.memory[unmirroredAddress]
	case PRG_ROM_START <= address && address <= PRG_ROM_END:
		return bus.readPrgROM(address)
	default:
		panic(fmt.Sprintf("Unsupported address %v", address))
	}
}

func (bus *Bus) MemoryWrite(address uint16, data uint8) {
	var unmirroredAddress uint16
	switch {
	case CPU_RAM_START <= address && address <= CPU_RAM_MIRRORS_END:
		unmirroredAddress = address & 0b00000111_11111111
	case PPU_REGISTERS_START <= address && address <= PPU_REGISTERS_MIRRORS_END:
		unmirroredAddress = address & 0b00100000_00000111
	case PRG_ROM_START <= address && address <= PRG_ROM_END:
		panic(fmt.Sprintf("Trying to write to address %v in PRG ROM", address))
	default:
		panic(fmt.Sprintf("Unsupported address %v", address))
	}

	bus.memory[unmirroredAddress] = data
}

// TODO : Some edge case here !
// What if address is CPU_RAM_MIRRORS_END ?? This will bug a lot
func (bus *Bus) MemoryReadU16(address uint16) uint16 {
	return binary.LittleEndian.Uint16([]uint8{bus.MemoryRead(address), bus.MemoryRead(address + 1)})
}

// TODO : Some edge case here !
// What if address is CPU_RAM_MIRRORS_END ?? This will bug a lot
func (bus *Bus) MemoryWriteU16(address uint16, data uint16) {
	bytes := make([]uint8, 2)
	binary.LittleEndian.PutUint16(bytes, data)
	bus.MemoryWrite(address, bytes[0])
	bus.MemoryWrite(address+1, bytes[1])
}

func NewBus() Bus {
	return Bus{
		memory: [0xffff]uint8{},
	}
}

func (bus *Bus) LoadRom(rom *Rom) {
	bus.rom = rom

}
