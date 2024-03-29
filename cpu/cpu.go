package cpu

import (
	"encoding/binary"
	"fmt"
	"nes-emulator/bus"
	"strings"
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
	bus            *bus.Bus
}

// Generic helpers

func isNegative(value uint8) bool {
	return value&0b1000_0000 != 0
}

// Flags

type StatusFlag uint8

const (
	CARRY_FLAG             StatusFlag = 0b0000_0001
	ZERO_FLAG              StatusFlag = 0b0000_0010
	INTERRUPT_DISABLE_FLAG StatusFlag = 0b0000_0100
	DECIMAL_FLAG           StatusFlag = 0b0000_1000
	BREAK_FLAG             StatusFlag = 0b0001_0000
	BREAK_2_FLAG           StatusFlag = 0b0010_0000
	OVERFLOW_FLAG          StatusFlag = 0b0100_0000
	NEGATIVE_FLAG          StatusFlag = 0b1000_0000
)

func (cpu *CPU) setFlagToValue(statusFlag StatusFlag, value bool) {
	if value {
		cpu.statusFlags = cpu.statusFlags | uint8(statusFlag)
	} else {
		cpu.statusFlags = cpu.statusFlags & (^uint8(statusFlag))
	}
}

func (cpu *CPU) isFlagSet(statusFlag StatusFlag) bool {
	return cpu.statusFlags&uint8(statusFlag) != 0
}

func (cpu *CPU) setZeroFlagAndNegativeFlagForResult(result uint8) {
	cpu.setFlagToValue(ZERO_FLAG, result == 0)
	cpu.setFlagToValue(NEGATIVE_FLAG, isNegative(result))
}

// Memory helpers

func (cpu *CPU) memoryRead(address uint16) uint8 {
	return cpu.bus.MemoryRead(address)
}

func (cpu *CPU) memoryWrite(address uint16, data uint8) {
	cpu.bus.MemoryWrite(address, data)
}

func (cpu *CPU) memoryReadU16(address uint16) uint16 {
	return cpu.bus.MemoryReadU16(address)
}

func (cpu *CPU) memoryWriteU16(address uint16, data uint16) {
	cpu.bus.MemoryWriteU16(address, data)
}

// Stack helpers

func (cpu *CPU) pushStack(value uint8) {
	cpu.memoryWrite(STACK_BASE+uint16(cpu.stackPointer), value)
	cpu.stackPointer -= 1
}

func (cpu *CPU) pullStack() uint8 {
	cpu.stackPointer += 1
	return cpu.memoryRead(STACK_BASE + uint16(cpu.stackPointer))
}

func (cpu *CPU) pushStackU16(value uint16) {
	var bytes = make([]uint8, 2)
	binary.LittleEndian.PutUint16(bytes, value)
	cpu.pushStack(bytes[1])
	cpu.pushStack(bytes[0])
}

func (cpu *CPU) pullStackU16() uint16 {
	var bytes = make([]uint8, 2)
	bytes[0] = cpu.pullStack()
	bytes[1] = cpu.pullStack()
	return binary.LittleEndian.Uint16(bytes)
}

