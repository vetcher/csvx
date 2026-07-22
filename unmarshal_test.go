package csvx

import "testing"

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
