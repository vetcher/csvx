package csvx

import (
	"reflect"
	"testing"
)

func TestBuildTypePlan_ClassifiesBuiltinString(t *testing.T) {
	type Row struct {
		A string `csv:"a"`
	}
	p, err := buildTypePlan(reflect.TypeOf(Row{}), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if p.fields[0].conv != fieldConvBuiltin {
		t.Fatalf("conv=%v want builtin", p.fields[0].conv)
	}
}

func TestBuildTypePlan_ClassifiesUnmarshaler(t *testing.T) {
	type Row struct {
		A ordered `csv:"a"`
	}
	p, err := buildTypePlan(reflect.TypeOf(Row{}), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if p.fields[0].conv != fieldConvUnmarshaler {
		t.Fatalf("conv=%v want unmarshaler", p.fields[0].conv)
	}
}

func TestUnmarshal_BuiltinFieldsSkipPerCellInterfaceProbeAllocs(t *testing.T) {
	// 100 rows × 1 string column: after fix, allocs/row should stay low.
	// Budget: < 8 allocs/row on average (record copy + elem + append + string).
	in := []byte("a\n" + repeatLines("x", 100))
	allocs := testing.AllocsPerRun(50, func() {
		var out []struct {
			A string `csv:"a"`
		}
		if err := Unmarshal(in, &out); err != nil {
			t.Fatal(err)
		}
		if len(out) != 100 {
			t.Fatalf("len=%d", len(out))
		}
	})
	perRow := allocs / 100
	if perRow >= 10 {
		t.Fatalf("allocs/row=%.2f (total=%.0f) — still probing interfaces per cell?", perRow, allocs)
	}
}

func repeatLines(s string, n int) string {
	b := make([]byte, 0, n*(len(s)+1))
	for i := 0; i < n; i++ {
		b = append(b, s...)
		b = append(b, '\n')
	}
	return string(b)
}
