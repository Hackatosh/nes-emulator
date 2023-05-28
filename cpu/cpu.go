package cpu

import (
	"encoding/binary"
	"fmt"
)

type CPU struct {
	registerA   uint8
	registerX   uint8
	registerY   uint8
	statusFlags uint8
	// Status flags :
	// 7  bit  0
	// ---- ----
	// NVss DIZC
	// |||| ||||
	// |||| |||+- Carry
	// |||| ||+-- Zero
	// |||| |+--- Interrupt Disable
	// |||| +---- Decimal
	// ||++------ No CPU effect, see: the B flag
	// |+-------- Overflow
	// +--------- Negative
	programCounter uint16
	memory         [0xffff]uint8
}

// Flags Helpers

func (cpu CPU) setCarryFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0000_0001
}

func (cpu CPU) unsetCarryFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1111_1110
}

func (cpu CPU) setZeroFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0000_0010
}

func (cpu CPU) unsetZeroFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1111_1101
}

func (cpu CPU) updateZeroFlagForResult(result uint8) {
	if result == 0 {
		cpu.setZeroFlag()
	} else {
		cpu.unsetZeroFlag()
	}
}

func (cpu CPU) setInterruptDisableFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0000_0100
}

func (cpu CPU) unsetInterruptDisableFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1111_1011
}

func (cpu CPU) setDecimalFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0000_1000
}

func (cpu CPU) unsetDecimalFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1111_0111
}

func (cpu CPU) setOverflowFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0100_0000
}

func (cpu CPU) unsetOverflowFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1011_1111
}

func (cpu CPU) setNegativeFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b1000_0000
}

func (cpu CPU) unsetNegativeFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b0111_1111
}

func (cpu CPU) updateNegativeFlagForResult(result uint8) {
	if result&0b1000_0000 != 0 {
		cpu.setNegativeFlag()
	} else {
		cpu.unsetNegativeFlag()
	}
}

// Memory helpers

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

// This does not get the operand but the address of the operand, which will be the retrieved using memory read
func (cpu CPU) getOperandAddress(mode AddressingMode) uint16 {
	switch mode {
	case Immediate:
		return cpu.programCounter
	case ZeroPage:
		// It's only a 8 bits address with Zero Page, so you can only get an address in the first 256 memory cells
		// But it's faster !
		return uint16(cpu.memoryRead(cpu.programCounter))
	case ZeroPageX:
		var pos = cpu.memoryRead(cpu.programCounter)
		return uint16(pos + cpu.registerX)
	case ZeroPageY:
		var pos = cpu.memoryRead(cpu.programCounter)
		return uint16(pos + cpu.registerY)
	case Absolute:
		return cpu.memoryReadU16(cpu.programCounter)
	case AbsoluteX:
		var pos = cpu.memoryReadU16(cpu.programCounter)
		return pos + uint16(cpu.registerX)
	case AbsoluteY:
		var pos = cpu.memoryReadU16(cpu.programCounter)
		return pos + uint16(cpu.registerY)
	case Indirect:
		var ref = cpu.memoryReadU16(cpu.programCounter)
		return cpu.memoryReadU16(ref)
	case IndirectX:
		var base = cpu.memoryRead(cpu.programCounter)
		return cpu.memoryReadU16(uint16(base + cpu.registerX))
	case IndirectY:
		var ref = cpu.memoryReadU16(cpu.programCounter)
		return cpu.memoryReadU16(ref) + uint16(cpu.registerY)
	default:
		panic(fmt.Sprintf("addressing mode %v is not supported", mode))
	}
}

// Ops Code operations

func (cpu CPU) lda(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.memoryRead(operandAddress)
	cpu.updateNegativeFlagForResult(cpu.registerA)
	cpu.updateZeroFlagForResult(cpu.registerA)
}

func (cpu CPU) tax() {
	cpu.registerX = cpu.registerA
	cpu.updateNegativeFlagForResult(cpu.registerX)
	cpu.updateZeroFlagForResult(cpu.registerX)
}

func (cpu CPU) inx() {
	cpu.registerX += 1
	cpu.updateNegativeFlagForResult(cpu.registerX)
	cpu.updateZeroFlagForResult(cpu.registerX)
}

func (cpu CPU) sta(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.memoryWrite(operandAddress, cpu.registerA)
}

// Load program and reset CPU

func (cpu CPU) load(program []uint8) {
	copy(cpu.memory[0x8000:0xFFFF], program)
	cpu.memoryWriteU16(0xFFFC, 0x8000)
}

func (cpu CPU) reset() {
	cpu.registerA = 0
	cpu.registerX = 0
	cpu.registerY = 0
	cpu.statusFlags = 0
	cpu.programCounter = cpu.memoryReadU16(0xFFFC)
}

func (cpu CPU) run() {
	for {
		var hexCode = cpu.memory[cpu.programCounter]
		cpu.programCounter += 1
		opCode := matchHexCodeWithOpsCode(hexCode)
		switch opCode.operation {
		case STA:
			cpu.sta(opCode.addressingMode)
		case LDA:
			cpu.lda(opCode.addressingMode)
		case TAX:
			cpu.tax()
		case INX:
			cpu.inx()
		case BRK:
			return
		default:
			panic(fmt.Sprintf("operation %v is unsupported", opCode.operation))
		}
		cpu.programCounter += opCode.bytes - 1
	}
}

func (cpu CPU) loadAndRUn(program []uint8) {
	cpu.load(program)
	cpu.reset()
	cpu.run()
}
