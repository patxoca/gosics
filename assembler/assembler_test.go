package assembler

import (
	"gosics/vm"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions. Start with a 't_' prefix in order to avoid name
// collisions.

// t_createComputerAndRun create a new Computer, load the programa
// assembled by a and execute n steps
func t_createComputerAndRun(a *Assembler, n int) vm.Computer {
	c := vm.Computer{}
	c.LoadMemory(a.Assemble())
	c.Step() // jump to '__start'
	for ; n > 0; n-- {
		c.Step()
	}
	return c
}

// t_resolve ...
func t_resolve(a *Assembler, l string) vm.Address {
	return vm.Address(Label(l).getAddress(a))
}

// t_peek return a 2 bytes value from the memory of c correspondig to
// the label l. a is used to resolve the label into an address.
func t_peek(c *vm.Computer, a *Assembler, l string) vm.Operand {
	return c.Peek(t_resolve(a, l))
}

// Actual tests

func TestGetAddressForUnresolvableLabel(t *testing.T) {
	as := New()
	as.ip = 1234
	l := Label("foo")
	_, ok := as.unresolved[l]
	assert.False(t, ok)
	ad := l.getAddress(&as)
	assert.Equal(t, ad, maxAddress)
	ips, ok := as.unresolved[l]
	assert.True(t, ok)
	assert.Equal(t, 1, ips.Len())
	assert.Equal(t, Address(1234), ips.Front().Value.(Address))
}

func TestGetAddressForResolvableLabel(t *testing.T) {
	as := New()
	as.ip = 1234
	l := Label("foo")
	as.Label(l)
	ad := l.getAddress(&as)
	assert.Equal(t, ad, Address(1234))
	_, ok := as.unresolved[l]
	assert.False(t, ok)
}

func TestUniqLabelCreatesDistinctLabels(t *testing.T) {
	as := New()
	assert.NotEqual(t, as.uniqLabel(), as.uniqLabel())
}

func TestLabelCreatesLabelForCurrentIP(t *testing.T) {
	as := New()
	as.ip = 1234
	lab := as.uniqLabel()
	as.Label(lab)
	assert.Equal(t, Address(1234), as.labels[lab])
}

// test macro instructions

func TestHLT(t *testing.T) {
	as := New()
	as.HLT()

	c := t_createComputerAndRun(&as, 0)
	assert.False(t, c.Halted())
	c.Step()
	assert.True(t, c.Halted())
}

func TestMOV(t *testing.T) {
	as := New()
	as.MOV(Label("SRC"), Label("DST"))
	as.Label("SRC")
	as.DD(0x1234)
	as.Label("DST")
	as.DD(0x0000)

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, vm.Operand(0x1234), t_peek(&c, &as, "SRC"))
	assert.Equal(t, vm.Operand(0x1234), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "SRC"), c.IP())
}

func TestJMP(t *testing.T) {
	as := New()
	as.JMP(Label("DST"))
	as.Label("SRC")
	as.DD(0x1234)
	as.Label("DST")

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, t_resolve(&as, "DST"), c.IP())
}

func TestBEQBranch(t *testing.T) {
	as := New()
	as.BEQ(Label("OP1"), Label("OP2"), Label("DST"))
	as.Label("OP1")
	as.DD(0x1234)
	as.Label("OP2")
	as.DD(0x1234)
	as.Label("DST")

	c := t_createComputerAndRun(&as, 2)
	assert.Equal(t, t_resolve(&as, "DST"), c.IP())
}

func TestBEQNoBranch(t *testing.T) {
	as := New()
	as.BEQ(Label("OP1"), Label("OP2"), Label("DST"))
	as.Label("OP1")
	as.DD(0x1234)
	as.Label("OP2")
	as.DD(0x0000)
	as.Label("DST")

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, t_resolve(&as, "OP1"), c.IP())
}

func TestNEG(t *testing.T) {
	as := New()
	as.NEG(Label("SRC"), Label("DST"))
	as.Label("SRC")
	as.DD(0x1234)
	as.Label("DST")
	as.DD(0x0000)

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, vm.Operand(-0x1234), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "SRC"), c.IP())
}

func TestNEGZero(t *testing.T) {
	as := New()
	as.NEG(Label("SRC"), Label("DST"))
	as.Label("SRC")
	as.DD(0x0000)
	as.Label("DST")
	as.DD(0x1234)

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, vm.Operand(0x0000), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "SRC"), c.IP())
}

func TestADD(t *testing.T) {
	as := New()
	as.ADD(Label("OP1"), Label("OP2"), Label("DST"))
	as.Label("OP1")
	as.DD(0x1234)
	as.Label("OP2")
	as.DD(0x2345)
	as.Label("DST")
	as.DD(0x0000)

	c := t_createComputerAndRun(&as, 2)
	assert.Equal(t, vm.Operand(0x3579), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "OP1"), c.IP())
}

