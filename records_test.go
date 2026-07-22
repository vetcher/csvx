package csvx

import "testing"

func TestUnmarshalRecords_HeaderStructSlice(t *testing.T) {
	records := [][]string{
		{"client_id", "client_name", "client_age"},
		{"1", "Jose", "42"},
		{"2", "Daniel", "26"},
	}
	var got []Client
	if err := UnmarshalRecords(records, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "Jose" || got[1].Age != 26 {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshalRecords_NoHeaderStructSlice(t *testing.T) {
	type Row struct {
		A string `csv:"0"`
		B string `csv:"1"`
	}
	records := [][]string{{"x", "y"}, {"a", "b"}}
	var got []Row
	if err := UnmarshalRecords(records, &got, NoHeader(true)); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].A != "x" || got[1].B != "b" {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshalRecord_NoHeaderStruct(t *testing.T) {
	type Row struct {
		A string `csv:"0"`
		B int    `csv:"1"`
	}
	var got Row
	if err := UnmarshalRecord([]string{"hi", "7"}, &got, NoHeader(true)); err != nil {
		t.Fatal(err)
	}
	if got.A != "hi" || got.B != 7 {
		t.Fatalf("%+v", got)
	}
}

func TestUnmarshalRecords_RejectUnknownColumn(t *testing.T) {
	records := [][]string{
		{"client_id", "client_name", "client_age", "extra"},
		{"1", "Jose", "42", "x"},
	}
	var got []Client
	if err := UnmarshalRecords(records, &got); err == nil {
		t.Fatal("expected error")
	}
}
