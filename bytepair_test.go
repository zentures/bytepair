package bytepair

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytePair(t *testing.T) {
	in := []byte("aaabdaaabac")
	out, dict := Encode(in)
	out2 := Decode(out, dict)
	require.Equal(t, in, out2)
}

func BenchmarkBytePairEncode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	data := []byte("aaabdaaabac")
	for i := 0; i < b.N; i++ {
		Encode(data)
	}
}
