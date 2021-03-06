// This package implements an in memory assembler:
//
// - supports labels to reference memory addresses. Labels prefixed
// with double underscore are reserved for internal use.
//
// - stores SBNZ instructions, maybe with unresolved references to
// addresses
//
// - provides directives DB and DD to store data in memory
//
// Example:
//    ass := assembler.New()
//
//    OP1 := assembler.Label("OP1")
//    OP2 := assembler.Label("OP2")
//    DST := assembler.Label("DST")
//
//    ass.SBNZ(OP1, OP2, DST, assembler.HLT)
//    ass.Label(OP1)
//    ass.DD(0x01)
//    ass.Label(OP2)
//    ass.DD(0x02)
//    ass.Label(DST)
//    ass.DD(0x00)
//
package assembler

import (
	"container/list"
	"fmt"
	"gosics/vm"
)

// Label symbolic name for an address
type Label string

// Address type for adresses
type Address vm.Address

const maxAddress = Address(vm.MaxAddress)

// HLT is the address to jump to in order to halt the computer
const HLT = maxAddress

// ONE is a label to a memory position containing a 1
const ONE = Label("__ONE")

// ZERO is a label to a memory position containing a 0. It's not
// strictrly required since we can get a 0 executing SBNZ ONE, ONE, x,
// y, but it's convenient
const ZERO = Label("__ZERO")

// JUNK is a label to a memory position where we can store temporary
// results
const JUNK = Label("__JUNK")

// Assembler in memory assembler
type Assembler struct {
	ip         Address
	labels     map[Label]Address
	unresolved map[Label]*list.List
	memory     [vm.MemorySize]uint8
	label_cnt  int
}

// The labeler interface is provided by all types that can be used as
// addresses in an assembler program, either literal addresses
// (Address) or symbolic addresses (Label).
type labeler interface {
	getAddress(a *Assembler) Address
}

// getAddress return the address for a literal address
func (self Address) getAddress(a *Assembler) Address {
	return self
}

// getAddress perform a lookup for the label in the label table and
// return the corresponding address. If the label is not found adds it
// to the unresolved-labels table and return a fake address.
func (self Label) getAddress(a *Assembler) Address {
	address, ok := a.labels[self]
	if ok {
		return address
	}
	l, ok := a.unresolved[self]
	if !ok {
		l = list.New()
		a.unresolved[self] = l
	}
	l.PushBack(a.ip)
	return Address(vm.MaxAddress)
}

// New create a new Assembler instance and initializes internal
// structures. Don't create an Assembler directly!!
func New() Assembler {
	ass := Assembler{}
	ass.labels = make(map[Label]Address)
	ass.unresolved = make(map[Label]*list.List)
	start := Label("__start")
	ass.SBNZ(ONE, ZERO, JUNK, start)

	ass.Label(ONE)
	ass.DD(1)
	ass.Label(ZERO)
	ass.DD(0)
	ass.Label(JUNK)
	ass.DD(0)

	ass.Label(Label("__push_operand"))
	ass.DD(0xFABA)
	ass.Label(Label("__SP"))
	ass.DD(uint16(maxAddress - 1))
	ass.Label(Label("__push"))
	// copy SP in the C parameter of the next instruction
	ass.SBNZ(Label("__SP"), ZERO, ass.ip+12, ass.ip+8)
	// copy value from __push_operand to the stack. The C operand has
	// been overwriten so that it point to the top of the stack
	ass.SBNZ(Label("__push_operand"), ZERO, maxAddress-1, ass.ip+8)
	// decrease the stack pointer twice
	ass.SBNZ(Label("__SP"), ONE, Label("__SP"), ass.ip+8)
	ass.SBNZ(Label("__SP"), ONE, Label("__SP"), ass.ip+8)
	// "return" to the caller. He caller must copy in __push_ret the
	// return address
	ass.DD(uint16(ass.labels[ONE]), uint16(ass.labels[ZERO]), uint16(ass.labels[JUNK]))
	ass.Label(Label("__push_ret"))
	ass.DD(uint16(0xFFFF))

	ass.Label(Label("__pop"))
	// increase the stack pointer twice, first we need -1 (SP - -1 ==
	// SP + 1)
	ass.SBNZ(ZERO, ONE, JUNK, ass.ip+8)
	ass.SBNZ(Label("__SP"), JUNK, Label("__SP"), ass.ip+8)
	ass.SBNZ(Label("__SP"), JUNK, Label("__SP"), ass.ip+8)
	// copy SP in the A parameter of the next instruction
	ass.SBNZ(Label("__SP"), ZERO, ass.ip+8, ass.ip+8)
	// copy the value from the stack to __push_operand
	ass.SBNZ(maxAddress-1, ZERO, Label("__push_operand"), ass.ip+8)
	// return to the "caller"
	ass.DD(uint16(ass.labels[ONE]), uint16(ass.labels[ZERO]), uint16(ass.labels[JUNK]))
	ass.Label(Label("__pop_ret"))
	ass.DD(uint16(0xFFFF))

	ass.Label(start)
	return ass
}