// This does not get the operand but the address of the operand, which will be the retrieved using memory read
func (cpu *CPU) getOperandAddress(mode AddressingMode, opCodeProgramCounter uint16) uint16 {
	// Program counter is where the opCode is located
	switch mode {
	case Implied:
		return 0
	case Accumulator:
		return 0
	case Immediate:
		return opCodeProgramCounter + 1
	case Relative:
		var offset = cpu.memoryRead(opCodeProgramCounter + 1)
		if !isNegative(offset) {
			return opCodeProgramCounter + uint16(offset) + 2
		} else {
			return opCodeProgramCounter - (0x100 - uint16(offset)) + 2
		}
	case ZeroPage:
		// It's only a 8 bits address with Zero Page, so you can only get an address in the first 256 memory cells
		// But it's faster !
		return uint16(cpu.memoryRead(opCodeProgramCounter + 1))
	case ZeroPageX:
		var pos = cpu.memoryRead(opCodeProgramCounter + 1)
		return uint16(pos + cpu.registerX)
	case ZeroPageY:
		var pos = cpu.memoryRead(opCodeProgramCounter + 1)
		return uint16(pos + cpu.registerY)
	case Absolute:
		return cpu.memoryReadU16(opCodeProgramCounter + 1)
	case AbsoluteX:
		var pos = cpu.memoryReadU16(opCodeProgramCounter + 1)
		return pos + uint16(cpu.registerX)
	case AbsoluteY:
		var pos = cpu.memoryReadU16(opCodeProgramCounter + 1)
		return pos + uint16(cpu.registerY)
	case Indirect:
		var ref = cpu.memoryReadU16(opCodeProgramCounter + 1)
		// Bug with page boundary:
		// If we try to read the end of a page X and the beginning of a page X + 1
		// Instead JMP will read the end of the page X and the beginning of the page X
		if ref&0x00FF == 0x00FF {
			var pageBeginning = ref & 0xFF00
			return binary.LittleEndian.Uint16([]uint8{cpu.memoryRead(ref), cpu.memoryRead(pageBeginning)})
		} else {
			return cpu.memoryReadU16(ref)
		}
	case IndirectX:
		var base = cpu.memoryRead(opCodeProgramCounter + 1)
		// Cannot use cpu.memoryRead16 as we need to wrap the address !
		return binary.LittleEndian.Uint16([]uint8{cpu.memoryRead(uint16(base + cpu.registerX)), cpu.memoryRead(uint16(base + cpu.registerX + 1))})
	case IndirectY:
		var base = cpu.memoryRead(opCodeProgramCounter + 1)
		// Cannot use cpu.memoryRead16 as we need to wrap the address !
		return binary.LittleEndian.Uint16([]uint8{cpu.memoryRead(uint16(base)), cpu.memoryRead(uint16(base + 1))}) + uint16(cpu.registerY)
	default:
		panic(fmt.Sprintf("addressing mode %v is not supported", mode))
	}
}

// Helpers for Ops Code operations

// http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
func (cpu *CPU) addWithCarry(a uint8, b uint8, carry bool) (uint8, bool, bool) {
	var sum = uint16(a) + uint16(b)
	if carry {
		sum += 1
	}
	var hasCarry = sum>>8 != 0
	var result = uint8(sum)
	var hasOverflow = (a^result)&(b^result)&0x80 != 0
	return result, hasCarry, hasOverflow
}

func (cpu *CPU) branch(cpuStepInfos *StepInfos, condition bool) {
	if condition {
		cpu.programCounter = cpuStepInfos.operandAddress
	}
}

// Ops code operations

func (cpu *CPU) adc(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	result, hasCarry, hasOverflow := cpu.addWithCarry(cpu.registerA, operand, cpu.isFlagSet(CARRY_FLAG))
	cpu.registerA = result
	cpu.setFlagToValue(CARRY_FLAG, hasCarry)
	cpu.setFlagToValue(OVERFLOW_FLAG, hasOverflow)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) and(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerA = cpu.registerA & operand
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) asl(cpuStepInfos *StepInfos) {
	if cpuStepInfos.opCode.addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b1000_0000 != 0)
		cpu.registerA = cpu.registerA << 1
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b1000_0000 != 0)
		var result = operand << 1
		cpu.memoryWrite(cpuStepInfos.operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu *CPU) bcc(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, !cpu.isFlagSet(CARRY_FLAG))
}

func (cpu *CPU) bcs(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, cpu.isFlagSet(CARRY_FLAG))
}

func (cpu *CPU) beq(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, cpu.isFlagSet(ZERO_FLAG))
}

func (cpu *CPU) bit(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var result = operand & cpu.registerA
	cpu.setFlagToValue(ZERO_FLAG, result == 0)
	cpu.setFlagToValue(NEGATIVE_FLAG, isNegative(operand))
	cpu.setFlagToValue(OVERFLOW_FLAG, operand&0b0100_0000 != 0)
}

func (cpu *CPU) bmi(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, cpu.isFlagSet(NEGATIVE_FLAG))
}

func (cpu *CPU) bne(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, !cpu.isFlagSet(ZERO_FLAG))
}

func (cpu *CPU) bpl(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, !cpu.isFlagSet(NEGATIVE_FLAG))
}

func (cpu *CPU) bvc(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, !cpu.isFlagSet(OVERFLOW_FLAG))
}