func TestADDZero(t *testing.T) {
	as := New()
	as.ADD(Label("OP1"), Label("OP2"), Label("DST"))
	as.Label("OP1")
	as.DD(0x1234)
	as.Label("OP2")
	as.DD(0xFFFF - 0x1234 + 1) // 2's complement == -0x1234
	as.Label("DST")
	as.DD(0xFFFF)

	c := t_createComputerAndRun(&as, 2)
	assert.Equal(t, vm.Operand(0x0000), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "OP1"), c.IP())
}

func TestSUB(t *testing.T) {
	as := New()
	as.SUB(Label("OP1"), Label("OP2"), Label("DST"))
	as.Label("OP1")
	as.DD(0x2345)
	as.Label("OP2")
	as.DD(0x1234)
	as.Label("DST")
	as.DD(0x0000)

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, vm.Operand(0x1111), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "OP1"), c.IP())
}

func TestSUBZero(t *testing.T) {
	as := New()
	as.SUB(Label("OP1"), Label("OP2"), Label("DST"))
	as.Label("OP1")
	as.DD(0x1234)
	as.Label("OP2")
	as.DD(0x1234)
	as.Label("DST")
	as.DD(0xFFFF)

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, vm.Operand(0x0000), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "OP1"), c.IP())
}

func TestINC(t *testing.T) {
	as := New()
	as.INC(Label("OP"))
	as.Label("OP")
	as.DD(0x1234)

	c := t_createComputerAndRun(&as, 2)
	assert.Equal(t, vm.Operand(0x1235), t_peek(&c, &as, "OP"))
	assert.Equal(t, t_resolve(&as, "OP"), c.IP())
}

func TestINCZero(t *testing.T) {
	as := New()
	as.INC(Label("OP"))
	as.Label("OP")
	as.DD(0xFFFF) // -1

	c := t_createComputerAndRun(&as, 2)
	assert.Equal(t, vm.Operand(0x0000), t_peek(&c, &as, "OP"))
	assert.Equal(t, t_resolve(&as, "OP"), c.IP())
}

func TestDEC(t *testing.T) {
	as := New()
	as.DEC(Label("OP"))
	as.Label("OP")
	as.DD(0x1234)

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, vm.Operand(0x1233), t_peek(&c, &as, "OP"))
	assert.Equal(t, t_resolve(&as, "OP"), c.IP())
}

func TestDECZero(t *testing.T) {
	as := New()
	as.DEC(Label("OP"))
	as.Label("OP")
	as.DD(0x0001)

	c := t_createComputerAndRun(&as, 1)
	assert.Equal(t, vm.Operand(0x0000), t_peek(&c, &as, "OP"))
	assert.Equal(t, t_resolve(&as, "OP"), c.IP())
}

func TestNOT(t *testing.T) {
	as := New()
	as.NOT(Label("SRC"), Label("DST"))
	as.Label("SRC")
	as.DD(0x1234)
	as.Label("DST")
	as.DD(0x0000)

	c := t_createComputerAndRun(&as, 3)
	assert.Equal(t, ^vm.Operand(0x1234), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "SRC"), c.IP())
}

func TestNOTZero(t *testing.T) {
	as := New()
	as.NOT(Label("SRC"), Label("DST"))
	as.Label("SRC")
	as.DD(0xFFFF)
	as.Label("DST")
	as.DD(0xFFFF)

	c := t_createComputerAndRun(&as, 3)
	assert.Equal(t, vm.Operand(0), t_peek(&c, &as, "DST"))
	assert.Equal(t, t_resolve(&as, "SRC"), c.IP())
}

func TestPUSH(t *testing.T) {
	as := New()
	as.PUSH(Label("SRC"))
	as.Label("SRC")
	as.DD(0x1234)

	c := t_createComputerAndRun(&as, 9)
	assert.Equal(t, vm.Operand(0x1234), c.Peek(vm.MaxAddress-1))
	assert.Equal(t, vm.Operand(-4), t_peek(&c, &as, "__SP"))
	assert.Equal(t, t_resolve(&as, "SRC"), c.IP())
}

func TestPOP(t *testing.T) {
	as := New()
	as.PUSH(Label("SRC"))
	as.POP(Label("DST"))
	as.Label("SRC")
	as.DD(0x1234)
	as.Label("DST")
	as.DD(0x0000)

	c := t_createComputerAndRun(&as, 9+10)
	assert.Equal(t, vm.Operand(0x1234), t_peek(&c, &as, "DST"))
	assert.Equal(t, vm.Operand(-2), t_peek(&c, &as, "__SP"))
	assert.Equal(t, t_resolve(&as, "SRC"), c.IP())
}
