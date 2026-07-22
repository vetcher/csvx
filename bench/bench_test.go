package bench_test

import (
	"bytes"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/vetcher/csvx"
)

const fixtureRows = 10_000

// benchRow has ten columns for ~10k × ~10 fixture.
type benchRow struct {
	C0 string `csv:"col_0"`
	C1 string `csv:"col_1"`
	C2 string `csv:"col_2"`
	C3 string `csv:"col_3"`
	C4 string `csv:"col_4"`
	C5 string `csv:"col_5"`
	C6 string `csv:"col_6"`
	C7 string `csv:"col_7"`
	C8 string `csv:"col_8"`
	C9 string `csv:"col_9"`
}

var (
	csvFixture []byte
	structSlice []benchRow
)

func init() {
	var b strings.Builder
	w := csv.NewWriter(&b)
	header := make([]string, 10)
	for i := range header {
		header[i] = "col_" + strconv.Itoa(i)
	}
	_ = w.Write(header)

	structSlice = make([]benchRow, fixtureRows)
	for r := 0; r < fixtureRows; r++ {
		row := make([]string, 10)
		for c := 0; c < 10; c++ {
			val := strconv.Itoa(r*10 + c)
			row[c] = val
			switch c {
			case 0:
				structSlice[r].C0 = val
			case 1:
				structSlice[r].C1 = val
			case 2:
				structSlice[r].C2 = val
			case 3:
				structSlice[r].C3 = val
			case 4:
				structSlice[r].C4 = val
			case 5:
				structSlice[r].C5 = val
			case 6:
				structSlice[r].C6 = val
			case 7:
				structSlice[r].C7 = val
			case 8:
				structSlice[r].C8 = val
			case 9:
				structSlice[r].C9 = val
			}
		}
		_ = w.Write(row)
	}
	w.Flush()
	csvFixture = []byte(b.String())
}

func BenchmarkEncodingCSV_Parse(b *testing.B) {
	b.SetBytes(int64(len(csvFixture)))
	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewReader(csvFixture))
		for {
			_, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkCSVx_UnmarshalStruct(b *testing.B) {
	b.SetBytes(int64(len(csvFixture)))
	for i := 0; i < b.N; i++ {
		var out []benchRow
		if err := csvx.Unmarshal(csvFixture, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCSVx_MarshalStruct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := csvx.Marshal(structSlice); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCSVx_DecodeAllocs(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(csvFixture)))
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(csvFixture)
		d := csvx.NewDecoder(r)
		var row benchRow
		for {
			err := d.Decode(&row)
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
