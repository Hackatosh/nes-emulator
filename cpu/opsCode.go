package cpu

import "fmt"

// Addressing modes

type AddressingMode int

const (
	Implicit AddressingMode = iota
	Relative
	Immediate
	ZeroPage
	ZeroPageX
	ZeroPageY
	Absolute
	AbsoluteX
	AbsoluteY
	Indirect
	IndirectX
	IndirectY
)

type Operation int

const (
	ADC Operation = iota
	AND
	ASL
	BCC
	BCS
	BEQ
	BIT
	BMI
	BNE
	BPL
	BRK
	BVC
	BVS
	CLC
	CLD
	CLI
	CLV
	CMP
	CPX
	CPY
	DEC
	DEX
	DEY
	EOR
	INC
	INX
	INY
	JMP
	JSR
	LDA
	LDX
	LDY
	LSR
	NOP
	ORA
	PHA
	PHP
	PLA
	PLP
	ROL
	ROR
	RTI
	RTS
	SBC
	SEC
	SED
	SEI
	STA
	STX
	STY
	TAX
	TAY
	TSX
	TXA
	TXS
	TYA
)

type OpCode struct {
	operation      Operation
	addressingMode AddressingMode
	bytes          uint16
	cycles         int
}

// https://www.nesdev.org/obelisk-6502-guide/reference.html#TYA
var hexToOpsCode = map[uint8]OpCode{
	// ADC
	0x69: {operation: ADC, addressingMode: Immediate, bytes: 2, cycles: 2},
	0x65: {operation: ADC, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x75: {operation: ADC, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0x6D: {operation: ADC, addressingMode: Absolute, bytes: 3, cycles: 4},
	0x7D: {operation: ADC, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	0x79: {operation: ADC, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	0x61: {operation: ADC, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0x71: {operation: ADC, addressingMode: IndirectY, bytes: 2, cycles: 5},
	// AND
	0x29: {operation: AND, addressingMode: Immediate, bytes: 2, cycles: 2},
	0x25: {operation: AND, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x35: {operation: AND, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0x2D: {operation: AND, addressingMode: Absolute, bytes: 3, cycles: 4},
	0x3D: {operation: AND, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	0x39: {operation: AND, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	0x21: {operation: AND, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0x31: {operation: AND, addressingMode: IndirectY, bytes: 2, cycles: 5},
	// ASL
	0x0A: {operation: ASL, addressingMode: Implicit, bytes: 1, cycles: 2},
	0x06: {operation: ASL, addressingMode: ZeroPage, bytes: 2, cycles: 5},
	0x16: {operation: ASL, addressingMode: ZeroPageX, bytes: 2, cycles: 6},
	0x0E: {operation: ASL, addressingMode: Absolute, bytes: 3, cycles: 6},
	0x1E: {operation: ASL, addressingMode: AbsoluteX, bytes: 3, cycles: 7},
	// BCC
	0x90: {operation: BCC, addressingMode: Relative, bytes: 2, cycles: 2},
	// BCS
	0xB0: {operation: BCS, addressingMode: Relative, bytes: 2, cycles: 2},
	// BEQ
	0xF0: {operation: BEQ, addressingMode: Relative, bytes: 2, cycles: 2},
	// BIT
	0x24: {operation: BIT, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x2C: {operation: BIT, addressingMode: Absolute, bytes: 3, cycles: 4},
	// BMI
	0x30: {operation: BMI, addressingMode: Relative, bytes: 2, cycles: 2},
	// BNE
	0xD0: {operation: BNE, addressingMode: Relative, bytes: 2, cycles: 2},
	// BPL
	0x10: {operation: BPL, addressingMode: Relative, bytes: 2, cycles: 2},
	// BRK
	0x00: {operation: BRK, addressingMode: Implicit, bytes: 1, cycles: 7},
	// BVC
	0x50: {operation: BVC, addressingMode: Relative, bytes: 2, cycles: 2},
	// BVS
	0x70: {operation: BVS, addressingMode: Relative, bytes: 2, cycles: 2},
	// CLC
	0x18: {operation: CLC, addressingMode: Implicit, bytes: 1, cycles: 2},
	// CLD
	0xD8: {operation: CLD, addressingMode: Implicit, bytes: 1, cycles: 2},
	// CLI
	0x58: {operation: CLI, addressingMode: Implicit, bytes: 1, cycles: 2},
	// CLV
	0xB8: {operation: CLD, addressingMode: Implicit, bytes: 1, cycles: 2},
	// CMP
	0xC9: {operation: CMP, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xC5: {operation: CMP, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xD5: {operation: CMP, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0xCD: {operation: CMP, addressingMode: Absolute, bytes: 3, cycles: 4},
	0xDD: {operation: CMP, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	0xD9: {operation: CMP, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	0xC1: {operation: CMP, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0xD1: {operation: CMP, addressingMode: IndirectY, bytes: 2, cycles: 5},
	// CPX
	0xE0: {operation: CPX, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xE4: {operation: CPX, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xEC: {operation: CPX, addressingMode: Absolute, bytes: 3, cycles: 4},
	// CPY
	0xC0: {operation: CPY, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xC4: {operation: CPY, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xCC: {operation: CPY, addressingMode: Absolute, bytes: 3, cycles: 4},
	// DEC
	0xC6: {operation: DEC, addressingMode: ZeroPage, bytes: 2, cycles: 5},
	0xD6: {operation: DEC, addressingMode: ZeroPageX, bytes: 2, cycles: 6},
	0xCE: {operation: DEC, addressingMode: Absolute, bytes: 3, cycles: 6},
	0xDE: {operation: DEC, addressingMode: AbsoluteX, bytes: 3, cycles: 7},
	// DEX
	0xCA: {operation: DEX, addressingMode: Implicit, bytes: 1, cycles: 2},
	// DEY
	0x88: {operation: DEY, addressingMode: Implicit, bytes: 1, cycles: 2},
	// EOR
	0x49: {operation: EOR, addressingMode: Immediate, bytes: 2, cycles: 2},
	0x45: {operation: EOR, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x55: {operation: EOR, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0x4D: {operation: EOR, addressingMode: Absolute, bytes: 3, cycles: 4},
	0x5D: {operation: EOR, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	0x59: {operation: EOR, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	0x41: {operation: EOR, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0x51: {operation: EOR, addressingMode: IndirectY, bytes: 2, cycles: 5},
	// INC
	0xE6: {operation: INC, addressingMode: ZeroPage, bytes: 2, cycles: 5},
	0xF6: {operation: INC, addressingMode: ZeroPageX, bytes: 2, cycles: 6},
	0xEE: {operation: INC, addressingMode: Absolute, bytes: 3, cycles: 6},
	0xFE: {operation: INC, addressingMode: AbsoluteX, bytes: 3, cycles: 7},
	// INX
	0xE8: {operation: INX, addressingMode: Implicit, bytes: 1, cycles: 2},
	// INY
	0xC8: {operation: INY, addressingMode: Implicit, bytes: 1, cycles: 2},
	// JMP
	0x4C: {operation: JMP, addressingMode: Absolute, bytes: 3, cycles: 3},
	0x6C: {operation: JMP, addressingMode: Indirect, bytes: 3, cycles: 5},
	// JSR
	0x20: {operation: JSR, addressingMode: Absolute, bytes: 3, cycles: 6},
	// LDA
	0xA9: {operation: LDA, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xA5: {operation: LDA, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xB5: {operation: LDA, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0xAD: {operation: LDA, addressingMode: Absolute, bytes: 3, cycles: 4},
	0xBD: {operation: LDA, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	0xB9: {operation: LDA, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	0xA1: {operation: LDA, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0xB1: {operation: LDA, addressingMode: IndirectY, bytes: 2, cycles: 5},
	// LDX
	// LDY
	// LSR
	// NOP
	// ORA
	// PHA
	// PHP
	// PLA
	// PLP
	// ROL
	// ROR
	// RTI
	// RTS
	// SBC
	// SEC
	// SED
	// SEI
	// STA
	// STX
	// STY
	// TAX
	0xAA: {operation: TAX, addressingMode: Implicit, bytes: 1, cycles: 2},
	// TAY
	0xA8: {operation: TAY, addressingMode: Implicit, bytes: 1, cycles: 2},
	// TSX
	// TXA
	0x8A: {operation: TXA, addressingMode: Implicit, bytes: 1, cycles: 2},
	// TXS
	// TYA
	0x98: {operation: TYA, addressingMode: Implicit, bytes: 1, cycles: 2},
}

func matchHexCodeWithOpsCode(hexCode uint8) OpCode {
	var opsCode, ok = hexToOpsCode[hexCode]
	if !ok {
		panic(fmt.Sprintf("hex code %v is unsupported", hexCode))
	}
	return opsCode
}
