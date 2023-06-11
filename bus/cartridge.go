package bus

import (
	"bytes"
	"errors"
)

const PRG_ROM_PAGE_SIZE int = 16384
const CHR_ROM_PAGE_SIZE int = 8192

type ScreenMirroring int

const (
	VERTICAL ScreenMirroring = iota
	HORIZONTAL
	FOUR_SCREEN
)

type Rom struct {
	prgRom          []uint8
	chrRom          []uint8
	mapper          uint8
	screenMirroring ScreenMirroring
}

func ParseRawRom(raw []byte) (*Rom, error) {
	/* PARSING HEADERS */
	var nesTag = raw[0:4]
	var numberOfROMBanks = int(raw[4])  // PRG ROM
	var numberOfVROMBanks = int(raw[5]) // CHR ROM
	var isVerticalMirroring = raw[6]&0b0000_0001 != 0
	// Unused in our emulator
	//var isBatteryBackedRAMEnabled = raw[6] & 0b0000_0010 != 0
	var isTrainerEnabled = raw[6]&0b0000_0100 != 0
	var isFourScreenEnabled = raw[6]&0b0000_1000 != 0
	var mapper = (raw[6] >> 4) | (raw[7] & 0b1111_0000)
	// TODO : this does not work ??
	//var isVerifiedINESV1 = raw[7]&0b0000_0011 == 0
	var isINESV2 = raw[7]&0b0000_1100 != 0

	/* SANITY CHECKS */

	if !bytes.Equal(nesTag, []byte{0x4E, 0x45, 0x53, 0x1A}) {
		return &Rom{}, errors.New("file is not in iNES file format (invalid tag)")
	}

	if isINESV2 {
		return &Rom{}, errors.New("iNES v2 is not supported")
	}

	//if isVerifiedINESV1 {
	//	return rom{}, errors.New("control bites for iNes v1 are incorrect")
	//}

	/* Building ROM */

	var screenMirroring ScreenMirroring
	switch {
	case isFourScreenEnabled:
		screenMirroring = FOUR_SCREEN
	case isVerticalMirroring:
		screenMirroring = VERTICAL
	default:
		screenMirroring = HORIZONTAL
	}

	var prgROMSize = numberOfROMBanks * PRG_ROM_PAGE_SIZE
	var chrROMSize = numberOfVROMBanks * CHR_ROM_PAGE_SIZE
	var prgROMStart = 16
	if isTrainerEnabled {
		prgROMStart += 512 // Trainer is of fixed size 512 bytes
	}
	var chrROMStart = prgROMStart + prgROMSize
	return &Rom{
		prgRom:          raw[prgROMStart : prgROMStart+prgROMSize],
		chrRom:          raw[chrROMStart : chrROMStart+chrROMSize],
		mapper:          mapper,
		screenMirroring: screenMirroring,
	}, nil
}