func (cpu *CPU) bvs(cpuStepInfos *StepInfos) {
	cpu.branch(cpuStepInfos, cpu.isFlagSet(OVERFLOW_FLAG))
}

func (cpu *CPU) clc(cpuStepInfos *StepInfos) {
	cpu.setFlagToValue(CARRY_FLAG, false)
}

func (cpu *CPU) cld(cpuStepInfos *StepInfos) {
	cpu.setFlagToValue(DECIMAL_FLAG, false)
}

func (cpu *CPU) cli(cpuStepInfos *StepInfos) {
	cpu.setFlagToValue(INTERRUPT_DISABLE_FLAG, false)
}

func (cpu *CPU) clv(cpuStepInfos *StepInfos) {
	cpu.setFlagToValue(OVERFLOW_FLAG, false)
}

func (cpu *CPU) compare(cpuStepInfos *StepInfos, compareWith uint8) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var result = compareWith - operand
	cpu.setZeroFlagAndNegativeFlagForResult(result)
	cpu.setFlagToValue(CARRY_FLAG, compareWith >= operand)
}

func (cpu *CPU) cmp(cpuStepInfos *StepInfos) {
	cpu.compare(cpuStepInfos, cpu.registerA)
}

func (cpu *CPU) cpx(cpuStepInfos *StepInfos) {
	cpu.compare(cpuStepInfos, cpu.registerX)
}

func (cpu *CPU) cpy(cpuStepInfos *StepInfos) {
	cpu.compare(cpuStepInfos, cpu.registerY)
}

func (cpu *CPU) dec(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var result = operand - 1
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
	cpu.setZeroFlagAndNegativeFlagForResult(result)
}

func (cpu *CPU) dex(cpuStepInfos *StepInfos) {
	cpu.registerX -= 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu *CPU) dey(cpuStepInfos *StepInfos) {
	cpu.registerY -= 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu *CPU) eor(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerA = cpu.registerA ^ operand
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) inc(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var result = operand + 1
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
	cpu.setZeroFlagAndNegativeFlagForResult(result)
}

func (cpu *CPU) inx(cpuStepInfos *StepInfos) {
	cpu.registerX += 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu *CPU) iny(cpuStepInfos *StepInfos) {
	cpu.registerY += 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu *CPU) jmp(cpuStepInfos *StepInfos) {
	// TODO : some shady shit is done here in the tutorial, wtf ??
	cpu.programCounter = cpuStepInfos.operandAddress
}

func (cpu *CPU) jsr(cpuStepInfos *StepInfos) {
	cpu.pushStackU16(cpu.programCounter + getNumberOfBytesReadForOperation(cpuStepInfos.opCode.addressingMode) - 1)
	cpu.programCounter = cpuStepInfos.operandAddress
}

func (cpu *CPU) lda(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerA = operand
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) ldx(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerX = operand
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu *CPU) ldy(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerY = operand
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu *CPU) lsr(cpuStepInfos *StepInfos) {
	if cpuStepInfos.opCode.addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b0000_0001 != 0)
		cpu.registerA = cpu.registerA >> 1
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b0000_0001 != 0)
		var result = operand >> 1
		cpu.memoryWrite(cpuStepInfos.operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu *CPU) nop(cpuStepInfos *StepInfos) {}

func (cpu *CPU) ora(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerA = cpu.registerA | operand
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) pha(cpuStepInfos *StepInfos) {
	cpu.pushStack(cpu.registerA)
}

func (cpu *CPU) php(cpuStepInfos *StepInfos) {
	// https://www.nesdev.org/wiki/Status_flags#The_B_flag
	cpu.pushStack(cpu.statusFlags | uint8(BREAK_FLAG) | uint8(BREAK_2_FLAG))
}

func (cpu *CPU) pla(cpuStepInfos *StepInfos) {
	cpu.registerA = cpu.pullStack()
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) plp(cpuStepInfos *StepInfos) {
	cpu.statusFlags = cpu.pullStack()
	cpu.setFlagToValue(BREAK_FLAG, false)
	cpu.setFlagToValue(BREAK_2_FLAG, true)
}

