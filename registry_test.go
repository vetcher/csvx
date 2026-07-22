package csvx

import (
	"reflect"
	"testing"
)

func TestRegisterMarshaler_RoundTripViaMarshalCell(t *testing.T) {
	ClearRegistry()
	type tag string
	RegisterMarshaler(func(v tag) (string, error) { return string(v), nil })
	t.Cleanup(ClearRegistry)
	out, err := marshalCell(reflect.ValueOf(tag("hi")), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if out != "hi" {
		t.Fatalf("got %q", out)
	}
}

func TestRegisterUnmarshaler_RoundTripViaUnmarshalCell(t *testing.T) {
	ClearRegistry()
	type tag string
	RegisterUnmarshaler(func(s string) (tag, error) { return tag(s), nil })
	t.Cleanup(ClearRegistry)
	dst := reflect.ValueOf(new(tag)).Elem()
	if err := unmarshalCell(dst, "bye", joinOptions(nil)); err != nil {
		t.Fatal(err)
	}
	if dst.String() != "bye" {
		t.Fatalf("got %q", dst.String())
	}
}

func TestClearRegistry_RemovesGlobalHooks(t *testing.T) {
	ClearRegistry()
	type lone int
	RegisterMarshaler(func(v lone) (string, error) { return "x", nil })
	ClearRegistry()
	out, err := marshalCell(reflect.ValueOf(lone(1)), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if out != "1" {
		t.Fatalf("after clear expected builtin %q, got %q", "1", out)
	}
}

func TestJoinMarshalers_LastEntryWinsForType(t *testing.T) {
	type t1 struct{ N int }
	m1 := MarshalFunc(func(v t1) (string, error) { return "a", nil })
	m2 := MarshalFunc(func(v t1) (string, error) { return "b", nil })
	joined := JoinMarshalers(m1, m2)
	out, err := marshalCell(reflect.ValueOf(t1{}), joinOptions([]Options{WithMarshalers(joined)}))
	if err != nil {
		t.Fatal(err)
	}
	if out != "b" {
		t.Fatalf("got %q want b (last wins)", out)
	}
}

func TestJoinUnmarshalers_LastEntryWinsForType(t *testing.T) {
	type t1 struct{ N int }
	u1 := UnmarshalFunc(func(s string) (t1, error) { return t1{N: 1}, nil })
	u2 := UnmarshalFunc(func(s string) (t1, error) { return t1{N: 2}, nil })
	joined := JoinUnmarshalers(u1, u2)
	dst := reflect.ValueOf(new(t1)).Elem()
	if err := unmarshalCell(dst, "x", joinOptions([]Options{WithUnmarshalers(joined)})); err != nil {
		t.Fatal(err)
	}
	if dst.Interface().(t1).N != 2 {
		t.Fatalf("got N=%d want 2", dst.Interface().(t1).N)
	}
}
