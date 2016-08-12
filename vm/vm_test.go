package vm

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVMStartsAtIPZero(t *testing.T) {
	c := Computer{}
	assert.Equal(t, c.ip, Address(0))
}

func TestLoadMemory(t *testing.T) {
	memory := make([]uint8, MemorySize)
	copy := make([]uint8, MemorySize)
	for i := uint(0); i < MemorySize; i++ {
		value := rand.Int31n(256)
		memory[i] = uint8(value)
		copy[i] = uint8(value)
	}
	c := Computer{}
	c.LoadMemory(memory)
	for i := uint(0); i < MemorySize; i++ {
		assert.Equal(t, copy[i], c.memory[i])
	}
}

func TestFetchAddress(t *testing.T) {
	c := Computer{}
	c.LoadMemory([]uint8{0xFA, 0xBA, 0xDA, 0xFF})
	assert.Equal(t, c.fetchAddress(0), Address(0xFABA))
	assert.Equal(t, c.fetchAddress(1), Address(0xBADA))
	assert.Equal(t, c.fetchAddress(2), Address(0xDAFF))
}

func TestFetchOperand(t *testing.T) {
	c := Computer{}
	c.LoadMemory([]uint8{0xFA, 0xBA, 0x0A, 0xFF})
	assert.Equal(t, c.fetchOperand(0), Operand(-1350))
	assert.Equal(t, c.fetchOperand(2), Operand(2815))
}

func TestPutOperand(t *testing.T) {
	c := Computer{}
	c.LoadMemory([]uint8{0x00, 0x00, 0x00, 0x00})
	c.putOperand(0, Operand(0x0102))
	c.putOperand(2, Operand(0x0304))
	assert.Equal(t, c.memory[:4], []uint8{0x01, 0x02, 0x03, 0x04})
}

func TestStep(t *testing.T) {
	data := []struct {
		memory []uint8
		eip    Address // expected IP
		emem   []uint8 // expected contents memory
	}{
		{ // branches
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x0C, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x02, // *b
				0x00, 0x00, // *c
			},
			0xFA,
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x0C, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x02, // *b
				0x00, 0x03, // *r,
			},
		},
		{ // not branches
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x0C, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x05, // *b
			},
			0x08,
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x0C, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x05, // *b
				0x00, 0x00, // *c,
			},
		},
		{ // A == B == C
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x08, // b
				0x00, 0x08, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
			},
			0x08,
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x08, // b
				0x00, 0x08, // c
				0x00, 0xFA, // d
				0x00, 0x00, // *c
			},
		},
		{ // A == B
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x08, // b == a
				0x00, 0x0A, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x01, // *c
			},
			0x08,
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x08, // b
				0x00, 0x0A, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x00, // *c
			},
		},
		{ // A == C
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x08, // c == a
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x02, // *b
			},
			0xFA,
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x08, // c
				0x00, 0xFA, // d
				0x00, 0x03, // *a
				0x00, 0x02, // *c
			},
		},
		{ // B == C
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x0A, // c == b
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x02, // *b
			},
			0xFA,
			[]uint8{
				0x00, 0x08, // a
				0x00, 0x0A, // b
				0x00, 0x0A, // c
				0x00, 0xFA, // d
				0x00, 0x05, // *a
				0x00, 0x03, // *c
			},
		},
	}
	for i, d := range data {
		t.Logf("Iteration %d, %#v", i, d)
		c := Computer{}
		c.LoadMemory(d.memory)

		c.Step()

		assert.Equal(t, d.eip, c.ip, "IP mismatch")
		for j, v := range d.emem {
			assert.Equal(t, v, c.memory[j], "Memory mismatch")
		}
	}
}

func TestHalt(t *testing.T) {
	memory := []uint8{
		0x00, 0x08, // a
		0x00, 0x0A, // b
		0x00, 0x0C, // c
		0xFF, 0xFF, // d
		0x00, 0x05, // *a
		0x00, 0x02, // *b
		0x00, 0x00, // *c
	}
	c := Computer{}
	c.LoadMemory(memory)
	c.Step()

	assert.Equal(t, MaxAddress, c.ip)
	assert.True(t, c.Halted())
}
