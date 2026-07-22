//go:build !nogocsv

package bench_test

import (
	"testing"

	"github.com/gocarina/gocsv"
)

func BenchmarkGoCSV_UnmarshalStruct(b *testing.B) {
	b.SetBytes(int64(len(csvFixture)))
	for i := 0; i < b.N; i++ {
		var out []benchRow
		if err := gocsv.UnmarshalBytes(csvFixture, &out); err != nil {
			b.Fatal(err)
		}
	}
}