func (cpu *CPU) rol(cpuStepInfos *StepInfos) {
	var carryMask uint8 = 0b0000_0000
	if cpu.isFlagSet(CARRY_FLAG) {
		carryMask = 0b0000_0001
	}
	if cpuStepInfos.opCode.addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b1000_0000 != 0)
		cpu.registerA = (cpu.registerA << 1) | carryMask
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b1000_0000 != 0)
		var result = operand<<1 | carryMask
		cpu.memoryWrite(cpuStepInfos.operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu *CPU) ror(cpuStepInfos *StepInfos) {
	var carryMask uint8 = 0b0000_0000
	if cpu.isFlagSet(CARRY_FLAG) {
		carryMask = 0b1000_0000
	}
	if cpuStepInfos.opCode.addressingMode == Accumulator {
		cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b0000_0001 != 0)
		cpu.registerA = (cpu.registerA >> 1) | carryMask
		cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	} else {
		var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
		cpu.setFlagToValue(CARRY_FLAG, operand&0b0000_0001 != 0)
		var result = operand>>1 | carryMask
		cpu.memoryWrite(cpuStepInfos.operandAddress, result)
		cpu.setZeroFlagAndNegativeFlagForResult(result)
	}
}

func (cpu *CPU) rti(cpuStepInfos *StepInfos) {
	cpu.statusFlags = cpu.pullStack()
	cpu.setFlagToValue(BREAK_FLAG, false)
	cpu.setFlagToValue(BREAK_2_FLAG, true)
	cpu.programCounter = cpu.pullStackU16()
}

func (cpu *CPU) rts(cpuStepInfos *StepInfos) {
	cpu.programCounter = cpu.pullStackU16() + 1
}

func (cpu *CPU) sbc(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	// Result calculated is A-M-(1-C) = A + (256 - M) - 1 + C = A + (255 - M) + C
	result, hasCarry, hasOverflow := cpu.addWithCarry(cpu.registerA, 255-operand, cpu.isFlagSet(CARRY_FLAG))
	cpu.registerA = result
	cpu.setFlagToValue(CARRY_FLAG, hasCarry)
	cpu.setFlagToValue(OVERFLOW_FLAG, hasOverflow)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) sec(cpuStepInfos *StepInfos) {
	cpu.setFlagToValue(CARRY_FLAG, true)
}

func (cpu *CPU) sed(cpuStepInfos *StepInfos) {
	cpu.setFlagToValue(DECIMAL_FLAG, true)
}

func (cpu *CPU) sei(cpuStepInfos *StepInfos) {
	cpu.setFlagToValue(INTERRUPT_DISABLE_FLAG, true)
}

func (cpu *CPU) sta(cpuStepInfos *StepInfos) {
	cpu.memoryWrite(cpuStepInfos.operandAddress, cpu.registerA)
}

func (cpu *CPU) stx(cpuStepInfos *StepInfos) {
	cpu.memoryWrite(cpuStepInfos.operandAddress, cpu.registerX)
}

func (cpu *CPU) sty(cpuStepInfos *StepInfos) {
	cpu.memoryWrite(cpuStepInfos.operandAddress, cpu.registerY)
}

func (cpu *CPU) tax(cpuStepInfos *StepInfos) {
	cpu.registerX = cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu *CPU) tay(cpuStepInfos *StepInfos) {
	cpu.registerY = cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerY)
}

func (cpu *CPU) tsx(cpuStepInfos *StepInfos) {
	cpu.registerX = cpu.stackPointer
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu *CPU) txa(cpuStepInfos *StepInfos) {
	cpu.registerA = cpu.registerX
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) txs(cpuStepInfos *StepInfos) {
	cpu.stackPointer = cpu.registerX
}

func (cpu *CPU) tya(cpuStepInfos *StepInfos) {
	cpu.registerA = cpu.registerY
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

/***********************/
/* UNDOCUMENTED OPCODES
// TODO: some of them could probably be simplified (like those combining operations could reuse basic operations)
*/
/***********************/

func (cpu *CPU) aac(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var result = operand & cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(result)
	cpu.setFlagToValue(CARRY_FLAG, isNegative(result))
}

func (cpu *CPU) aax(cpuStepInfos *StepInfos) {
	cpu.memoryWrite(cpuStepInfos.operandAddress, cpu.registerA&cpu.registerX)
}

func (cpu *CPU) arr(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var carryMask uint8 = 0b0000_0000
	if cpu.isFlagSet(CARRY_FLAG) {
		carryMask = 0b1000_0000
	}
	cpu.registerA = (operand & cpu.registerA >> 1) | carryMask
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
	cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b0000_0001 != 0)
	var isBit5Set = cpu.registerA&0b0001_0000 == 1
	var isBit6Set = cpu.registerA&0b0010_0000 == 1
	cpu.setFlagToValue(CARRY_FLAG, isBit6Set)
	cpu.setFlagToValue(OVERFLOW_FLAG, isBit6Set != isBit5Set) // XOR
}

