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

// Generic helpers

func isNegative(value uint8) bool {
	return value&0b1000_0000 != 0
}

// Flags

type StatusFlag uint8

const (
	CARRY_FLAG             StatusFlag = 0b0000_0001
	ZERO_FLAG                         = 0b0000_0010
	INTERRUPT_DISABLE_FLAG            = 0b0000_0100
	DECIMAL_FLAG                      = 0b0000_1000
	BREAK_FLAG                        = 0b0001_0000
	BREAK_2_FLAG                      = 0b0010_0000
	OVERFLOW_FLAG                     = 0b0100_0000
	NEGATIVE_FLAG                     = 0b1000_0000
)

func (cpu CPU) setFlagToValue(statusFlag StatusFlag, value bool) {
	if value {
		cpu.statusFlags = cpu.statusFlags | uint8(statusFlag)
	} else {
		cpu.statusFlags = cpu.statusFlags ^ uint8(statusFlag)
	}
}

func (cpu CPU) isFlagSet(statusFlag StatusFlag) bool {
	return cpu.statusFlags&uint8(statusFlag) != 0
}

func (cpu CPU) setZeroFlagAndNegativeFlagForResult(result uint8) {
	cpu.setFlagToValue(ZERO_FLAG, result == 0)
	cpu.setFlagToValue(NEGATIVE_FLAG, isNegative(result))
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

func (cpu CPU) pushStackU16(value uint16) {
	var bytes = make([]uint8, 2)
	binary.LittleEndian.PutUint16(bytes, value)
	cpu.pushStack(bytes[0])
	cpu.pushStack(bytes[1])
}

func (cpu CPU) pullStackU16() uint16 {
	var bytes = make([]uint8, 2)
	bytes[1] = cpu.pullStack()
	bytes[0] = cpu.pullStack()
	return binary.LittleEndian.Uint16(bytes)
}

// This does not get the operand but the address of the operand, which will be the retrieved using memory read
func (cpu CPU) getOperandAddress(mode AddressingMode) uint16 {
	switch mode {
	case Implied:
		panic("trying to resolve implicit addressing mode")
	case Accumulator:
		panic("trying to resolve accumulator addressing mode")
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
			// 0x100 is 256
			cpu.programCounter += 0x100 - uint16(operand)
		}
	}
}

// Ops code operations

func (cpu CPU) adc(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	result, hasCarry, hasOverflow := cpu.addWithCarry(cpu.registerA, operand, cpu.isFlagSet(CARRY_FLAG))
	cpu.registerA = result
	cpu.setFlagToValue(CARRY_FLAG, hasCarry)
	cpu.setFlagToValue(OVERFLOW_FLAG, hasOverflow)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) and(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.registerA & cpu.memoryRead(operandAddress)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) asl(addressingMode AddressingMode) {
	if addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b1000_0000 != 0)
		cpu.registerA = cpu.registerA << 1
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operandAddress = cpu.getOperandAddress(addressingMode)
		var operand = cpu.memoryRead(operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b1000_0000 != 0)
		var result = operand << 1
		cpu.memoryWrite(operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu CPU) bcc() {
	cpu.branch(!cpu.isFlagSet(CARRY_FLAG))
}

func (cpu CPU) bcs() {
	cpu.branch(cpu.isFlagSet(CARRY_FLAG))
}

func (cpu CPU) beq() {
	cpu.branch(cpu.isFlagSet(ZERO_FLAG))
}

func (cpu CPU) bit(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	var result = operand & cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(result)
	cpu.setFlagToValue(OVERFLOW_FLAG, result&0b0100_0000 != 0)
}

func (cpu CPU) bmi() {
	cpu.branch(cpu.isFlagSet(NEGATIVE_FLAG))
}

func (cpu CPU) bne() {
	cpu.branch(!cpu.isFlagSet(ZERO_FLAG))
}

func (cpu CPU) bpl() {
	cpu.branch(!cpu.isFlagSet(NEGATIVE_FLAG))
}

