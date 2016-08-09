package main

import (
	"gosics/assembler"
	"gosics/vm"
)

func main() {
	const COUNTER vm.DataAddress = 3
	const RESULT vm.DataAddress = 2
	ass := assembler.NewAssembler()
	ass.MUL(0, 1, 2)
	ass.HLT()

	data := []vm.DataCell{
		2, 3,
	}
	computer := new(vm.Computer)
	computer.LoadProgram(ass.Assemble())
	computer.LoadData(data)

	computer.PrintProgramMemory()
	computer.PrintDataMemory(4)
	for !computer.Halted() {
		computer.Step()
	}
	computer.PrintDataMemory(4)
}