func (cpu *CPU) asr(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerA = operand & cpu.registerA
	cpu.setFlagToValue(CARRY_FLAG, cpu.registerA&0b0000_0001 != 0)
	cpu.registerA = cpu.registerA >> 1
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) atx(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerA = operand & cpu.registerA
	cpu.registerX = cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu *CPU) axa(cpuStepInfos *StepInfos) {
	var result = cpu.registerA & cpu.registerX & uint8(cpuStepInfos.operandAddress>>8)
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
}

func (cpu *CPU) axs(cpuStepInfos *StepInfos) {
	cpu.memoryWrite(cpuStepInfos.operandAddress, cpu.registerA)
	cpu.memoryWrite(cpuStepInfos.operandAddress, cpu.registerX)
}

func (cpu *CPU) dcp(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.memoryWrite(cpuStepInfos.operandAddress, operand-1)
	cpu.compare(cpuStepInfos, cpu.registerA)
}

func (cpu *CPU) dop(cpuStepInfos *StepInfos) {}

func (cpu *CPU) isc(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var result = operand + 1
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
	// Result calculated is A-M-(1-C) = A + (256 - M) - 1 + C = A + (255 - M) + C
	var resultAdd, hasCarry, hasOverflow = cpu.addWithCarry(cpu.registerA, 255-result, cpu.isFlagSet(CARRY_FLAG))
	cpu.registerA = resultAdd
	cpu.setFlagToValue(CARRY_FLAG, hasCarry)
	cpu.setFlagToValue(OVERFLOW_FLAG, hasOverflow)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) kil(cpuStepInfos *StepInfos) {
	panic(fmt.Sprintf("kil opcode executed (hex code : %X)", cpuStepInfos.opHexCode))
}

func (cpu *CPU) lar(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	var result = operand & cpu.stackPointer
	cpu.registerA = result
	cpu.registerX = result
	cpu.stackPointer = result
	cpu.setZeroFlagAndNegativeFlagForResult(result)
}

func (cpu *CPU) lax(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.registerA = operand
	cpu.registerX = cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerX)
}

func (cpu *CPU) rla(cpuStepInfos *StepInfos) {
	var carryMask uint8 = 0b0000_0000
	if cpu.isFlagSet(CARRY_FLAG) {
		carryMask = 0b0000_0001
	}
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.setFlagToValue(CARRY_FLAG, operand&0b1000_0000 != 0)
	var result = operand<<1 | carryMask
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
	cpu.registerA = result & cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) rra(cpuStepInfos *StepInfos) {
	var carryMask uint8 = 0b0000_0000
	if cpu.isFlagSet(CARRY_FLAG) {
		carryMask = 0b1000_0000
	}
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.setFlagToValue(CARRY_FLAG, operand&0b0000_0001 != 0)
	var result = operand>>1 | carryMask
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
	cpu.setZeroFlagAndNegativeFlagForResult(result)
	resultAdd, hasCarry, hasOverflow := cpu.addWithCarry(cpu.registerA, result, cpu.isFlagSet(CARRY_FLAG))
	cpu.registerA = resultAdd
	cpu.setFlagToValue(CARRY_FLAG, hasCarry)
	cpu.setFlagToValue(OVERFLOW_FLAG, hasOverflow)
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) slo(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.setFlagToValue(CARRY_FLAG, operand&0b1000_0000 != 0)
	var result = operand << 1
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
	cpu.registerA = result | cpu.registerA
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) sre(cpuStepInfos *StepInfos) {
	var operand = cpu.memoryRead(cpuStepInfos.operandAddress)
	cpu.setFlagToValue(CARRY_FLAG, operand&0b0000_0001 != 0)
	var result = operand >> 1
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
	cpu.registerA = cpu.registerA ^ result
	cpu.setZeroFlagAndNegativeFlagForResult(cpu.registerA)
}