func (cpu CPU) bvc() {
	cpu.branch(!cpu.isFlagSet(OVERFLOW_FLAG))
}

func (cpu CPU) bvs() {
	cpu.branch(cpu.isFlagSet(OVERFLOW_FLAG))
}

func (cpu CPU) clc() {
	cpu.setFlagToValue(CARRY_FLAG, false)
}

func (cpu CPU) cld() {
	cpu.setFlagToValue(DECIMAL_FLAG, false)
}

func (cpu CPU) cli() {
	cpu.setFlagToValue(INTERRUPT_DISABLE_FLAG, false)
}

func (cpu CPU) clv() {
	cpu.setFlagToValue(OVERFLOW_FLAG, false)
}

func (cpu CPU) compare(addressingMode AddressingMode, compareWith uint8) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	var result = compareWith - operand
	cpu.setZeroFlagAndNegativeFlagForResult(result)
	cpu.setFlagToValue(CARRY_FLAG, compareWith > operand)
}

func (cpu CPU) cmp(addressingMode AddressingMode) {
	cpu.compare(addressingMode, cpu.registerA)
}

func (cpu CPU) cpx(addressingMode AddressingMode) {
	cpu.compare(addressingMode, cpu.registerX)
}

func (cpu CPU) cpy(addressingMode AddressingMode) {
	cpu.compare(addressingMode, cpu.registerY)
}

func (cpu CPU) dec(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	var result = operand - 1
	cpu.memoryWrite(operandAddress, result)
	cpu.setZeroFlagAndNegativeFlagForResult(result)
}

func (cpu CPU) dex() {
	cpu.registerX -= 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu CPU) dey() {
	cpu.registerY -= 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu CPU) eor(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.registerA ^ cpu.memoryRead(operandAddress)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) inc(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	var result = operand + 1
	cpu.memoryWrite(operandAddress, result)
	cpu.setZeroFlagAndNegativeFlagForResult(result)
}

func (cpu CPU) inx() {
	cpu.registerX += 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu CPU) iny() {
	cpu.registerY += 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu CPU) jmp(addressingMode AddressingMode) {
	// TODO : some shady shit is done here in the tutorial, wtf ??
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.programCounter = operandAddress
}

func (cpu CPU) jsr(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	// +2 is for absolute read
	cpu.pushStackU16(cpu.programCounter + 2 - 1)
	cpu.programCounter = operandAddress
}

func (cpu CPU) lda(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.memoryRead(operandAddress)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) ldx(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerX = cpu.memoryRead(operandAddress)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu CPU) ldy(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerY = cpu.memoryRead(operandAddress)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu CPU) lsr(addressingMode AddressingMode) {
	if addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b0000_0001 != 0)
		cpu.registerA = cpu.registerA >> 1
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operandAddress = cpu.getOperandAddress(addressingMode)
		var operand = cpu.memoryRead(operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b0000_0001 != 0)
		var result = operand >> 1
		cpu.memoryWrite(operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu CPU) nop() {}

func (cpu CPU) ora(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	cpu.registerA = cpu.registerA | cpu.memoryRead(operandAddress)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) pha() {
	cpu.pushStack(cpu.registerA)
}

func (cpu CPU) php() {
	cpu.pushStack(cpu.statusFlags)
	cpu.setFlagToValue(BREAK_FLAG, false)
	cpu.setFlagToValue(BREAK_2_FLAG, true)
}

func (cpu CPU) pla() {
	cpu.registerA = cpu.pullStack()
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) plp() {
	cpu.statusFlags = cpu.pullStack() | BREAK_FLAG | BREAK_2_FLAG
}

