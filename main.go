package main

import (
	"nes-emulator/bus"
	"nes-emulator/nes_console"
	"os"
)

const ROM_PATH string = "resources\\nestest.nes"

func main() {
	var rawRom, errorRead = os.ReadFile(ROM_PATH)
	if errorRead != nil {
		panic(errorRead)
	}
	var rom, errorParse = bus.ParseRawRom(rawRom)
	if errorParse != nil {
		panic(errorParse)
	}
	var console = nes_console.NewConsole()
	console.RunRom(rom)
}
