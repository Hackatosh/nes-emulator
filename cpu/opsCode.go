package cpu

import "fmt"

// Addressing modes

type AddressingMode int

const (
	Implicit AddressingMode = iota
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

var hexToOpsCode = map[uint8]OpCode{
	0x00: {operation: BRK, addressingMode: Implicit, bytes: 1, cycles: 7},
	0x85: {operation: STA, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0x95: {operation: STA, addressingMode: ZeroPageX, bytes: 2, cycles: 4},
	0xA9: {operation: LDA, addressingMode: Immediate, bytes: 2, cycles: 2},
	0xA5: {operation: LDA, addressingMode: ZeroPage, bytes: 2, cycles: 3},
	0xAD: {operation: LDA, addressingMode: Absolute, bytes: 3, cycles: 4},
	0xAA: {operation: TAX, addressingMode: Implicit, bytes: 1, cycles: 2},
	0xE8: {operation: INX, addressingMode: Implicit, bytes: 1, cycles: 2},
}

func matchHexCodeWithOpsCode(hexCode uint8) OpCode {
	var opsCode, ok = hexToOpsCode[hexCode]
	if !ok {
		panic(fmt.Sprintf("hex code %v is unsupported", hexCode))
	}
	return opsCode
}
