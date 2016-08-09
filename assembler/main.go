package assembler

import (
	"fmt"
	"gosics/vm"
)

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

type ProgramAddress vm.ProgramAddress

// The ILabel interface is provided by all types that can be used as
// addresses in an assembler program, either literal addresses
// (ProgramAddress) or symbolic addresses (Label).
type ILabel interface {
	GetAddress(a Assembler) vm.ProgramAddress
}

type Label string

// GetAddress return the address
func (self ProgramAddress) GetAddress(a Assembler) vm.ProgramAddress {
	return vm.ProgramAddress(self)
}

// GetAddress perform a lookup of the label in the label table and
// return the corresponding address. Panic if the label is not found.
func (self Label) GetAddress(a Assembler) vm.ProgramAddress {
	return vm.ProgramAddress(a.labels[self])
}

type PseudoInstruction struct {
	a vm.DataAddress
	b vm.DataAddress
	c vm.DataAddress
	d ILabel
}

type Assembler struct {
	ip           ProgramAddress
	labels       map[Label]ProgramAddress
	instructions [vm.MaxProgramAddress]PseudoInstruction
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
func (self *Assembler) SBNZ(a, b, c vm.DataAddress, d ILabel) {
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
func (self *Assembler) Assemble() []vm.Instruction {
	res := make([]vm.Instruction, self.ip)
	for i, instr := range self.instructions[:self.ip] {
		res[i] = vm.NewInstruction(instr.a, instr.b, instr.c, instr.d.GetAddress(*self))
	}
	return res
}

// Sinthetized instructions
//
// The following methods define macro instructions for some usual
// opcodes in terms of the SBNZ instruction. Unless otherwise stated
// the execution continues in the next instruction.

// MOV copy content of 'a' to 'b'.
func (self *Assembler) MOV(a, b vm.DataAddress) {
	self.SBNZ(a, vm.ZERO, b, self.ip+1)
}

// ----------------------------------------------------- flow control

// JMP incoditional jump to 'a'
func (self *Assembler) JMP(a ILabel) {
	self.SBNZ(vm.ONE, vm.ZERO, vm.JUNK, a)
}

// BEQ branch execution to 'c' if contents of 'a' and 'b' are equal.
func (self *Assembler) BEQ(a, b vm.DataAddress, c ILabel) {
	label := self.make_label()
	self.SBNZ(a, b, vm.JUNK, label)
	self.JMP(c)
	self.label(label)
}

// ------------------------------------------------- assorted opcodes

// HLT halt execution
func (self *Assembler) HLT() {
	self.SBNZ(vm.ONE, vm.ZERO, vm.JUNK, ProgramAddress(vm.MaxProgramAddress))
}

// NOP do nothing
func (self *Assembler) NOP() {
	self.SBNZ(vm.JUNK, vm.JUNK, vm.JUNK, self.ip+1)
}

// ----------------------------------------------- arithmetic opcodes

// NEG negate the content of 'a' and store the result in 'b'. 'a' and
// 'b' may point to the same data address.
func (self *Assembler) NEG(a, b vm.DataAddress) {
	self.SBNZ(vm.ZERO, a, b, self.ip+1)
}

// ADD add content of 'a' to content of 'b' and store the result in
// 'c'. 'a', 'b' and 'c' may point to the same data address.
func (self *Assembler) ADD(a, b, c vm.DataAddress) {
	self.NEG(b, vm.JUNK)
	self.SBNZ(a, vm.JUNK, c, self.ip+1)
}

// SUB substract contets of 'b' from 'a' and stores the result in 'c'.
func (self *Assembler) SUB(a, b, c vm.DataAddress) {
	self.SBNZ(a, b, c, self.ip+1)
}

func (self *Assembler) INC(a vm.DataAddress) {
	self.ADD(a, vm.ONE, a)
}

// DEC decrement content of 'a'
func (self *Assembler) DEC(a vm.DataAddress) {
	self.SBNZ(a, vm.ONE, a, self.ip+1)
}

// MUL multiplies content of 'a' by 'b' and stores the result in 'c'.
// The content of 'a' is lost. Does not check for overflow.
func (self *Assembler) MUL(a, b, c vm.DataAddress) {
	loop := self.make_label()
	exit_loop := self.make_label()

	self.MOV(vm.ZERO, c)
	self.label(loop)
	self.BEQ(a, vm.ZERO, exit_loop)
	self.ADD(b, c, c)
	self.DEC(a)
	self.JMP(loop)
	self.label(exit_loop)
}

// -------------------------------------------------- logical opcodes

// NOT perform the bitwise not on the contents of 'a' and stores the
// result in 'b'.
func (self *Assembler) NOT(a, b vm.DataAddress) {
	self.ADD(a, vm.ONE, b)
	self.NEG(b, b)
}

// --------------------------------------------- emulating other OISC

// SUBLEQ Subtract and branch if less than or equal to zero OISC.
// Substrat content of address 'a' from content of 'b' and stores the
// result en address 'c'. If the result is less than or equal to zero
// jump to address 'c'.
func (self *Assembler) SUBLEQ(a, b vm.DataAddress, c ProgramAddress) {
	// not sure how to test 'value < 0'
}
