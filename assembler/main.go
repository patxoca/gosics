// This package implements an in memory assembler:
//
// - supports labels to reference memory addresses
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

const _ONE Address = maxAddress - 1
const _ZERO Address = maxAddress - 3
const _JUNK Address = maxAddress - 5

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
	getAddress(a Assembler) Address
}

// getAddress return the address for a literal address
func (self Address) getAddress(a Assembler) Address {
	return self
}

// getAddress perform a lookup for the label in the label table and
// return the corresponding address. If the label is not found adds it
// to the unresolved-labels table and return a fake address.
func (self Label) getAddress(a Assembler) Address {
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
		addr := v.getAddress(*self)
		self.memory[self.ip] = uint8(addr >> 8)
		self.memory[self.ip+1] = uint8(addr & 0xFF)
		self.ip += 2
	}
}