func (cpu CPU) rol(addressingMode AddressingMode) {
	var carryMask uint8 = 0b0000_0000
	if cpu.isFlagSet(CARRY_FLAG) {
		carryMask = 0b0000_0001
	}
	if addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b1000_0000 != 0)
		cpu.registerA = (cpu.registerA << 1) | carryMask
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operandAddress = cpu.getOperandAddress(addressingMode)
		var operand = cpu.memoryRead(operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b1000_0000 != 0)
		var result = operand<<1 | carryMask
		cpu.memoryWrite(operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu CPU) ror(addressingMode AddressingMode) {
	var carryMask uint8 = 0b0000_0000
	if cpu.isFlagSet(CARRY_FLAG) {
		carryMask = 0b1000_0000
	}
	if addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b0000_0001 != 0)
		cpu.registerA = (cpu.registerA >> 1) | carryMask
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operandAddress = cpu.getOperandAddress(addressingMode)
		var operand = cpu.memoryRead(operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b0000_0001 != 0)
		var result = operand>>1 | carryMask
		cpu.memoryWrite(operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu CPU) rti() {
	cpu.statusFlags = cpu.pullStack()
	cpu.setFlagToValue(BREAK_FLAG, false)
	cpu.setFlagToValue(BREAK_2_FLAG, true)
	cpu.programCounter = cpu.pullStackU16()
}

func (cpu CPU) rts() {
	cpu.programCounter = cpu.pullStackU16() + 1
}

func (cpu CPU) sbc(addressingMode AddressingMode) {
	var operandAddress = cpu.getOperandAddress(addressingMode)
	var operand = cpu.memoryRead(operandAddress)
	result, hasCarry, hasOverflow := cpu.addWithCarry(cpu.registerA, ^operand+1, cpu.isFlagSet(CARRY_FLAG))
	cpu.registerA = result
	cpu.setFlagToValue(CARRY_FLAG, hasCarry)
	cpu.setFlagToValue(OVERFLOW_FLAG, hasOverflow)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) sec() {
	cpu.setFlagToValue(CARRY_FLAG, true)
}

func (cpu CPU) sed() {
	cpu.setFlagToValue(DECIMAL_FLAG, true)
}

func (cpu CPU) sei() {
	cpu.setFlagToValue(INTERRUPT_DISABLE_FLAG, true)
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
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu CPU) tay() {
	cpu.registerY = cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu CPU) tsx() {
	cpu.registerX = cpu.stackPointer
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu CPU) txa() {
	cpu.registerA = cpu.registerX
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu CPU) txs() {
	cpu.stackPointer = cpu.registerX
}

func (cpu CPU) tya() {
	cpu.registerA = cpu.registerY
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
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
		var programCounterBeforeOperation = cpu.programCounter
		var opCode = matchHexCodeWithOpsCode(hexCode)
		switch opCode.operation {
		case ADC:
			cpu.adc(opCode.addressingMode)
		case AND:
			cpu.and(opCode.addressingMode)
		case ASL:
			cpu.asl(opCode.addressingMode)
		case BCC:
			cpu.bcc()
		case BCS:
			cpu.bcs()
		case BEQ:
			cpu.beq()
		case BIT:
			cpu.bit(opCode.addressingMode)
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
			cpu.cmp(opCode.addressingMode)
		case CPX:
			cpu.cpx(opCode.addressingMode)
		case CPY:
			cpu.cpy(opCode.addressingMode)
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
			cpu.jmp(opCode.addressingMode)
		case JSR:
			cpu.jsr(opCode.addressingMode)
		case LDA:
			cpu.lda(opCode.addressingMode)
		case LDX:
			cpu.ldx(opCode.addressingMode)
		case LDY:
			cpu.ldy(opCode.addressingMode)
		case LSR:
			cpu.lsr(opCode.addressingMode)
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
			cpu.rol(opCode.addressingMode)
		case ROR:
			cpu.ror(opCode.addressingMode)
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
		// No jump or branch has occurred
		if programCounterBeforeOperation == cpu.programCounter {
			cpu.programCounter += getNumberOfBytesReadForAddressingMode(opCode.addressingMode)
		}
	}
}

func (cpu CPU) loadAndRUn(program []uint8) {
	cpu.load(program)
	cpu.reset()
	cpu.run()
}
