package cpu

import (
	"encoding/binary"
	"fmt"
)

const STACK_BASE uint16 = 0x0100
const STACK_RESET uint8 = 0xfd

type CPU struct {
	registerA    uint8
	registerX    uint8
	registerY    uint8
	stackPointer uint8
	// Memory space [0x0100 .. 0x1FF] is used for stack
	// The stack pointer holds the address of the top of that space.
	// NES Stack (as all stacks) grows from top to bottom
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

func (cpu CPU) isCarryFlagSet() bool {
	return cpu.statusFlags&0b0000_0001 != 0
}

func (cpu CPU) updateCarryFlagForResult(hasCarry bool) {
	if hasCarry {
		cpu.setCarryFlag()
	} else {
		cpu.unsetCarryFlag()
	}
}

func (cpu CPU) setZeroFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0000_0010
}

func (cpu CPU) unsetZeroFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1111_1101
}

func (cpu CPU) isZeroFlagSet() bool {
	return cpu.statusFlags&0b0000_0010 != 0
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

func (cpu CPU) isInterruptDisableFlagSet() bool {
	return cpu.statusFlags&0b0000_0100 != 0
}

func (cpu CPU) setDecimalFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0000_1000
}

func (cpu CPU) unsetDecimalFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1111_0111
}

func (cpu CPU) isDecimalFlagSet() bool {
	return cpu.statusFlags&0b0000_1000 != 0
}

func (cpu CPU) setOverflowFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b0100_0000
}

func (cpu CPU) unsetOverflowFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b1011_1111
}

func (cpu CPU) isOverflowFlagSet() bool {
	return cpu.statusFlags&0b0100_0000 != 0
}

func (cpu CPU) updateOverflowFlagForResult(hasOverflow bool) {
	if hasOverflow {
		cpu.setOverflowFlag()
	} else {
		cpu.unsetOverflowFlag()
	}
}

func (cpu CPU) setNegativeFlag() {
	cpu.statusFlags = cpu.statusFlags | 0b1000_0000
}

func (cpu CPU) unsetNegativeFlag() {
	cpu.statusFlags = cpu.statusFlags & 0b0111_1111
}

func (cpu CPU) isNegativeFlagSet() bool {
	return cpu.statusFlags&0b1000_0000 != 0
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
	return binary.LittleEndian.Uint16(cpu.memory[address : address+1])
}

func (cpu CPU) memoryWriteU16(address uint16, data uint16) {
	binary.LittleEndian.PutUint16(cpu.memory[address:address+1], data)
}

// Stack helpers

func (cpu CPU) pushStack(value uint8) {
	cpu.memoryWrite(STACK_BASE+uint16(cpu.stackPointer), value)
	cpu.stackPointer -= 1
}

func (cpu CPU) pullStack() uint8 {
	cpu.stackPointer += 1
	return cpu.memoryRead(STACK_BASE + uint16(cpu.stackPointer))
}

// This does not get the operand but the address of the operand, which will be the retrieved using memory read
func (cpu CPU) getOperandAddress(mode AddressingMode) uint16 {
	switch mode {
	case Immediate:
		return cpu.programCounter
	case Relative:
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
	case Implied:
		panic("trying to resolve implicit addressing mode")
	default:
		panic(fmt.Sprintf("addressing mode %v is not supported", mode))
	}
}

// Helpers for Ops Code operations

// http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
func (cpu CPU) addWithCarry(a uint8, b uint8, carry bool) (uint8, bool, bool) {
	var sum = uint16(a) + uint16(b)
	if carry {
		sum += 1
	}
	var hasCarry = sum>>8 != 0
	var result = uint8(sum)
	var hasOverflow = (a^result)&(b^result)&0x80 != 0
	return result, hasCarry, hasOverflow
}

func (cpu CPU) branch(condition bool) {
	if condition {
		var operandAddress = cpu.getOperandAddress(Relative)
		var operand = cpu.memoryRead(operandAddress)
		isPositive := operand&0b1000_0000 == 0
		if isPositive {
			cpu.programCounter += uint16(operand)
		} else {
			cpu.programCounter -= uint16(operand)
		}
	}
}

// Ops code operations

func (cpu CPU) adc(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	result, hasCarry, hasOverflow := cpu.addWithCarry(cpu.registerA, operand, cpu.isCarryFlagSet())
	cpu.registerA = result
	cpu.updateCarryFlagForResult(hasCarry)
	cpu.updateOverflowFlagForResult(hasOverflow)
	cpu.updateZeroFlagForResult(cpu.registerA)
	cpu.updateNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) and(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.registerA & cpu.memoryRead(operandAddress)
	cpu.updateNegativeFlagForResult(cpu.registerA)
	cpu.updateZeroFlagForResult(cpu.registerA)
}

func (cpu CPU) asl() {
	// TODO
}

