package cpu

import "fmt"

// Addressing modes

type AddressingMode int

const (
	Implied AddressingMode = iota
	Accumulator
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
	0x0A: {operation: ASL, addressingMode: Accumulator, bytes: 1, cycles: 2},
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
	0x00: {operation: BRK, addressingMode: Implied, bytes: 1, cycles: 7},
	// BVC
	0x50: {operation: BVC, addressingMode: Relative, bytes: 2, cycles: 2},
	// BVS
	0x70: {operation: BVS, addressingMode: Relative, bytes: 2, cycles: 2},
	// CLC
	0x18: {operation: CLC, addressingMode: Implied, bytes: 1, cycles: 2},
	// CLD
	0xD8: {operation: CLD, addressingMode: Implied, bytes: 1, cycles: 2},
	// CLI
	0x58: {operation: CLI, addressingMode: Implied, bytes: 1, cycles: 2},
	// CLV
	0xB8: {operation: CLD, addressingMode: Implied, bytes: 1, cycles: 2},
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
	0xCA: {operation: DEX, addressingMode: Implied, bytes: 1, cycles: 2},
	// DEY
	0x88: {operation: DEY, addressingMode: Implied, bytes: 1, cycles: 2},
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
	0xE8: {operation: INX, addressingMode: Implied, bytes: 1, cycles: 2},
	// INY
	0xC8: {operation: INY, addressingMode: Implied, bytes: 1, cycles: 2},
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
	0xA2: {operation: LDX, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xA6: {operation: LDX, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xB6: {operation: LDX, addressingMode: ZeroPageY, bytes: 2, cycles: 4},
	0xAE: {operation: LDX, addressingMode: Absolute, bytes: 3, cycles: 4},
	0xBE: {operation: LDX, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	// LDY
	0xA0: {operation: LDY, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xA4: {operation: LDY, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xB4: {operation: LDY, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0xAC: {operation: LDY, addressingMode: Absolute, bytes: 3, cycles: 4},
	0xBC: {operation: LDY, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	// LSR
	0x4A: {operation: LSR, addressingMode: Accumulator, bytes: 1, cycles: 2},
	0x46: {operation: LSR, addressingMode: ZeroPage, bytes: 2, cycles: 5},
	0x56: {operation: LSR, addressingMode: ZeroPageX, bytes: 2, cycles: 6},
	0x4E: {operation: LSR, addressingMode: Absolute, bytes: 3, cycles: 6},
	0x5E: {operation: LSR, addressingMode: AbsoluteX, bytes: 3, cycles: 7},
	// NOP
	0xEA: {operation: NOP, addressingMode: Implied, bytes: 1, cycles: 2},
	// ORA
	0x09: {operation: ORA, addressingMode: Immediate, bytes: 2, cycles: 2},
	0x05: {operation: ORA, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x15: {operation: ORA, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0x0D: {operation: ORA, addressingMode: Absolute, bytes: 3, cycles: 4},
	0x1D: {operation: ORA, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	0x19: {operation: ORA, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	0x01: {operation: ORA, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0x11: {operation: ORA, addressingMode: IndirectY, bytes: 2, cycles: 5},
	// PHA
	0x48: {operation: PHA, addressingMode: Implied, bytes: 1, cycles: 3},
	// PHP
	0x08: {operation: PHP, addressingMode: Implied, bytes: 1, cycles: 3},
	// PLA
	0x68: {operation: PLA, addressingMode: Implied, bytes: 1, cycles: 4},
	// PLP
	0x28: {operation: PLP, addressingMode: Implied, bytes: 1, cycles: 4},
	// ROL
	0x2A: {operation: ROL, addressingMode: Accumulator, bytes: 1, cycles: 2},
	0x26: {operation: ROL, addressingMode: ZeroPage, bytes: 2, cycles: 5},
	0x36: {operation: ROL, addressingMode: ZeroPageX, bytes: 2, cycles: 6},
	0x2E: {operation: ROL, addressingMode: Absolute, bytes: 3, cycles: 6},
	0x3E: {operation: ROL, addressingMode: AbsoluteX, bytes: 3, cycles: 7},
	// ROR
	0x6A: {operation: ROR, addressingMode: Accumulator, bytes: 1, cycles: 2},
	0x66: {operation: ROR, addressingMode: ZeroPage, bytes: 2, cycles: 5},
	0x76: {operation: ROR, addressingMode: ZeroPageX, bytes: 2, cycles: 6},
	0x6E: {operation: ROR, addressingMode: Absolute, bytes: 3, cycles: 6},
	0x7E: {operation: ROR, addressingMode: AbsoluteX, bytes: 3, cycles: 7},
	// RTI
	0x40: {operation: RTI, addressingMode: Implied, bytes: 1, cycles: 6},
	// RTS
	0x60: {operation: RTS, addressingMode: Implied, bytes: 1, cycles: 6},
	// SBC
	0xE9: {operation: SBC, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xE5: {operation: SBC, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xF5: {operation: SBC, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0xED: {operation: SBC, addressingMode: Absolute, bytes: 3, cycles: 4},
	0xFD: {operation: SBC, addressingMode: AbsoluteX, bytes: 3, cycles: 4},
	0xF9: {operation: SBC, addressingMode: AbsoluteY, bytes: 3, cycles: 4},
	0xE1: {operation: SBC, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0xF1: {operation: SBC, addressingMode: IndirectY, bytes: 2, cycles: 5},
	// SEC
	0x38: {operation: SEC, addressingMode: Implied, bytes: 1, cycles: 2},
	// SED
	0xF8: {operation: SED, addressingMode: Implied, bytes: 1, cycles: 2},
	// SEI
	0x78: {operation: SEI, addressingMode: Implied, bytes: 1, cycles: 2},
	// STA
	0x85: {operation: STA, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x95: {operation: STA, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0x8D: {operation: STA, addressingMode: Absolute, bytes: 3, cycles: 4},
	0x9D: {operation: STA, addressingMode: AbsoluteX, bytes: 3, cycles: 5},
	0x99: {operation: STA, addressingMode: AbsoluteY, bytes: 3, cycles: 5},
	0x81: {operation: STA, addressingMode: IndirectX, bytes: 2, cycles: 6},
	0x91: {operation: STA, addressingMode: IndirectY, bytes: 2, cycles: 6},
	// STX
	0x86: {operation: STX, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x96: {operation: STX, addressingMode: ZeroPageY, bytes: 2, cycles: 4},
	0x8E: {operation: STX, addressingMode: Absolute, bytes: 3, cycles: 4},
	// STY
	0x84: {operation: STY, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x94: {operation: STY, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0x8C: {operation: STY, addressingMode: Absolute, bytes: 3, cycles: 4},
	// TAX
	0xAA: {operation: TAX, addressingMode: Implied, bytes: 1, cycles: 2},
	// TAY
	0xA8: {operation: TAY, addressingMode: Implied, bytes: 1, cycles: 2},
	// TSX
	0xBA: {operation: TSX, addressingMode: Implied, bytes: 1, cycles: 2},
	// TXA
	0x8A: {operation: TXA, addressingMode: Implied, bytes: 1, cycles: 2},
	// TXS
	0x9A: {operation: TXS, addressingMode: Implied, bytes: 1, cycles: 2},
	// TYA
	0x98: {operation: TYA, addressingMode: Implied, bytes: 1, cycles: 2},
}

func matchHexCodeWithOpsCode(hexCode uint8) OpCode {
	var opsCode, ok = hexToOpsCode[hexCode]
	if !ok {
		panic(fmt.Sprintf("hex code %v is unsupported", hexCode))
	}
	return opsCode
}
