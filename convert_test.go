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

func TestFormatBuiltin_SliceByteAsString(t *testing.T) {
	s, err := formatBuiltin(reflect.ValueOf([]byte("hi")))
	if err != nil {
		t.Fatal(err)
	}
	if s != "hi" {
		t.Fatalf("got %q", s)
	}
}
