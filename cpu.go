package main

import "encoding/binary"

type CPU struct {
	registerA      uint8
	registerX      uint8
	status         uint8
	programCounter uint16
	memory         [0xffff]uint8
}

func NewCPU() CPU {
	return CPU{registerA: 0, status: 0, programCounter: 0}
}

// Utilities to manipulate memory

func (cpu CPU) memoryRead(address uint16) uint8 {
	return cpu.memory[address]
}

func (cpu CPU) memoryWrite(address uint16, data uint8) {
	cpu.memory[address] = data
}

func (cpu CPU) memoryReadU16(address uint16) uint16 {
	return binary.BigEndian.Uint16(cpu.memory[address : address+1])
}

func (cpu CPU) memoryWriteU16(address uint16, data uint16) {
	binary.LittleEndian.PutUint16(cpu.memory[address:address+1], data)
}

// Ops Code operations

func (cpu CPU) readNextProgramValue() uint8 {
	var value = cpu.memory[cpu.programCounter]
	cpu.programCounter += 1
	return value
}

func (cpu CPU) updateZeroAndNegativeFlags(result uint8) {
	if result == 0 {
		cpu.status = cpu.status | 0b0000_0010
	} else {
		cpu.status = cpu.status & 0b1111_1101
	}
	if result&0b1000_0000 != 0 {
		cpu.status = cpu.status | 0b1000_0000
	} else {
		cpu.status = cpu.status & 0b0111_1111
	}
}

func (cpu CPU) lda(param uint8) {
	cpu.registerA = param
	cpu.updateZeroAndNegativeFlags(cpu.registerA)
}

func (cpu CPU) tax() {
	cpu.registerX = cpu.registerA
	cpu.updateZeroAndNegativeFlags(cpu.registerX)
}

func (cpu CPU) inx() {
	cpu.registerX += 1
	cpu.updateZeroAndNegativeFlags(cpu.registerX)
}

// Load program and reset CPU

func (cpu CPU) load(program []uint8) {
	copy(cpu.memory[0x8000:0xffff], program)
	cpu.memoryWriteU16(0xfffc, 0x8000)
}

func (cpu CPU) reset() {
	cpu.registerA = 0
	cpu.registerX = 0
	cpu.status = 0
	cpu.programCounter = cpu.memoryReadU16(0xfffc)
}

func (cpu CPU) run() {
	for {
		var opsCode = cpu.readNextProgramValue()
		switch opsCode {
		case 0xA9: // LDA
			var param = cpu.readNextProgramValue()
			cpu.lda(param)
		case 0xAA: // TAX
			cpu.tax()
		case 0xe8: // INX
			cpu.inx()
		case 0x00: // BRK
			return
		default:
		}
	}
}

func (cpu CPU) loadAndRUn(program []uint8) {
	cpu.load(program)
	cpu.reset()
	cpu.run()
}