func (cpu *CPU) sxa(cpuStepInfos *StepInfos) {
	var result = (uint8(cpuStepInfos.operandAddress>>8) + 1) & cpu.registerX
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
}

func (cpu *CPU) sya(cpuStepInfos *StepInfos) {
	var result = (uint8(cpuStepInfos.operandAddress>>8) + 1) & cpu.registerY
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
}

func (cpu *CPU) top(cpuStepInfos *StepInfos) {}

func (cpu *CPU) xaa(cpuStepInfos *StepInfos) {
	// Unreliable Opcode : https://www.nesdev.org/wiki/Visual6502wiki/6502_Opcode_8B_(XAA,_ANE)
	panic(fmt.Sprintf("xaa opcode executed (hex code : %X)", cpuStepInfos.opHexCode))
}

func (cpu *CPU) xas(cpuStepInfos *StepInfos) {
	cpu.stackPointer = cpu.registerA & cpu.registerX
	var result = (uint8(cpuStepInfos.operandAddress>>8) + 1) & cpu.stackPointer
	cpu.memoryWrite(cpuStepInfos.operandAddress, result)
}

/***********************/
/* UNDOCUMENTED OPCODES
/***********************/

// Load program and reset CPU

func NewCPU(consoleBus *bus.Bus) CPU {
	var cpu = CPU{
		registerA:      0,
		registerX:      0,
		registerY:      0,
		statusFlags:    0b00100100,
		stackPointer:   STACK_RESET,
		programCounter: 0,
		bus:            consoleBus,
	}
	return cpu
}

func (cpu *CPU) Reset() {
	cpu.registerA = 0
	cpu.registerX = 0
	cpu.registerY = 0
	cpu.statusFlags = 0b00100100
	cpu.stackPointer = STACK_RESET
	cpu.programCounter = 0xC000 //cpu.memoryReadU16(0xFFFC) uncomment when PPU is implemented
}

type StepInfos struct {
	opHexCode      uint8
	opCode         OpCode
	operandAddress uint16
}

