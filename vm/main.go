package vm

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////
//
// the virtual machine
//
// - Separated address spaces for program and data.
//
// - Uses uint8 to index memory, so each block of memory is 256 words
//   long.
//
// - words in the program block are of type Instruction (32 bits long)
//   while words in data block are of type DataCell (8 bits long).

type ProgramAddress uint8
type DataAddress uint8

const MaxProgramAddress = ProgramAddress(^ProgramAddress(0))
const MaxDataAddress = ^DataAddress(0)

// NOTE: maybe only ONE is strictly required, ZERO is just ONE - ONE
const JUNK DataAddress = MaxDataAddress
const ONE DataAddress = MaxDataAddress - 2
const ZERO DataAddress = MaxDataAddress - 1

type Instruction struct {
	a DataAddress
	b DataAddress
	c DataAddress
	d ProgramAddress
}

func NewInstruction(a, b, c DataAddress, d ProgramAddress) Instruction {
	return Instruction{a: a, b: b, c: c, d: d}
}

type DataCell int8

type Computer struct {
	ip             ProgramAddress
	program_memory [MaxProgramAddress]Instruction
	data_memory    [uint(MaxDataAddress) + 1]DataCell
}

// LoadProgram loads the code into the program memory
func (self *Computer) LoadProgram(code []Instruction) {
	for i, instr := range code {
		self.program_memory[i] = instr
	}
}

// LoadData load the data into the data memory. Takes care of
// initializing the ZERO and ONE addresses, after the data has been
// loaded, so it may overwrite those memory cells.
func (self *Computer) LoadData(data []DataCell) {
	for i, cell := range data {
		self.data_memory[i] = cell
	}
	self.data_memory[ZERO] = 0
	self.data_memory[ONE] = 1
}

// Halted return true if the computer is halted
func (self *Computer) Halted() bool {
	return (self.ip == MaxProgramAddress)
}

// Step if the computer is not halted executes the next instruction,
// updates the IP pointer and returns
func (self *Computer) Step() {
	if !self.Halted() {
		instr := self.program_memory[self.ip]
		a := self.data_memory[instr.a]
		b := self.data_memory[instr.b]
		r := a - b
		self.data_memory[instr.c] = r
		if r != 0 {
			self.ip = instr.d
		} else {
			self.ip++
		}
	}
}

// PrintDataMemory prints the first n bytes of the data memory plus
// the three higher bytes
func (self *Computer) PrintDataMemory(n DataAddress) {
	if n >= ONE {
		n = ONE - 1
	}
	fmt.Println("IP=", self.ip)
	for i, cell := range self.data_memory[:n] {
		fmt.Printf("%02X ", cell)
		if (i+1)%28 == 0 {
			fmt.Println("")
		}
	}
	if n < ONE-1 {
		fmt.Print("... ")
	}
	fmt.Printf("%02x %02x %02x", self.data_memory[ONE], self.data_memory[ZERO], self.data_memory[JUNK])
	fmt.Println("")
}

// formatDataAddress if 'a' points to some of the reserved addresses a
// symbolic representation is used (j for JUNK, o for ONE and z for
// ZERO) otherwise the numerical representation of the address is used.
func formatDataAddress(a DataAddress) string {
	switch a {
	case JUNK:
		return "j"
	case ONE:
		return "o"
	case ZERO:
		return "z"
	default:
		return fmt.Sprintf("%d", a)
	}
}

// formatProgramAddress if 'a' points to the halt address a symbolic
// representation is used (h) otherwise the numerical representation
// of the address is used.
func formatProgramAddress(a ProgramAddress) string {
	if a == MaxProgramAddress {
		return "h"
	} else {
		return fmt.Sprintf("%d", a)
	}
}

// PrintProgramMemory prints to stdout the program memory contents
// conveniently formatted.
func (self *Computer) PrintProgramMemory() {
	fmt.Println("Program memory dump")
	fmt.Println(" IP   A   B   C   D")
	skipping := false
	for i, tmp := range self.program_memory {
		if tmp.a == 0 && tmp.b == 0 && tmp.c == 0 && tmp.d == 0 {
			if !skipping {
				fmt.Println("    ...")
				skipping = true
			}
		} else {
			skipping = false
			a := formatDataAddress(tmp.a)
			b := formatDataAddress(tmp.b)
			c := formatDataAddress(tmp.c)
			d := formatProgramAddress(tmp.d)
			fmt.Printf("%3d %3s %3s %3s %3s\n", i, a, b, c, d)
		}
	}
}
