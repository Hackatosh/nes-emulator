package main

import (
	"fmt"
	"nes-emulator/bus"
	"nes-emulator/nes_console"
	"os"
)

const ROM_PATH string = "resources\\nestest.nes"

func main() {
	fmt.Println(fmt.Sprintf("Reading rom  file at path %s...", ROM_PATH))
	var rawRom, errorRead = os.ReadFile(ROM_PATH)
	if errorRead != nil {
		panic(errorRead)
	}

	fmt.Println("Parsing rom...")
	var rom, errorParse = bus.ParseRawRom(rawRom)
	if errorParse != nil {
		panic(errorParse)
	}

	fmt.Println("Running rom in nes emulator...")
	var console = nes_console.NewConsole()
	console.RunRom(rom)
}
