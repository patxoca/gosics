package vm

import "testing"

// assertIP ...
func assertIP(t *testing.T, i, e ProgramAddress) {
	if i != e {
		t.Errorf("IP, expected %d, got %d", e, i)
	}
}

// assertData ...
func assertData(t *testing.T, d, e DataCell) {
	if d != e {
		t.Errorf("Data memory, expected %d, got %d", e, d)
	}
}

func TestVMStartsAtIPZero(t *testing.T) {
	c := Computer{}
	assertIP(t, c.ip, 0)
}

func TestStep(t *testing.T) {
	data := []struct {
		program []Instruction
		data    []DataCell
		eip     ProgramAddress // expected IP
		edata   [3]DataCell    // expected contents for a, b and c adresses
	}{
		{ // branches
			[]Instruction{
				{a: 0, b: 1, c: 2, d: 42},
			},
			[]DataCell{5, 2},
			42,
			[3]DataCell{5, 2, 3},
		},
		{ // don't branch
			[]Instruction{
				{a: 0, b: 1, c: 2, d: 42},
			},
			[]DataCell{5, 5},
			1,
			[3]DataCell{5, 5, 0},
		},
		{ // A == C
			[]Instruction{
				{a: 0, b: 1, c: 0, d: 42},
			},
			[]DataCell{5, 2, 0},
			42,
			[3]DataCell{3, 2, 0},
		},
		{ // B == C
			[]Instruction{
				{a: 0, b: 1, c: 1, d: 42},
			},
			[]DataCell{5, 2, 0},
			42,
			[3]DataCell{5, 3, 0},
		},
		{ // A == B == C
			[]Instruction{
				{a: 1, b: 1, c: 1, d: 42},
			},
			[]DataCell{0, 5, 0},
			1,
			[3]DataCell{0, 0, 0},
		},
	}
	for i, d := range data {
		t.Logf("Iteration %d, %#v", i, d)
		c := Computer{}
		c.LoadProgram(d.program)
		c.LoadData(d.data)

		c.Step()

		assertData(t, c.data_memory[0], d.edata[0])
		assertData(t, c.data_memory[1], d.edata[1])
		assertData(t, c.data_memory[2], d.edata[2])
		assertIP(t, c.ip, d.eip)
	}
}
