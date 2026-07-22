package csvx

import (
	"errors"
	"strings"
	"testing"
)

func TestStdlibReader_ReadsRecords(t *testing.T) {
	in := "a,b\n1,2\n"
	r := newStdlibReader(strings.NewReader(in), joinOptions(nil))
	rec, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec) != 2 || rec[0] != "a" || rec[1] != "b" {
		t.Fatalf("got %#v", rec)
	}
	rec, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if rec[0] != "1" || rec[1] != "2" {
		t.Fatalf("got %#v", rec)
	}
}

func TestStdlibReader_ParseErrorIsSyntaxError(t *testing.T) {
	in := "\"unclosed\n"
	r := newStdlibReader(strings.NewReader(in), joinOptions(nil))
	_, err := r.Read()
	if err == nil {
		t.Fatal("expected parse error")
	}
	var se *SyntaxError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SyntaxError, got %T: %v", err, err)
	}
}

func TestStdlibWriter_WritesCommaOverride(t *testing.T) {
	var buf strings.Builder
	w := newStdlibWriter(&buf, joinOptions([]Options{Comma(';')}))
	if err := w.Write([]string{"x", "y"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "x;y\n" {
		t.Fatalf("got %q", buf.String())
	}
}
