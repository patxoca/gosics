package main

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

// load_program loads the code into the program memory
func (self *Computer) load_program(code []Instruction) {
	for i, instr := range code {
		self.program_memory[i] = instr
	}
}

// load_data load the data into the data memory. Takes care of
// initializing the ZERO and ONE addresses, after the data has been
// loaded, so it may overwrite those memory cells.
func (self *Computer) load_data(data []DataCell) {
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

////////////////////////////////////////////////////////////////////////
//
// the assembler
//
// - supports labels to reference addresses in program memory
//
// - stores SBNZ instructions, maybe with unresolved references to
//   labels
//
//

// The ILabel interface is provided by all types that can be used as
// addresses in an assembler program, either literal addresses
// (ProgramAddress) or symbolic addresses (Label).
type ILabel interface {
	GetAddress(a Assembler) ProgramAddress
}

type Label string

// GetAddress return the address
func (self ProgramAddress) GetAddress(a Assembler) ProgramAddress {
	return self
}

// GetAddress perform a lookup of the label in the label table and
// return the corresponding address. Panic if the label is not found.
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
	label_cnt    int
}

// NewAssembler create a new Assembler instance and initialized
// internal structures. Don't create an Assembler directly!!
func NewAssembler() Assembler {
	ass := Assembler{}
	ass.labels = make(map[Label]ProgramAddress)
	return ass
}

// SBNZ adds a new SBNZ instruction to the program with a program
// address possibly unresolved.
func (self *Assembler) SBNZ(a, b, c DataAddress, d ILabel) {
	self.instructions[self.ip] = PseudoInstruction{a: a, b: b, c: c, d: d}
	self.ip++
}

// make_label create a unique label. It's intended to be used in macro
// instructions that use other macro instructions. Avoids the
// requirement of knowing before hand how much instructions to skip.
func (self *Assembler) make_label() Label {
	self.label_cnt++
	return Label(fmt.Sprintf("__label_%04d", self.label_cnt))
}

// label defines a label named 'label' for the current instruction
// pointer
func (self *Assembler) label(label Label) {
	self.labels[label] = self.ip
}

// assemble resolves unresolved program addresses and retuns a valid
// program (an slice of Instruction).
func (self *Assembler) assemble() []Instruction {
	res := make([]Instruction, self.ip)
	for i, instr := range self.instructions[:self.ip] {
		res[i] = Instruction{instr.a, instr.b, instr.c, instr.d.GetAddress(*self)}
	}
	return res
}

// Sinthetized instructions
//
// The following methods define macro instructions for some usual
// opcodes in terms of the SBNZ instruction. Unless stated the
// execution continues in the next instruction.

// JMP incoditional jump to 'a'
func (self *Assembler) JMP(a ILabel) {
	self.SBNZ(ONE, ZERO, JUNK, a)
}

// HLT halt execution
func (self *Assembler) HLT() {
	self.SBNZ(ONE, ZERO, JUNK, MaxProgramAddress)
}

// NEG negate the content of 'a' and store the result in 'b'. 'a' and
// 'b' may point to the same data address.
func (self *Assembler) NEG(a, b DataAddress) {
	self.SBNZ(ZERO, a, b, self.ip+1)
}

// ADD add content of 'a' to content of 'b' and store the result in
// 'c'. 'a', 'b' and 'c' may point to the same data address.
func (self *Assembler) ADD(a, b, c DataAddress) {
	self.NEG(b, JUNK)
	self.SBNZ(a, JUNK, c, self.ip+1)
}

// MOV copy content of 'a' to 'b'.
func (self *Assembler) MOV(a, b DataAddress) {
	self.SBNZ(a, ZERO, b, self.ip+1)
}

// BEQ branch execution to 'c' if contents of 'a' and 'b' are equal.
func (self *Assembler) BEQ(a, b DataAddress, c ILabel) {
	label := self.make_label()
	self.SBNZ(a, b, JUNK, label)
	self.JMP(c)
	self.label(label)
}

// NOP do nothing
func (self *Assembler) NOP() {
	self.SBNZ(JUNK, JUNK, JUNK, self.ip+1)
}

// DEC decrement content of 'a'
func (self *Assembler) DEC(a DataAddress) {
	self.SBNZ(a, ONE, a, self.ip+1)
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
