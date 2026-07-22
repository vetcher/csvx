package csvx

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarshal_StructSliceHeader(t *testing.T) {
	in := []Client{{ID: "1", Name: "Jose", Age: 42}}
	b, err := Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	want := "client_id,client_name,client_age\n1,Jose,42\n"
	if string(b) != want {
		t.Fatalf("got %q", b)
	}
}

func TestMarshal_OmitEmpty(t *testing.T) {
	type Row struct {
		A string `csv:"a"`
		B string `csv:"b,omitempty"`
	}
	b, err := Marshal([]Row{{A: "x"}})
	if err != nil {
		t.Fatal(err)
	}
	// Header still lists b; empty cell for b. CSV cannot drop columns mid-row.
	if string(b) != "a,b\nx,\n" {
		t.Fatalf("got %q", b)
	}
}

func TestMarshal_OmitZero(t *testing.T) {
	type Row struct {
		A string `csv:"a"`
		N int    `csv:"n,omitzero"`
	}
	b, err := Marshal([]Row{{A: "x", N: 0}})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "a,n\nx,\n" {
		t.Fatalf("got %q", b)
	}
}

func TestMarshal_StructPointerSlice(t *testing.T) {
	in := []*Client{{ID: "1", Name: "Jose", Age: 42}}
	b, err := Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	want := "client_id,client_name,client_age\n1,Jose,42\n"
	if string(b) != want {
		t.Fatalf("got %q", b)
	}
}

func TestMarshalWrite(t *testing.T) {
	var buf bytes.Buffer
	if err := MarshalWrite(&buf, []Client{{ID: "1", Name: "Jose", Age: 42}}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "client_id,client_name,client_age\n1,Jose,42\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestMarshal_RejectNonSlice(t *testing.T) {
	_, err := Marshal(Client{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMarshal_NoHeader(t *testing.T) {
	in := []Client{{ID: "1", Name: "Jose", Age: 42}}
	b, err := Marshal(in, NoHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "client_id") {
		t.Fatalf("header should be omitted: %q", b)
	}
	if string(b) != "1,Jose,42\n" {
		t.Fatalf("got %q", b)
	}
}