func (cpu CPU) bcc() {
	cpu.branch(!cpu.isCarryFlagSet())
}

func (cpu CPU) bcs() {
	cpu.branch(cpu.isCarryFlagSet())
}

func (cpu CPU) beq() {
	cpu.branch(cpu.isZeroFlagSet())
}

func (cpu CPU) bit() {
	// TODO
}

func (cpu CPU) bmi() {
	cpu.branch(cpu.isNegativeFlagSet())
}

func (cpu CPU) bne() {
	cpu.branch(!cpu.isZeroFlagSet())
}

func (cpu CPU) bpl() {
	cpu.branch(!cpu.isNegativeFlagSet())
}

func (cpu CPU) bvc() {
	cpu.branch(!cpu.isOverflowFlagSet())
}

func (cpu CPU) bvs() {
	cpu.branch(cpu.isOverflowFlagSet())
}

func (cpu CPU) clc() {
	cpu.unsetDecimalFlag()
}

func (cpu CPU) cld() {
	cpu.unsetDecimalFlag()
}

func (cpu CPU) cli() {
	cpu.unsetInterruptDisableFlag()
}

func (cpu CPU) clv() {
	cpu.unsetOverflowFlag()
}

func (cpu CPU) cmp() {
	// TODO
}

func (cpu CPU) cpx() {
	// TODO
}

func (cpu CPU) cpy() {
	// TODO
}

func (cpu CPU) dec(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	var result = operand - 1
	cpu.memoryWrite(operandAddress, result)
	cpu.updateNegativeFlagForResult(result)
	cpu.updateZeroFlagForResult(result)
}

func (cpu CPU) dex() {
	cpu.registerX -= 1
	cpu.updateNegativeFlagForResult(cpu.registerX)
	cpu.updateZeroFlagForResult(cpu.registerX)
}

func (cpu CPU) dey() {
	cpu.registerY -= 1
	cpu.updateNegativeFlagForResult(cpu.registerY)
	cpu.updateZeroFlagForResult(cpu.registerY)
}

func (cpu CPU) eor(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.registerA ^ cpu.memoryRead(operandAddress)
	cpu.updateNegativeFlagForResult(cpu.registerA)
	cpu.updateZeroFlagForResult(cpu.registerA)
}

func (cpu CPU) inc(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	var result = operand + 1
	cpu.memoryWrite(operandAddress, result)
	cpu.updateNegativeFlagForResult(result)
	cpu.updateZeroFlagForResult(result)
}

func (cpu CPU) inx() {
	cpu.registerX += 1
	cpu.updateNegativeFlagForResult(cpu.registerX)
	cpu.updateZeroFlagForResult(cpu.registerX)
}

func (cpu CPU) iny() {
	cpu.registerY += 1
	cpu.updateNegativeFlagForResult(cpu.registerY)
	cpu.updateZeroFlagForResult(cpu.registerY)
}

func (cpu CPU) jmp() {
	// TODO
}

func (cpu CPU) jsr() {
	// TODO
}

func (cpu CPU) lda(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.memoryRead(operandAddress)
	cpu.updateNegativeFlagForResult(cpu.registerA)
	cpu.updateZeroFlagForResult(cpu.registerA)
}

func (cpu CPU) ldx(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerX = cpu.memoryRead(operandAddress)
	cpu.updateNegativeFlagForResult(cpu.registerX)
	cpu.updateZeroFlagForResult(cpu.registerX)
}

func (cpu CPU) ldy(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerY = cpu.memoryRead(operandAddress)
	cpu.updateNegativeFlagForResult(cpu.registerY)
	cpu.updateZeroFlagForResult(cpu.registerY)
}

func (cpu CPU) lsr() {
	// TODO
}

func (cpu CPU) nop() {}

func (cpu CPU) ora(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.registerA | cpu.memoryRead(operandAddress)
	cpu.updateNegativeFlagForResult(cpu.registerA)
	cpu.updateZeroFlagForResult(cpu.registerA)
}

func (cpu CPU) pha() {
	cpu.pushStack(cpu.registerA)
}

func (cpu CPU) php() {
	cpu.pushStack(cpu.statusFlags)
}

func (cpu CPU) pla() {
	cpu.registerA = cpu.pullStack()
	cpu.updateZeroFlagForResult(cpu.registerA)
	cpu.updateNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) plp() {
	cpu.statusFlags = cpu.pullStack()
}

func (cpu CPU) rol() {
	// TODO
}

func (cpu CPU) ror() {
	// TODO
}

func (cpu CPU) rti() {
	// TODO
}

func (cpu CPU) rts() {
	// TODO
}

