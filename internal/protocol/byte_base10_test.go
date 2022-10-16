package protocol

import (
	"github.com/nsqio/nsq/internal/test"
	"testing"
)

var result uint64

func BenchmarkByteToBase10Valid(b *testing.B) {
	bt := []byte{'3', '1', '4', '1', '5', '9', '2', '5'}
	var n uint64
	for i := 0; i < b.N; i++ {
		n, _ = ByteToBase10(bt)
	}
	result = n
}

func BenchmarkByteToBase10Invalid(b *testing.B) {
	bt := []byte{'?', '1', '4', '1', '5', '9', '2', '5'}
	var n uint64
	for i := 0; i < b.N; i++ {
		n, _ = ByteToBase10(bt)
	}
	result = n
}

func TestByteToBase10(t *testing.T) {
	bt := []byte{'2', '6'}
	base10, err := ByteToBase10(bt)
	test.Equal(t, nil, err)
	test.Equal(t, uint64(26), base10)
}