// uniqLabel create a unique label. It's intended to be used in macro
// instructions that use other macro instructions and need to branch.
// Avoids the requirement of knowing before hand how much instructions
// to skip.
func (self *Assembler) uniqLabel() Label {
	self.label_cnt++
	label := Label(fmt.Sprintf("__label_%04d", self.label_cnt))
	return label
}

// Label define a label pointing to the current IP.
// TODO: maybe the argument can be just a string
func (self *Assembler) Label(label Label) {
	self.labels[label] = self.ip
}

// TODO: define methods GetStorage and FreeStorage for allocating
// temporary storage in a stack. Intended to be used for macro
// instructions that require temporary storage.

// Assemble resolves unresolved program addresses and retuns a valid
// program.
func (self *Assembler) Assemble() []uint8 {
	res := make([]uint8, self.ip)
	copy(res, self.memory[:self.ip])
	for lab, lst := range self.unresolved {
		// TODO: manage undefined labels
		a := self.labels[lab]
		ah := uint8(a >> 8)
		al := uint8(a & 0xFF)
		for e := lst.Front(); e != nil; e = e.Next() {
			i := e.Value.(Address)
			res[i] = ah
			res[i+1] = al
		}
	}
	return res
}

//////////////////////////////////////////////////////////////////////////
// Assembler directives

// DB insert a sequence of bytes into memory at IP, updates IP
func (self *Assembler) DB(bytes ...uint8) {
	for _, b := range bytes {
		self.memory[self.ip] = b
		self.ip++
	}
}

// DD insert a sequence of 16 bits values into memory at IP, updates
// IP
func (self *Assembler) DD(words ...uint16) {
	for _, d := range words {
		self.memory[self.ip] = uint8(d >> 8)
		self.memory[self.ip+1] = uint8(d & 0xFF)
		self.ip += 2
	}
}

//////////////////////////////////////////////////////////////////////////
// Assembler opcodes

// SBNZ adds a new SBNZ instruction to the program and advances the
// IP.
func (self *Assembler) SBNZ(a, b, c, d labeler) {
	for _, v := range [4]labeler{a, b, c, d} {
		addr := v.getAddress(self)
		self.memory[self.ip] = uint8(addr >> 8)
		self.memory[self.ip+1] = uint8(addr & 0xFF)
		self.ip += 2
	}
}

// Sinthetized instructions
//
// The following methods define macro instructions for some usual
// opcodes in terms of the SBNZ instruction. Unless otherwise stated
// the execution continues in the next instruction.

// MOV copy content of 'a' to 'b'.
func (self *Assembler) MOV(src, dst labeler) {
	label := self.uniqLabel()
	self.SBNZ(src, ZERO, dst, label)
	self.Label(label)
}

// ----------------------------------------------------- flow control

// JMP incoditional jump to 'a'
func (self *Assembler) JMP(a labeler) {
	self.SBNZ(ONE, ZERO, JUNK, a)
}

// BEQ branch execution to 'c' if contents of 'a' and 'b' are equal.
func (self *Assembler) BEQ(a, b, dst labeler) {
	label := self.uniqLabel()
	self.SBNZ(a, b, JUNK, label)
	self.JMP(dst)
	self.Label(label)
}

