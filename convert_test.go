package csvx

import (
	"reflect"
	"testing"
)

func TestParseBuiltin_IntAndBool(t *testing.T) {
	var i int64
	if err := parseBuiltin(reflect.ValueOf(&i).Elem(), "42", false); err != nil {
		t.Fatal(err)
	}
	if i != 42 {
		t.Fatalf("got %d", i)
	}
	var b bool
	if err := parseBuiltin(reflect.ValueOf(&b).Elem(), "1", false); err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("want true")
	}
}

func TestParseBuiltin_RejectsInterface(t *testing.T) {
	var v any
	err := parseBuiltin(reflect.ValueOf(&v).Elem(), "x", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseBuiltin_PointerEmptyNilDefault(t *testing.T) {
	var p *int
	rv := reflect.ValueOf(&p).Elem()
	if err := parseBuiltin(rv, "", false); err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Fatal("want nil pointer for empty cell")
	}
}

func TestParseBuiltin_AllocEmptyPointers(t *testing.T) {
	var p *int
	rv := reflect.ValueOf(&p).Elem()
	if err := parseBuiltin(rv, "", true); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 0 {
		t.Fatalf("got %#v", p)
	}
}

func TestParseBuiltin_AllocEmptyPointersExistingNonNil(t *testing.T) {
	seven := 7
	p := &seven
	rv := reflect.ValueOf(&p).Elem()
	if err := parseBuiltin(rv, "", true); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 0 {
		t.Fatalf("got %#v", p)
	}
}

func TestParseBuiltin_EmptyCellZeroValueNonPointer(t *testing.T) {
	cases := []struct {
		name string
		set  func() reflect.Value
	}{
		{"int", func() reflect.Value { var v int; return reflect.ValueOf(&v).Elem() }},
		{"bool", func() reflect.Value { var v bool; return reflect.ValueOf(&v).Elem() }},
		{"float64", func() reflect.Value { var v float64; return reflect.ValueOf(&v).Elem() }},
		{"string", func() reflect.Value { var v string; return reflect.ValueOf(&v).Elem() }},
		{"[]byte", func() reflect.Value { var v []byte; return reflect.ValueOf(&v).Elem() }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rv := tc.set()
			if err := parseBuiltin(rv, "", false); err != nil {
				t.Fatalf("empty cell: %v", err)
			}
			if !rv.IsZero() {
				t.Fatalf("want zero value, got %#v", rv.Interface())
			}
		})
	}
}

func TestParseBuiltin_EmptyCellPreservesNonZeroIntWhenNonEmpty(t *testing.T) {
	var i int = 99
	rv := reflect.ValueOf(&i).Elem()
	if err := parseBuiltin(rv, "5", false); err != nil {
		t.Fatal(err)
	}
	if i != 5 {
		t.Fatalf("got %d", i)
	}
}

func TestFormatBuiltin_SliceByteAsString(t *testing.T) {
	s, err := formatBuiltin(reflect.ValueOf([]byte("hi")))
	if err != nil {
		t.Fatal(err)
	}
	if s != "hi" {
		t.Fatalf("got %q", s)
	}
}
