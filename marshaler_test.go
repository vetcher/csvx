package csvx

import (
	"encoding"
	"reflect"
	"testing"
)

type ordered struct{ S string }

func (o ordered) MarshalCSV() (string, error) { return "iface:" + o.S, nil }

type orderedUnmarshal struct{ S string }

func (o *orderedUnmarshal) UnmarshalCSV(s string) error {
	o.S = "iface:" + s
	return nil
}

func TestMarshalCell_LocalBeatsInterface(t *testing.T) {
	ClearRegistry()
	local := MarshalFunc(func(o ordered) (string, error) { return "local:" + o.S, nil })
	out, err := marshalCell(reflect.ValueOf(ordered{S: "x"}), joinOptions([]Options{WithMarshalers(local)}))
	if err != nil {
		t.Fatal(err)
	}
	if out != "local:x" {
		t.Fatalf("got %q", out)
	}
}

func TestMarshalCell_InterfaceBeatsGlobal(t *testing.T) {
	ClearRegistry()
	RegisterMarshaler(func(o ordered) (string, error) { return "global:" + o.S, nil })
	t.Cleanup(ClearRegistry)
	out, err := marshalCell(reflect.ValueOf(ordered{S: "x"}), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if out != "iface:x" {
		t.Fatalf("got %q want iface", out)
	}
}

func TestMarshalCell_GlobalBeatsBuiltin(t *testing.T) {
	ClearRegistry()
	type myInt int
	RegisterMarshaler(func(v myInt) (string, error) { return "g", nil })
	t.Cleanup(ClearRegistry)
	out, err := marshalCell(reflect.ValueOf(myInt(3)), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if out != "g" {
		t.Fatalf("got %q", out)
	}
}

type textMarshalInt int

func (v textMarshalInt) MarshalText() ([]byte, error) { return []byte("text"), nil }

func TestMarshalCell_TextBeatsBuiltin(t *testing.T) {
	ClearRegistry()
	out, err := marshalCell(reflect.ValueOf(textMarshalInt(7)), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if out != "text" {
		t.Fatalf("got %q want text", out)
	}
}

func TestUnmarshalCell_LocalBeatsInterface(t *testing.T) {
	ClearRegistry()
	local := UnmarshalFunc(func(s string) (orderedUnmarshal, error) {
		return orderedUnmarshal{S: "local:" + s}, nil
	})
	dst := reflect.ValueOf(&orderedUnmarshal{})
	err := unmarshalCell(dst.Elem(), "x", joinOptions([]Options{WithUnmarshalers(local)}))
	if err != nil {
		t.Fatal(err)
	}
	if dst.Elem().Interface().(orderedUnmarshal).S != "local:x" {
		t.Fatalf("got %q", dst.Elem().Interface().(orderedUnmarshal).S)
	}
}

func TestUnmarshalCell_InterfaceBeatsGlobal(t *testing.T) {
	ClearRegistry()
	RegisterUnmarshaler(func(s string) (orderedUnmarshal, error) {
		return orderedUnmarshal{S: "global:" + s}, nil
	})
	t.Cleanup(ClearRegistry)
	dst := reflect.ValueOf(&orderedUnmarshal{})
	err := unmarshalCell(dst.Elem(), "x", joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if dst.Elem().Interface().(orderedUnmarshal).S != "iface:x" {
		t.Fatalf("got %q want iface:x", dst.Elem().Interface().(orderedUnmarshal).S)
	}
}

func TestUnmarshalCell_GlobalBeatsBuiltin(t *testing.T) {
	ClearRegistry()
	type myInt int
	RegisterUnmarshaler(func(s string) (myInt, error) { return myInt(99), nil })
	t.Cleanup(ClearRegistry)
	dst := reflect.ValueOf(new(myInt)).Elem()
	err := unmarshalCell(dst, "3", joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if dst.Interface().(myInt) != 99 {
		t.Fatalf("got %v", dst.Interface())
	}
}

type textUnmarshalInt int

func (v *textUnmarshalInt) UnmarshalText(text []byte) error {
	*v = 42
	return nil
}

func TestUnmarshalCell_TextBeatsBuiltin(t *testing.T) {
	ClearRegistry()
	dst := reflect.ValueOf(new(textUnmarshalInt)).Elem()
	err := unmarshalCell(dst, "7", joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if dst.Interface().(textUnmarshalInt) != 42 {
		t.Fatalf("got %v", dst.Interface())
	}
}

// Ensure encoding imports are used for interface satisfaction checks.
var (
	_ encoding.TextMarshaler   = textMarshalInt(0)
	_ encoding.TextUnmarshaler = (*textUnmarshalInt)(nil)
)
