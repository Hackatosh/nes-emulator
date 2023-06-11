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

func getNumberOfBytesReadForOperation(addressingMode AddressingMode) uint16 {
	switch addressingMode {
	case Implied, Accumulator:
		return 1
	case Relative, Immediate, ZeroPage, ZeroPageX, ZeroPageY, IndirectX, IndirectY:
		return 2
	case Indirect, Absolute, AbsoluteX, AbsoluteY:
		return 3
	default:
		panic(fmt.Sprintf("addressing mode %v is unsupported for get number of bytes read"))
	}
}

type Operation string

const (
	ADC Operation = "ADC"
	AND           = "AND"
	ASL           = "ASL"
	BCC           = "BCC"
	BCS           = "BCS"
	BEQ           = "BEQ"
	BIT           = "BIT"
	BMI           = "BMI"
	BNE           = "BNE"
	BPL           = "BPL"
	BRK           = "BRK"
	BVC           = "BVC"
	BVS           = "BVS"
	CLC           = "CLC"
	CLD           = "CLD"
	CLI           = "CLI"
	CLV           = "CLV"
	CMP           = "CMP"
	CPX           = "CPX"
	CPY           = "CPY"
	DEC           = "DEC"
	DEX           = "DEX"
	DEY           = "DEY"
	EOR           = "EOR"
	INC           = "INC"
	INX           = "INX"
	INY           = "INY"
	JMP           = "JMP"
	JSR           = "JSR"
	LDA           = "LDA"
	LDX           = "LDX"
	LDY           = "LDY"
	LSR           = "LSR"
	NOP           = "NOP"
	ORA           = "ORA"
	PHA           = "PHA"
	PHP           = "PHP"
	PLA           = "PLA"
	PLP           = "PLP"
	ROL           = "ROL"
	ROR           = "ROR"
	RTI           = "RTI"
	RTS           = "RTS"
	SBC           = "SBC"
	SEC           = "SEC"
	SED           = "SED"
	SEI           = "SEI"
	STA           = "STA"
	STX           = "STX"
	STY           = "STY"
	TAX           = "TAX"
	TAY           = "TAY"
	TSX           = "TSX"
	TXA           = "TXA"
	TXS           = "TXS"
	TYA           = "TYA"
)

type OpCode struct {
	operation      Operation
	addressingMode AddressingMode
	cycles         int
}

