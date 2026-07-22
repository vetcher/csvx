package fastcsv

import (
	"io"
	"strings"
	"testing"
)

func TestRead_simpleFields(t *testing.T) {
	r := NewReader(strings.NewReader("a,b,c\n1,2,3\n"))
	rec, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(rec, "|") != "a|b|c" {
		t.Fatalf("header: %#v", rec)
	}
	rec, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(rec, "|") != "1|2|3" {
		t.Fatalf("row: %#v", rec)
	}
	_, err = r.Read()
	if err != io.EOF {
		t.Fatalf("want EOF, got %v", err)
	}
}

func TestRead_quotedEmbeddedComma(t *testing.T) {
	in := `a,"b,c",d` + "\n"
	r := NewReader(strings.NewReader(in))
	rec, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a", "b,c", "d"}
	for i := range want {
		if rec[i] != want[i] {
			t.Fatalf("field %d: got %q want %q (rec=%#v)", i, rec[i], want[i], rec)
		}
	}
}

func TestRead_escapedQuotesRFC4180(t *testing.T) {
	in := `"say ""hello""",x` + "\n"
	r := NewReader(strings.NewReader(in))
	rec, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if rec[0] != `say "hello"` {
		t.Fatalf("got %q", rec[0])
	}
	if rec[1] != "x" {
		t.Fatalf("got %q", rec[1])
	}
}

func TestRead_CRLF(t *testing.T) {
	in := "a,b\r\nc,d\r\n"
	r := NewReader(strings.NewReader(in))
	rec, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(rec, ",") != "a,b" {
		t.Fatalf("first row: %#v", rec)
	}
	rec, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(rec, ",") != "c,d" {
		t.Fatalf("second row: %#v", rec)
	}
}

func TestRead_emptyFields(t *testing.T) {
	r := NewReader(strings.NewReader(",,\n"))
	rec, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec) != 3 || rec[0] != "" || rec[1] != "" || rec[2] != "" {
		t.Fatalf("got %#v", rec)
	}
}

func TestRead_quotedNewlineNotAllowed(t *testing.T) {
	// RFC4180 allows newlines in quoted fields; minimal reader must support it.
	in := "\"line1\nline2\",b\n"
	r := NewReader(strings.NewReader(in))
	rec, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if rec[0] != "line1\nline2" || rec[1] != "b" {
		t.Fatalf("got %#v", rec)
	}
}
