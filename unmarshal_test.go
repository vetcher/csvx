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