// ------------------------------------------------- assorted opcodes

// HLT halt execution
func (self *Assembler) HLT() {
	self.SBNZ(ONE, ZERO, JUNK, maxAddress)
}

// NOP do nothing
func (self *Assembler) NOP() {
	label := self.uniqLabel()
	self.SBNZ(JUNK, JUNK, JUNK, label)
	self.Label(label)
}

// ----------------------------------------------- arithmetic opcodes

// NEG negate the content of src and store the result in dst. src and
// dst may point to the same address.
func (self *Assembler) NEG(src, dst labeler) {
	label := self.uniqLabel()
	self.SBNZ(ZERO, src, dst, label)
	self.Label(label)
}

// ADD add content of a to content of b and store the result in
// dst. a, b and c may point to the same data address.
func (self *Assembler) ADD(a, b, dst labeler) {
	label := self.uniqLabel()
	self.NEG(b, JUNK)
	self.SBNZ(a, JUNK, dst, label)
	self.Label(label)
}

// SUB substract content of b from a and stores the result in dst.
func (self *Assembler) SUB(a, b, dst labeler) {
	label := self.uniqLabel()
	self.SBNZ(a, b, dst, label)
	self.Label(label)
}

// INC increments content of 'a'
func (self *Assembler) INC(a labeler) {
	self.ADD(a, ONE, a)
}

// DEC decrement content of 'a'
func (self *Assembler) DEC(a labeler) {
	label := self.uniqLabel()
	self.SBNZ(a, ONE, a, label)
	self.Label(label)
}

// // MUL multiplies content of 'a' by 'b' and stores the result in 'c'.
// // The content of 'a' is lost. Does not check for overflow.
// func (self *Assembler) MUL(a, b, c vm.DataAddress) {
// 	loop := self.make_label()
// 	exit_loop := self.make_label()

// 	self.MOV(vm.ZERO, c)
// 	self.label(loop)
// 	self.BEQ(a, vm.ZERO, exit_loop)
// 	self.ADD(b, c, c)
// 	self.DEC(a)
// 	self.JMP(loop)
// 	self.label(exit_loop)
// }

// ------------------------------------------------- stack management
//
// stack management is a bit tricky

func (self *Assembler) PUSH(a labeler) {
	data := self.uniqLabel()
	exit := self.uniqLabel()
	self.SBNZ(a, ZERO, Label("__push_operand"), self.ip+8)
	self.SBNZ(data, ZERO, Label("__push_ret"), self.ip+8)
	self.SBNZ(ONE, ZERO, JUNK, Label("__push"))
	self.SBNZ(ONE, ZERO, JUNK, exit)
	self.Label(data)
	self.DD(uint16(self.ip - 8))
	self.Label(exit)
}

func (self *Assembler) POP(a labeler) {
	data := self.uniqLabel()
	exit := self.uniqLabel()
	self.SBNZ(data, ZERO, Label("__pop_ret"), self.ip+8)
	self.SBNZ(ONE, ZERO, JUNK, Label("__pop"))
	self.SBNZ(Label("__push_operand"), ZERO, a, self.ip+8)
	self.SBNZ(ONE, ZERO, JUNK, exit)
	self.Label(data)
	self.DD(uint16(self.ip - 16))
	self.Label(exit)
}

// -------------------------------------------------- logical opcodes

// NOT perform the bitwise not on the contents of 'a' and stores the
// result in 'b'.
func (self *Assembler) NOT(a, b labeler) {
	self.ADD(a, ONE, b)
	self.NEG(b, b)
}

// // --------------------------------------------- emulating other OISC

// // SUBLEQ Subtract and branch if less than or equal to zero OISC.
// // Substrat content of address 'a' from content of 'b' and stores the
// // result en address 'c'. If the result is less than or equal to zero
// // jump to address 'c'.
// func (self *Assembler) SUBLEQ(a, b vm.DataAddress, c Address) {
// 	// not sure how to test 'value < 0'
// }
