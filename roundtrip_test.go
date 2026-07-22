package csvx

import (
	"reflect"
	"testing"
)

func TestRoundTrip_Clients(t *testing.T) {
	in := []Client{{"1", "Jose", 42}, {"2", "Daniel", 26}}
	b, err := Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out []Client
	if err := Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("%+v vs %+v", in, out)
	}
}
