package bus

import (
	"encoding/binary"
	"fmt"
)

const CPU_RAM_START uint16 = 0x0000
const CPU_RAM_MIRRORS_END uint16 = 0x1FFF
const PPU_REGISTERS_START uint16 = 0x2000
const PPU_REGISTERS_MIRRORS_END uint16 = 0x3FFF

type Bus struct {
	memory [0xffff]uint8
	// More info on memory structure here : https://www.nesdev.org/wiki/CPU_memory_map
}

// Memory helpers

func (bus Bus) MemoryRead(address uint16) uint8 {
	var unmirroredAddress uint16
	switch true {
	case CPU_RAM_START <= address && address <= CPU_RAM_MIRRORS_END:
		unmirroredAddress = address & 0b00000111_11111111
	case PPU_REGISTERS_START <= address && address <= PPU_REGISTERS_MIRRORS_END:
		unmirroredAddress = address & 0b00100000_00000111
	default:
		panic(fmt.Sprintf("Unsupported address %v", address))
	}

	return bus.memory[unmirroredAddress]
}

func (bus Bus) MemoryWrite(address uint16, data uint8) {
	var unmirroredAddress uint16
	switch true {
	case CPU_RAM_START <= address && address <= CPU_RAM_MIRRORS_END:
		unmirroredAddress = address & 0b00000111_11111111
	case PPU_REGISTERS_START <= address && address <= PPU_REGISTERS_MIRRORS_END:
		unmirroredAddress = address & 0b00100000_00000111
	default:
		panic(fmt.Sprintf("Unsupported address %v", address))
	}

	bus.memory[unmirroredAddress] = data
}

// TODO : Some edge case here !
// What if address is CPU_RAM_MIRRORS_END ?? This will bug a lot
func (bus Bus) MemoryReadU16(address uint16) uint16 {
	return binary.LittleEndian.Uint16([]uint8{bus.MemoryRead(address), bus.MemoryRead(address + 1)})
}

// TODO : Some edge case here !
// What if address is CPU_RAM_MIRRORS_END ?? This will bug a lot
func (bus Bus) MemoryWriteU16(address uint16, data uint16) {
	bytes := make([]uint8, 2)
	binary.LittleEndian.PutUint16(bytes, data)
	bus.MemoryWrite(address, bytes[0])
	bus.MemoryWrite(address+1, bytes[1])
}

func (bus Bus) LoadProgram(program []uint8) {
	copy(bus.memory[0x8000:0xFFFF], program)
}

func NewBus() Bus {
	return Bus{
		memory: [0xffff]uint8{},
	}
}
