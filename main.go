package main

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////
//
// the virtual machine

type ProgramAddress uint8
type DataAddress uint8

const MaxProgramAddress = ProgramAddress(^ProgramAddress(0))
const MaxDataAddress = ^DataAddress(0)

type Instruction struct {
	a DataAddress
	b DataAddress
	c DataAddress
	d ProgramAddress
}

type DataCell int8

type Computer struct {
	ip             ProgramAddress
	program_memory [MaxProgramAddress]Instruction
	data_memory    [uint(MaxDataAddress) + 1]DataCell
}

func (self *Computer) load_program(code []Instruction) {
	for i, instr := range code {
		self.program_memory[i] = instr
	}
}

func (self *Computer) load_data(data []DataCell) {
	for i, cell := range data {
		self.data_memory[i] = cell
	}
	self.data_memory[ZERO] = 0
	self.data_memory[ONE] = 1
}

func (self *Computer) Halted() bool {
	return (self.ip == MaxProgramAddress)
}

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

// NOTE: maybe only ONE is strictly required, ZERO is just ONE - ONE
const JUNK DataAddress = MaxDataAddress
const ONE DataAddress = MaxDataAddress - 2
const ZERO DataAddress = MaxDataAddress - 1

var SP DataAddress = MaxDataAddress - 3

////////////////////////////////////////////////////////////////////////
//
// the assembler

type ILabel interface {
	GetAddress(a Assembler) ProgramAddress
}

type Label string

func (self ProgramAddress) GetAddress(a Assembler) ProgramAddress {
	return self
}

func (self Label) GetAddress(a Assembler) ProgramAddress {
	return a.labels[self]
}

type PseudoInstruction struct {
	a DataAddress
	b DataAddress
	c DataAddress
	d ILabel
}

type Assembler struct {
	ip           ProgramAddress
	labels       map[Label]ProgramAddress
	instructions [MaxProgramAddress]PseudoInstruction
}

func NewAssembler() Assembler {
	ass := Assembler{}
	ass.labels = make(map[Label]ProgramAddress)
	return ass
}

func (self *Assembler) add(a, b, c DataAddress, d ILabel) {
	self.instructions[self.ip] = PseudoInstruction{a: a, b: b, c: c, d: d}
	self.ip++
}

func (self *Assembler) label(label Label) {
	self.labels[label] = self.ip
}

func (self *Assembler) assemble() []Instruction {
	res := make([]Instruction, self.ip)
	for i, instr := range self.instructions[:self.ip] {
		res[i] = Instruction{instr.a, instr.b, instr.c, instr.d.GetAddress(*self)}
	}
	return res
}

func (self *Assembler) SBNZ(a, b, c DataAddress, d ILabel) {
	self.add(a, b, c, d)
}

func (self *Assembler) JMP(a ILabel) {
	self.add(ONE, ZERO, JUNK, a)
}

func (self *Assembler) HLT() {
	self.add(ONE, ZERO, JUNK, MaxProgramAddress)
}

func (self *Assembler) NEG(a, b DataAddress) {
	self.add(ZERO, a, b, self.ip+1)
}

func (self *Assembler) ADD(a, b, c DataAddress) {
	self.NEG(b, JUNK)
	self.add(a, JUNK, c, self.ip+1)
}

func (self *Assembler) MOV(a, b DataAddress) {
	self.add(a, ZERO, b, self.ip+1)
}

func (self *Assembler) BEQ(a, b DataAddress, c ILabel) {
	self.add(a, b, JUNK, self.ip+2)
	self.JMP(c)
}

func (self *Assembler) NOP() {
	self.add(JUNK, JUNK, JUNK, self.ip+1)
}

func (self *Assembler) DEC(a DataAddress) {
	self.add(a, ONE, a, self.ip+1)
}

func main() {
	const COUNTER DataAddress = 3
	const RESULT DataAddress = 2
	ass := NewAssembler()
	ass.MOV(0, COUNTER)
	ass.MOV(ZERO, RESULT)
	ass.label("loop")
	ass.BEQ(COUNTER, ZERO, Label("exit_loop"))
	ass.ADD(1, RESULT, RESULT)
	ass.DEC(COUNTER)
	ass.JMP(Label("loop"))
	ass.label("exit_loop")
	ass.HLT()

	data := []DataCell{
		2, 3,
	}
	computer := new(Computer)
	computer.load_program(ass.assemble())
	computer.load_data(data)

	computer.PrintProgramMemory()
	computer.PrintDataMemory(4)
	for !computer.Halted() {
		computer.Step()
	}
	computer.PrintDataMemory(4)
}
