package vm

import "fmt"

////////////////////////////////////////////////////////////////////////
//
// the virtual machine
//
// - Unified address spaces for program and data.
//
// - big endian
//
// - Pointers are 16 bits long
//
// - Operands are 16 bit long, signed

type Address uint16
type Operand int16

const MaxAddress = Address(^Address(0))
const MemorySize = uint(MaxAddress) + 1
const HALT Address = MaxAddress
const bytesPerAddress = 2
const bytesPerOperand = 2

type Computer struct {
	ip     Address
	memory [MemorySize]uint8
}

// LoadMemory loads the memory image into memory
func (self *Computer) LoadMemory(data []uint8) {
	for i, c := range data {
		self.memory[i] = c
	}
}

// Halted return true if the computer is halted
func (self *Computer) Halted() bool {
	return (self.ip == HALT)
}

func (self *Computer) fetchAddress(p Address) Address {
	res := Address(0)
	for i := 0; i < bytesPerAddress; i++ {
		res = (res << 8) | Address(self.memory[int(p)+i])
	}
	return res
}

func (self *Computer) fetchOperand(p Address) Operand {
	res := Operand(0)
	for i := 0; i < bytesPerOperand; i++ {
		res = (res << 8) | Operand(self.memory[int(p)+i])
	}
	return res
}

func (self *Computer) putOperand(p Address, o Operand) {
	for i := bytesPerOperand - 1; i >= 0; i-- {
		self.memory[int(p)+i] = uint8(o & Operand(0xFF))
		o = o >> 8
	}
}

// Step execute the next instruction and updates the IP pointer, if
// the computer is not halted
func (self *Computer) Step() {
	if !self.Halted() {
		a := self.fetchOperand(self.fetchAddress(self.ip))
		b := self.fetchOperand(self.fetchAddress(self.ip + bytesPerAddress))
		r := a - b
		self.putOperand(self.fetchAddress(self.ip+2*bytesPerAddress), r)
		if r != 0 {
			self.ip = self.fetchAddress(self.ip + 3*bytesPerAddress)
		} else {
			self.ip += 4 * bytesPerAddress
		}
	}
}

func (self *Computer) Print(n int) {
	fmt.Printf("%5d: ", self.ip)
	for _, v := range self.memory[:n] {
		fmt.Printf("%03d ", v)
	}
	fmt.Println("")
}
