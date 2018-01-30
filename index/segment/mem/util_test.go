package mem

import "testing"

var (
	testBytes []byte
	emptyStr  string
)

func BenchmarkEmptyStringToByteSlice(b *testing.B) {
	// Converting an empty string to byte slice does not allocate.
	for n := 0; n < b.N; n++ {
		testBytes = []byte(emptyStr)
	}
}
