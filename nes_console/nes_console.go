package nes_console

import (
	"nes-emulator/bus"
	"nes-emulator/cpu"
)

type NesConsole struct {
	bus bus.Bus
	cpu cpu.CPU
}

func NewConsole() NesConsole {
	var consoleBus = bus.NewBus()
	var consoleCPU = cpu.NewCPU(consoleBus)
	return NesConsole{
		bus: consoleBus,
		cpu: consoleCPU,
	}
}

func (console *NesConsole) RunRom(rom bus.Rom) {
	console.bus.LoadRom(rom)
	console.cpu.Reset()
	console.cpu.Run()
}
