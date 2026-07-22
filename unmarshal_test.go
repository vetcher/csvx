package csvx

import (
	"errors"
	"testing"
)

type Client struct {
	ID   string `csv:"client_id"`
	Name string `csv:"client_name"`
	Age  int    `csv:"client_age"`
}

func TestUnmarshal_StructSlice(t *testing.T) {
	in := []byte("client_id,client_name,client_age\n1,Jose,42\n2,Daniel,26\n")
	var got []Client
	if err := Unmarshal(in, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "Jose" || got[1].Age != 26 {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshal_StructPointerSlice(t *testing.T) {
	in := []byte("client_id,client_name,client_age\n1,Jose,42\n")
	var got []*Client
	if err := Unmarshal(in, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Jose" || got[0].Age != 42 {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshal_RejectUnknownColumn(t *testing.T) {
	in := []byte("client_id,client_name,client_age,extra\n1,Jose,42,x\n")
	var got []Client
	err := Unmarshal(in, &got)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshal_AllowUnknownColumn(t *testing.T) {
	in := []byte("client_id,client_name,client_age,extra\n1,Jose,42,x\n")
	var got []Client
	if err := Unmarshal(in, &got, AllowUnknownColumns(true)); err != nil {
		t.Fatal(err)
	}
	if got[0].Name != "Jose" {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshal_MissingExpectedColumn(t *testing.T) {
	in := []byte("client_id,client_name\n1,Jose\n")
	var got []Client
	err := Unmarshal(in, &got)
	if err == nil {
		t.Fatal("expected error for missing client_age column")
	}
}

func TestUnmarshal_FieldErrorOnBadCell(t *testing.T) {
	in := []byte("client_id,client_name,client_age\n1,Jose,nope\n")
	var got []Client
	err := Unmarshal(in, &got)
	if err == nil {
		t.Fatal("expected error for invalid age cell")
	}
	var fe *FieldError
	if !errors.As(err, &fe) {
		t.Fatalf("expected *FieldError, got %T: %v", err, err)
	}
	if fe.Row != 1 {
		t.Fatalf("Row = %d, want 1", fe.Row)
	}
	if fe.Column != "client_age" {
		t.Fatalf("Column = %q, want client_age", fe.Column)
	}
}

func TestUnmarshal_NoHeaderStruct(t *testing.T) {
	type Row struct {
		A string `csv:"0"`
		B string `csv:"1"`
	}
	in := []byte("x,y\n")
	var got []Row
	if err := Unmarshal(in, &got, NoHeader(true)); err != nil {
		t.Fatal(err)
	}
	if got[0].A != "x" || got[0].B != "y" {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshal_NoHeaderExtraCellsRejected(t *testing.T) {
	type Row struct {
		A string `csv:"0"`
	}
	in := []byte("x,y\n")
	var got []Row
	if err := Unmarshal(in, &got, NoHeader(true)); err == nil {
		t.Fatal("expected unknown cell error")
	}
}

func TestUnmarshal_MapSlice(t *testing.T) {
	in := []byte("a,b\n1,2\n")
	var got []map[string]string
	if err := Unmarshal(in, &got); err != nil {
		t.Fatal(err)
	}
	if got[0]["a"] != "1" || got[0]["b"] != "2" {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshal_NoHeaderStringMatrix(t *testing.T) {
	in := []byte("a,b\nc,d\n")
	var got [][]string
	if err := Unmarshal(in, &got, NoHeader(true)); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0][0] != "a" || got[0][1] != "b" || got[1][0] != "c" || got[1][1] != "d" {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshal_StringMatrixRequiresNoHeader(t *testing.T) {
	in := []byte("a,b\nc,d\n")
	var got [][]string
	err := Unmarshal(in, &got)
	if err == nil {
		t.Fatal("expected error without NoHeader")
	}
	var se *SemanticError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SemanticError, got %T: %v", err, err)
	}
	if se.Msg != "unmarshal into [][]string requires NoHeader" {
		t.Fatalf("Msg = %q", se.Msg)
	}
}
