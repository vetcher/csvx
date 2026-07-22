package csvx

import (
	"io"
	"strings"
	"testing"
)

func TestDecoder_OneRowThenEOF(t *testing.T) {
	r := strings.NewReader("client_id,client_name,client_age\n1,Jose,42\n")
	d := NewDecoder(r)
	var c Client
	if err := d.Decode(&c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "Jose" {
		t.Fatalf("%+v", c)
	}
	if err := d.Decode(&c); err != io.EOF {
		t.Fatalf("got %v want EOF", err)
	}
}

func TestDecoder_SliceTarget(t *testing.T) {
	r := strings.NewReader("client_id,client_name,client_age\n1,Jose,42\n2,Daniel,26\n")
	d := NewDecoder(r)
	var all []Client
	if err := d.Decode(&all); err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("%d", len(all))
	}
	if err := d.Decode(&all); err != io.EOF {
		t.Fatalf("got %v want EOF", err)
	}
}

func TestRows(t *testing.T) {
	r := strings.NewReader("client_id,client_name,client_age\n1,Jose,42\n")
	n := 0
	for row, err := range Rows[Client](r) {
		if err != nil {
			t.Fatal(err)
		}
		if row.Name != "Jose" {
			t.Fatalf("%+v", row)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestEncoder_WritesHeaderOnce(t *testing.T) {
	var buf strings.Builder
	e := NewEncoder(&buf)
	if err := e.Encode(Client{ID: "1", Name: "A", Age: 1}); err != nil {
		t.Fatal(err)
	}
	if err := e.Encode(Client{ID: "2", Name: "B", Age: 2}); err != nil {
		t.Fatal(err)
	}
	want := "client_id,client_name,client_age\n1,A,1\n2,B,2\n"
	if buf.String() != want {
		t.Fatalf("got %q", buf.String())
	}
}
