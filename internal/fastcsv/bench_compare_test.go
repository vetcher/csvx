package fastcsv

import (
	"bytes"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"testing"
)

const benchRows = 10_000

var benchCSV []byte

func init() {
	var b strings.Builder
	w := csv.NewWriter(&b)
	header := make([]string, 10)
	for i := range header {
		header[i] = "col_" + strconv.Itoa(i)
	}
	_ = w.Write(header)
	for r := 0; r < benchRows; r++ {
		row := make([]string, 10)
		for c := 0; c < 10; c++ {
			row[c] = strconv.Itoa(r*10 + c)
		}
		_ = w.Write(row)
	}
	w.Flush()
	benchCSV = []byte(b.String())
}

func BenchmarkParse_FastCSV(b *testing.B) {
	b.SetBytes(int64(len(benchCSV)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := NewReader(bytes.NewReader(benchCSV))
		n := 0
		for {
			_, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
			n++
		}
		if n != benchRows+1 {
			b.Fatalf("rows: got %d want %d", n, benchRows+1)
		}
	}
}

func BenchmarkParse_EncodingCSV(b *testing.B) {
	b.SetBytes(int64(len(benchCSV)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cr := csv.NewReader(bytes.NewReader(benchCSV))
		cr.ReuseRecord = true
		n := 0
		for {
			_, err := cr.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
			n++
		}
		if n != benchRows+1 {
			b.Fatalf("rows: got %d want %d", n, benchRows+1)
		}
	}
}
