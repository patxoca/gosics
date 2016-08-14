package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAddressForUnresolvableLabel(t *testing.T) {
	as := New()
	as.ip = 1234
	l := Label("foo")
	_, ok := as.unresolved[l]
	assert.False(t, ok)
	ad := l.getAddress(as)
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
	ad := l.getAddress(as)
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
	assert.Equal(t, 0, len(as.labels))
	lab := Label("test")
	as.Label(lab)
	assert.Equal(t, 1, len(as.labels))
	assert.Equal(t, Address(1234), as.labels[lab])
}