func (cpu *CPU) Run() {
	for {
		var opHexCode = cpu.memoryRead(cpu.programCounter)
		var programCounterBeforeOperation = cpu.programCounter
		var opCode = matchOpHexCodeWithOpCode(opHexCode)
		var operandAddress = cpu.getOperandAddress(opCode.addressingMode, cpu.programCounter)
		var stepInfos = &StepInfos{
			opHexCode:      opHexCode,
			opCode:         opCode,
			operandAddress: operandAddress,
		}
		printCPUState(cpu, stepInfos)
		switch opCode.operation {
		case ADC:
			cpu.adc(stepInfos)
		case AND:
			cpu.and(stepInfos)
		case ASL:
			cpu.asl(stepInfos)
		case BCC:
			cpu.bcc(stepInfos)
		case BCS:
			cpu.bcs(stepInfos)
		case BEQ:
			cpu.beq(stepInfos)
		case BIT:
			cpu.bit(stepInfos)
		case BMI:
			cpu.bmi(stepInfos)
		case BNE:
			cpu.bne(stepInfos)
		case BPL:
			cpu.bpl(stepInfos)
		case BRK:
			return
		case BVS:
			cpu.bvs(stepInfos)
		case BVC:
			cpu.bvc(stepInfos)
		case CLC:
			cpu.clc(stepInfos)
		case CLD:
			cpu.cld(stepInfos)
		case CLI:
			cpu.cli(stepInfos)
		case CLV:
			cpu.clv(stepInfos)
		case CMP:
			cpu.cmp(stepInfos)
		case CPX:
			cpu.cpx(stepInfos)
		case CPY:
			cpu.cpy(stepInfos)
		case DEC:
			cpu.dec(stepInfos)
		case DEX:
			cpu.dex(stepInfos)
		case DEY:
			cpu.dey(stepInfos)
		case EOR:
			cpu.eor(stepInfos)
		case INC:
			cpu.inc(stepInfos)
		case INX:
			cpu.inx(stepInfos)
		case INY:
			cpu.iny(stepInfos)
		case JMP:
			cpu.jmp(stepInfos)
		case JSR:
			cpu.jsr(stepInfos)
		case LDA:
			cpu.lda(stepInfos)
		case LDX:
			cpu.ldx(stepInfos)
		case LDY:
			cpu.ldy(stepInfos)
		case LSR:
			cpu.lsr(stepInfos)
		case NOP:
			cpu.nop(stepInfos)
		case ORA:
			cpu.ora(stepInfos)
		case PHA:
			cpu.pha(stepInfos)
		case PHP:
			cpu.php(stepInfos)
		case PLA:
			cpu.pla(stepInfos)
		case PLP:
			cpu.plp(stepInfos)
		case ROL:
			cpu.rol(stepInfos)
		case ROR:
			cpu.ror(stepInfos)
		case RTI:
			cpu.rti(stepInfos)
		case RTS:
			cpu.rts(stepInfos)
		case SBC:
			cpu.sbc(stepInfos)
		case SEC:
			cpu.sec(stepInfos)
		case SED:
			cpu.sed(stepInfos)
		case SEI:
			cpu.sei(stepInfos)
		case STA:
			cpu.sta(stepInfos)
		case STX:
			cpu.stx(stepInfos)
		case STY:
			cpu.sty(stepInfos)
		case TAX:
			cpu.tax(stepInfos)
		case TAY:
			cpu.tay(stepInfos)
		case TSX:
			cpu.tsx(stepInfos)
		case TXA:
			cpu.txa(stepInfos)
		case TXS:
			cpu.txs(stepInfos)
		case TYA:
			cpu.tya(stepInfos)
		/***********************/
		/* UNDOCUMENTED OPCODES
		/***********************/
		case _AAC:
			cpu.aac(stepInfos)
		case _AAX:
			cpu.aax(stepInfos)
		case _ARR:
			cpu.arr(stepInfos)
		case _ASR:
			cpu.asr(stepInfos)
		case _ATX:
			cpu.atx(stepInfos)
		case _AXA:
			cpu.axa(stepInfos)
		case _AXS:
			cpu.axs(stepInfos)
		case _DCP:
			cpu.dcp(stepInfos)
		case _DOP:
			cpu.dop(stepInfos)
		case _ISC:
			cpu.isc(stepInfos)
		case _KIL:
			cpu.kil(stepInfos)
		case _LAR:
			cpu.lar(stepInfos)
		case _LAX:
			cpu.lax(stepInfos)
		case _NOP:
			cpu.nop(stepInfos)
		case _RLA:
			cpu.rla(stepInfos)
		case _RRA:
			cpu.rra(stepInfos)
		case _SBC:
			cpu.sbc(stepInfos)
		case _SLO:
			cpu.slo(stepInfos)
		case _SRE:
			cpu.sre(stepInfos)
		case _SXA:
			cpu.sxa(stepInfos)
		case _SYA:
			cpu.sya(stepInfos)
		case _TOP:
			cpu.top(stepInfos)
		case _XAA:
			cpu.xaa(stepInfos)
		case _XAS:
			cpu.xas(stepInfos)
		default:
			panic(fmt.Sprintf("operation %v is unsupported", opCode.operation))
		}
		// No jump or branch has occurred
		if programCounterBeforeOperation == cpu.programCounter {
			cpu.programCounter += getNumberOfBytesReadForOperation(opCode.addressingMode)
		}
	}
}

// TODO : change illegal opcode to match those
func convertOperationForPrinting(operation Operation) string {
	switch operation {
	case _DOP:
		return "*NOP"
	case _TOP:
		return "*NOP"
	case _AAX:
		return "*SAX"
	case _ISC:
		return "*ISB"
	default:
		return string(operation)
	}
}