// https://www.nesdev.org/obelisk-6502-guide/reference.html
// TODO : cycles miss pages crossed / branching taken considerations
var hexToOpsCode = map[uint8]OpCode{
	// ADC
	0x69: {operation: ADC, addressingMode: Immediate, cycles: 2},
	0x65: {operation: ADC, addressingMode: ZeroPage, cycles: 3},
	0x75: {operation: ADC, addressingMode: ZeroPageX, cycles: 4},
	0x6D: {operation: ADC, addressingMode: Absolute, cycles: 4},
	0x7D: {operation: ADC, addressingMode: AbsoluteX, cycles: 4},
	0x79: {operation: ADC, addressingMode: AbsoluteY, cycles: 4},
	0x61: {operation: ADC, addressingMode: IndirectX, cycles: 6},
	0x71: {operation: ADC, addressingMode: IndirectY, cycles: 5},
	// AND
	0x29: {operation: AND, addressingMode: Immediate, cycles: 2},
	0x25: {operation: AND, addressingMode: ZeroPage, cycles: 3},
	0x35: {operation: AND, addressingMode: ZeroPageX, cycles: 4},
	0x2D: {operation: AND, addressingMode: Absolute, cycles: 4},
	0x3D: {operation: AND, addressingMode: AbsoluteX, cycles: 4},
	0x39: {operation: AND, addressingMode: AbsoluteY, cycles: 4},
	0x21: {operation: AND, addressingMode: IndirectX, cycles: 6},
	0x31: {operation: AND, addressingMode: IndirectY, cycles: 5},
	// ASL
	0x0A: {operation: ASL, addressingMode: Accumulator, cycles: 2},
	0x06: {operation: ASL, addressingMode: ZeroPage, cycles: 5},
	0x16: {operation: ASL, addressingMode: ZeroPageX, cycles: 6},
	0x0E: {operation: ASL, addressingMode: Absolute, cycles: 6},
	0x1E: {operation: ASL, addressingMode: AbsoluteX, cycles: 7},
	// BCC
	0x90: {operation: BCC, addressingMode: Relative, cycles: 2},
	// BCS
	0xB0: {operation: BCS, addressingMode: Relative, cycles: 2},
	// BEQ
	0xF0: {operation: BEQ, addressingMode: Relative, cycles: 2},
	// BIT
	0x24: {operation: BIT, addressingMode: ZeroPage, cycles: 3},
	0x2C: {operation: BIT, addressingMode: Absolute, cycles: 4},
	// BMI
	0x30: {operation: BMI, addressingMode: Relative, cycles: 2},
	// BNE
	0xD0: {operation: BNE, addressingMode: Relative, cycles: 2},
	// BPL
	0x10: {operation: BPL, addressingMode: Relative, cycles: 2},
	// BRK
	0x00: {operation: BRK, addressingMode: Implied, cycles: 7},
	// BVC
	0x50: {operation: BVC, addressingMode: Relative, cycles: 2},
	// BVS
	0x70: {operation: BVS, addressingMode: Relative, cycles: 2},
	// CLC
	0x18: {operation: CLC, addressingMode: Implied, cycles: 2},
	// CLD
	0xD8: {operation: CLD, addressingMode: Implied, cycles: 2},
	// CLI
	0x58: {operation: CLI, addressingMode: Implied, cycles: 2},
	// CLV
	0xB8: {operation: CLV, addressingMode: Implied, cycles: 2},
	// CMP
	0xC9: {operation: CMP, addressingMode: Immediate, cycles: 2},
	0xC5: {operation: CMP, addressingMode: ZeroPage, cycles: 3},
	0xD5: {operation: CMP, addressingMode: ZeroPageX, cycles: 4},
	0xCD: {operation: CMP, addressingMode: Absolute, cycles: 4},
	0xDD: {operation: CMP, addressingMode: AbsoluteX, cycles: 4},
	0xD9: {operation: CMP, addressingMode: AbsoluteY, cycles: 4},
	0xC1: {operation: CMP, addressingMode: IndirectX, cycles: 6},
	0xD1: {operation: CMP, addressingMode: IndirectY, cycles: 5},
	// CPX
	0xE0: {operation: CPX, addressingMode: Immediate, cycles: 2},
	0xE4: {operation: CPX, addressingMode: ZeroPage, cycles: 3},
	0xEC: {operation: CPX, addressingMode: Absolute, cycles: 4},
	// CPY
	0xC0: {operation: CPY, addressingMode: Immediate, cycles: 2},
	0xC4: {operation: CPY, addressingMode: ZeroPage, cycles: 3},
	0xCC: {operation: CPY, addressingMode: Absolute, cycles: 4},
	// DEC
	0xC6: {operation: DEC, addressingMode: ZeroPage, cycles: 5},
	0xD6: {operation: DEC, addressingMode: ZeroPageX, cycles: 6},
	0xCE: {operation: DEC, addressingMode: Absolute, cycles: 6},
	0xDE: {operation: DEC, addressingMode: AbsoluteX, cycles: 7},
	// DEX
	0xCA: {operation: DEX, addressingMode: Implied, cycles: 2},
	// DEY
	0x88: {operation: DEY, addressingMode: Implied, cycles: 2},
	// EOR
	0x49: {operation: EOR, addressingMode: Immediate, cycles: 2},
	0x45: {operation: EOR, addressingMode: ZeroPage, cycles: 3},
	0x55: {operation: EOR, addressingMode: ZeroPageX, cycles: 4},
	0x4D: {operation: EOR, addressingMode: Absolute, cycles: 4},
	0x5D: {operation: EOR, addressingMode: AbsoluteX, cycles: 4},
	0x59: {operation: EOR, addressingMode: AbsoluteY, cycles: 4},
	0x41: {operation: EOR, addressingMode: IndirectX, cycles: 6},
	0x51: {operation: EOR, addressingMode: IndirectY, cycles: 5},
	// INC
	0xE6: {operation: INC, addressingMode: ZeroPage, cycles: 5},
	0xF6: {operation: INC, addressingMode: ZeroPageX, cycles: 6},
	0xEE: {operation: INC, addressingMode: Absolute, cycles: 6},
	0xFE: {operation: INC, addressingMode: AbsoluteX, cycles: 7},
	// INX
	0xE8: {operation: INX, addressingMode: Implied, cycles: 2},
	// INY
	0xC8: {operation: INY, addressingMode: Implied, cycles: 2},
	// JMP
	0x4C: {operation: JMP, addressingMode: Absolute, cycles: 3},
	0x6C: {operation: JMP, addressingMode: Indirect, cycles: 5},
	// JSR
	0x20: {operation: JSR, addressingMode: Absolute, cycles: 6},
	// LDA
	0xA9: {operation: LDA, addressingMode: Immediate, cycles: 2},
	0xA5: {operation: LDA, addressingMode: ZeroPage, cycles: 3},
	0xB5: {operation: LDA, addressingMode: ZeroPageX, cycles: 4},
	0xAD: {operation: LDA, addressingMode: Absolute, cycles: 4},
	0xBD: {operation: LDA, addressingMode: AbsoluteX, cycles: 4},
	0xB9: {operation: LDA, addressingMode: AbsoluteY, cycles: 4},
	0xA1: {operation: LDA, addressingMode: IndirectX, cycles: 6},
	0xB1: {operation: LDA, addressingMode: IndirectY, cycles: 5},
	// LDX
	0xA2: {operation: LDX, addressingMode: Immediate, cycles: 2},
	0xA6: {operation: LDX, addressingMode: ZeroPage, cycles: 3},
	0xB6: {operation: LDX, addressingMode: ZeroPageY, cycles: 4},
	0xAE: {operation: LDX, addressingMode: Absolute, cycles: 4},
	0xBE: {operation: LDX, addressingMode: AbsoluteY, cycles: 4},
	// LDY
	0xA0: {operation: LDY, addressingMode: Immediate, cycles: 2},
	0xA4: {operation: LDY, addressingMode: ZeroPage, cycles: 3},
	0xB4: {operation: LDY, addressingMode: ZeroPageX, cycles: 4},
	0xAC: {operation: LDY, addressingMode: Absolute, cycles: 4},
	0xBC: {operation: LDY, addressingMode: AbsoluteX, cycles: 4},
	// LSR
	0x4A: {operation: LSR, addressingMode: Accumulator, cycles: 2},
	0x46: {operation: LSR, addressingMode: ZeroPage, cycles: 5},
	0x56: {operation: LSR, addressingMode: ZeroPageX, cycles: 6},
	0x4E: {operation: LSR, addressingMode: Absolute, cycles: 6},
	0x5E: {operation: LSR, addressingMode: AbsoluteX, cycles: 7},
	// NOP
	0xEA: {operation: NOP, addressingMode: Implied, cycles: 2},
	// ORA
	0x09: {operation: ORA, addressingMode: Immediate, cycles: 2},
	0x05: {operation: ORA, addressingMode: ZeroPage, cycles: 3},
	0x15: {operation: ORA, addressingMode: ZeroPageX, cycles: 4},
	0x0D: {operation: ORA, addressingMode: Absolute, cycles: 4},
	0x1D: {operation: ORA, addressingMode: AbsoluteX, cycles: 4},
	0x19: {operation: ORA, addressingMode: AbsoluteY, cycles: 4},
	0x01: {operation: ORA, addressingMode: IndirectX, cycles: 6},
	0x11: {operation: ORA, addressingMode: IndirectY, cycles: 5},
	// PHA
	0x48: {operation: PHA, addressingMode: Implied, cycles: 3},
	// PHP
	0x08: {operation: PHP, addressingMode: Implied, cycles: 3},
	// PLA
	0x68: {operation: PLA, addressingMode: Implied, cycles: 4},
	// PLP
	0x28: {operation: PLP, addressingMode: Implied, cycles: 4},
	// ROL
	0x2A: {operation: ROL, addressingMode: Accumulator, cycles: 2},
	0x26: {operation: ROL, addressingMode: ZeroPage, cycles: 5},
	0x36: {operation: ROL, addressingMode: ZeroPageX, cycles: 6},
	0x2E: {operation: ROL, addressingMode: Absolute, cycles: 6},
	0x3E: {operation: ROL, addressingMode: AbsoluteX, cycles: 7},
	// ROR
	0x6A: {operation: ROR, addressingMode: Accumulator, cycles: 2},
	0x66: {operation: ROR, addressingMode: ZeroPage, cycles: 5},
	0x76: {operation: ROR, addressingMode: ZeroPageX, cycles: 6},
	0x6E: {operation: ROR, addressingMode: Absolute, cycles: 6},
	0x7E: {operation: ROR, addressingMode: AbsoluteX, cycles: 7},
	// RTI
	0x40: {operation: RTI, addressingMode: Implied, cycles: 6},
	// RTS
	0x60: {operation: RTS, addressingMode: Implied, cycles: 6},
	// SBC
	0xE9: {operation: SBC, addressingMode: Immediate, cycles: 2},
	0xE5: {operation: SBC, addressingMode: ZeroPage, cycles: 3},
	0xF5: {operation: SBC, addressingMode: ZeroPageX, cycles: 4},
	0xED: {operation: SBC, addressingMode: Absolute, cycles: 4},
	0xFD: {operation: SBC, addressingMode: AbsoluteX, cycles: 4},
	0xF9: {operation: SBC, addressingMode: AbsoluteY, cycles: 4},
	0xE1: {operation: SBC, addressingMode: IndirectX, cycles: 6},
	0xF1: {operation: SBC, addressingMode: IndirectY, cycles: 5},
	// SEC
	0x38: {operation: SEC, addressingMode: Implied, cycles: 2},
	// SED
	0xF8: {operation: SED, addressingMode: Implied, cycles: 2},
	// SEI
	0x78: {operation: SEI, addressingMode: Implied, cycles: 2},
	// STA
	0x85: {operation: STA, addressingMode: ZeroPage, cycles: 3},
	0x95: {operation: STA, addressingMode: ZeroPageX, cycles: 4},
	0x8D: {operation: STA, addressingMode: Absolute, cycles: 4},
	0x9D: {operation: STA, addressingMode: AbsoluteX, cycles: 5},
	0x99: {operation: STA, addressingMode: AbsoluteY, cycles: 5},
	0x81: {operation: STA, addressingMode: IndirectX, cycles: 6},
	0x91: {operation: STA, addressingMode: IndirectY, cycles: 6},
	// STX
	0x86: {operation: STX, addressingMode: ZeroPage, cycles: 3},
	0x96: {operation: STX, addressingMode: ZeroPageY, cycles: 4},
	0x8E: {operation: STX, addressingMode: Absolute, cycles: 4},
	// STY
	0x84: {operation: STY, addressingMode: ZeroPage, cycles: 3},
	0x94: {operation: STY, addressingMode: ZeroPageX, cycles: 4},
	0x8C: {operation: STY, addressingMode: Absolute, cycles: 4},
	// TAX
	0xAA: {operation: TAX, addressingMode: Implied, cycles: 2},
	// TAY
	0xA8: {operation: TAY, addressingMode: Implied, cycles: 2},
	// TSX
	0xBA: {operation: TSX, addressingMode: Implied, cycles: 2},
	// TXA
	0x8A: {operation: TXA, addressingMode: Implied, cycles: 2},
	// TXS
	0x9A: {operation: TXS, addressingMode: Implied, cycles: 2},
	// TYA
	0x98: {operation: TYA, addressingMode: Implied, cycles: 2},
}

func matchOpHexCodeWithOpCode(hexCode uint8) OpCode {
	var opsCode, ok = hexToOpsCode[hexCode]
	if !ok {
		panic(fmt.Sprintf("hex code %v is unsupported", hexCode))
	}
	return opsCode
}