func (cpu CPU) sbc(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	result, hasCarry, hasOverflow := cpu.addWithCarry(cpu.registerA, ^operand+1, cpu.isCarryFlagSet())
	cpu.registerA = result
	cpu.updateCarryFlagForResult(hasCarry)
	cpu.updateOverflowFlagForResult(hasOverflow)
	cpu.updateZeroFlagForResult(cpu.registerA)
	cpu.updateNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) sec() {
	cpu.setCarryFlag()
}

func (cpu CPU) sed() {
	cpu.setDecimalFlag()
}

func (cpu CPU) sei() {
	cpu.setInterruptDisableFlag()
}

func (cpu CPU) sta(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.memoryWrite(operandAddress, cpu.registerA)
}

func (cpu CPU) stx(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.memoryWrite(operandAddress, cpu.registerX)
}

func (cpu CPU) sty(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.memoryWrite(operandAddress, cpu.registerY)
}

func (cpu CPU) tax() {
	cpu.registerX = cpu.registerA
	cpu.updateNegativeFlagForResult(cpu.registerX)
	cpu.updateZeroFlagForResult(cpu.registerX)
}

func (cpu CPU) tay() {
	cpu.registerY = cpu.registerA
	cpu.updateNegativeFlagForResult(cpu.registerY)
	cpu.updateZeroFlagForResult(cpu.registerY)
}

func (cpu CPU) tsx() {
	cpu.registerX = cpu.stackPointer
	cpu.updateNegativeFlagForResult(cpu.registerX)
	cpu.updateZeroFlagForResult(cpu.registerX)
}

func (cpu CPU) txa() {
	cpu.registerA = cpu.registerX
	cpu.updateNegativeFlagForResult(cpu.registerA)
	cpu.updateZeroFlagForResult(cpu.registerA)
}

func (cpu CPU) txs() {
	cpu.stackPointer = cpu.registerX
}

func (cpu CPU) tya() {
	cpu.registerA = cpu.registerY
	cpu.updateNegativeFlagForResult(cpu.registerA)
	cpu.updateZeroFlagForResult(cpu.registerA)
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
	cpu.stackPointer = STACK_RESET
	cpu.programCounter = cpu.memoryReadU16(0xFFFC)
}

func (cpu CPU) run() {
	for {
		var hexCode = cpu.memory[cpu.programCounter]
		cpu.programCounter += 1
		var opCode = matchHexCodeWithOpsCode(hexCode)
		switch opCode.operation {
		case ADC:
			cpu.adc(opCode.addressingMode)
		case AND:
			cpu.and(opCode.addressingMode)
		case ASL:
			cpu.asl()
		case BCC:
			cpu.bcc()
		case BCS:
			cpu.bcs()
		case BEQ:
			cpu.beq()
		case BIT:
			cpu.bit()
		case BMI:
			cpu.bmi()
		case BNE:
			cpu.bne()
		case BPL:
			cpu.bpl()
		case BRK:
			return
		case BVS:
			cpu.bvs()
		case BVC:
			cpu.bvc()
		case CLC:
			cpu.clc()
		case CLD:
			cpu.cld()
		case CLI:
			cpu.cli()
		case CLV:
			cpu.clv()
		case CMP:
			cpu.cmp()
		case CPX:
			cpu.cpx()
		case CPY:
			cpu.cpy()
		case DEC:
			cpu.dec(opCode.addressingMode)
		case DEX:
			cpu.dex()
		case DEY:
			cpu.dey()
		case EOR:
			cpu.eor(opCode.addressingMode)
		case INC:
			cpu.inc(opCode.addressingMode)
		case INX:
			cpu.inx()
		case INY:
			cpu.iny()
		case JMP:
			cpu.jmp()
		case JSR:
			cpu.jsr()
		case LDA:
			cpu.lda(opCode.addressingMode)
		case LDX:
			cpu.ldx(opCode.addressingMode)
		case LDY:
			cpu.ldy(opCode.addressingMode)
		case LSR:
			cpu.lsr()
		case NOP:
			cpu.nop()
		case ORA:
			cpu.ora(opCode.addressingMode)
		case PHA:
			cpu.pha()
		case PHP:
			cpu.php()
		case PLA:
			cpu.pla()
		case PLP:
			cpu.plp()
		case ROL:
			cpu.rol()
		case ROR:
			cpu.ror()
		case RTI:
			cpu.rti()
		case RTS:
			cpu.rts()
		case SBC:
			cpu.sbc(opCode.addressingMode)
		case SEC:
			cpu.sec()
		case SED:
			cpu.sed()
		case SEI:
			cpu.sei()
		case STA:
			cpu.sta(opCode.addressingMode)
		case STX:
			cpu.stx(opCode.addressingMode)
		case STY:
			cpu.sty(opCode.addressingMode)
		case TAX:
			cpu.tax()
		case TAY:
			cpu.tay()
		case TSX:
			cpu.tsx()
		case TXA:
			cpu.txa()
		case TXS:
			cpu.txs()
		case TYA:
			cpu.tya()
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