// Must be run at the beginning of the loop
func printCPUState(cpu *CPU, cpuStepInfos *StepInfos) {
	var builder = strings.Builder{}
	var param1 = cpu.memoryRead(cpu.programCounter + 1)
	var param2 = cpu.memoryRead(cpu.programCounter + 2)
	var bytesReadForAddressing = getNumberOfBytesReadForOperation(cpuStepInfos.opCode.addressingMode)

	// Program Counter
	builder.WriteString(fmt.Sprintf("%04X  ", cpu.programCounter))

	// CPU opcode
	var hexOpCodeTrace string
	switch bytesReadForAddressing {
	case 3:
		hexOpCodeTrace = fmt.Sprintf("%02X %02X %02X", cpuStepInfos.opHexCode, cpu.memoryRead(cpu.programCounter+1), cpu.memoryRead(cpu.programCounter+2))
	case 2:
		hexOpCodeTrace = fmt.Sprintf("%02X %02X", cpuStepInfos.opHexCode, cpu.memoryRead(cpu.programCounter+1))
	case 1:
		hexOpCodeTrace = fmt.Sprintf("%02X", cpuStepInfos.opHexCode)
	}

	// Format log properly for unofficial operations
	if strings.HasPrefix(string(cpuStepInfos.opCode.operation), "*") {
		builder.WriteString(fmt.Sprintf("%-9s", hexOpCodeTrace))
	} else {
		builder.WriteString(fmt.Sprintf("%-10s", hexOpCodeTrace))
	}

	// CPU opcode in assembly
	builder.WriteString(fmt.Sprintf("%s ", convertOperationForPrinting(cpuStepInfos.opCode.operation)))

	var addressingTrace string
	switch cpuStepInfos.opCode.addressingMode {
	case Implied:
		addressingTrace = fmt.Sprintf("")
	case Accumulator:
		addressingTrace = fmt.Sprintf("A")
	case Immediate:
		addressingTrace = fmt.Sprintf("#$%02X", param1)
	case Relative:
		// Branching instruction
		addressingTrace = fmt.Sprintf("$%04X", cpuStepInfos.operandAddress)
	case ZeroPage:
		addressingTrace = fmt.Sprintf("$%02X = %02X", param1, cpu.memoryRead(cpuStepInfos.operandAddress))
	case ZeroPageX:
		addressingTrace = fmt.Sprintf("$%02X,X @ %02X = %02X", param1, cpuStepInfos.operandAddress, cpu.memoryRead(cpuStepInfos.operandAddress))
	case ZeroPageY:
		addressingTrace = fmt.Sprintf("$%02X,Y @ %02X = %02X", param1, cpuStepInfos.operandAddress, cpu.memoryRead(cpuStepInfos.operandAddress))
	case Absolute:
		if cpuStepInfos.opCode.operation == JMP || cpuStepInfos.opCode.operation == JSR {
			addressingTrace = fmt.Sprintf("$%02X%02X", param2, param1)
		} else {
			addressingTrace = fmt.Sprintf("$%02X%02X = %02X", param2, param1, cpu.memoryRead(cpuStepInfos.operandAddress))
		}
	case AbsoluteX:
		addressingTrace = fmt.Sprintf("$%02X%02X,X @ %04X = %02X", param2, param1, cpuStepInfos.operandAddress, cpu.memoryRead(cpuStepInfos.operandAddress))
	case AbsoluteY:
		addressingTrace = fmt.Sprintf("$%02X%02X,Y @ %04X = %02X", param2, param1, cpuStepInfos.operandAddress, cpu.memoryRead(cpuStepInfos.operandAddress))
	case Indirect:
		// JMP
		addressingTrace = fmt.Sprintf("($%02X%02X) = %04X", param2, param1, cpuStepInfos.operandAddress)
	case IndirectX:
		addressingTrace = fmt.Sprintf("($%02X,X) @ %02X = %04X = %02X", param1, param1+cpu.registerX, cpuStepInfos.operandAddress, cpu.memoryRead(cpuStepInfos.operandAddress))
	case IndirectY:
		addressingTrace = fmt.Sprintf("($%02X),Y = %04X @ %04X = %02X", param1, cpuStepInfos.operandAddress-uint16(cpu.registerY), cpuStepInfos.operandAddress, cpu.memoryRead(cpuStepInfos.operandAddress))
	default:
		panic(fmt.Sprintf("addressing mode %v is not supported for tracing", cpuStepInfos.opCode.addressingMode))
	}
	builder.WriteString(fmt.Sprintf("%-28s", addressingTrace))

	// CPU Registers
	builder.WriteString(fmt.Sprintf("A:%02X X:%02X Y:%02X P:%02X SP:%02X", cpu.registerA, cpu.registerX, cpu.registerY, cpu.statusFlags, cpu.stackPointer))
	// TODO : CPU and PPU cycles

	fmt.Println(builder.String())
}
