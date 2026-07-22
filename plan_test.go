package csvx

import (
	"errors"
	"reflect"
	"testing"
)

type sample struct {
	ID   int    `csv:"id"`
	Name string `csv:"name"`
	Skip string `csv:"-"`
	Age  int    `csv:",omitempty"`
}

func TestBuildTypePlan_HeaderNames(t *testing.T) {
	p, err := buildTypePlan(reflect.TypeOf(sample{}), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.fields) != 3 {
		t.Fatalf("fields=%d", len(p.fields))
	}
	if p.fields[0].name != "id" || p.fields[1].name != "name" {
		t.Fatalf("%+v", p.fields)
	}
	if p.fields[2].name != "Age" || !p.fields[2].omitEmpty {
		t.Fatalf("age field: %+v", p.fields[2])
	}
	if !p.hasHeader {
		t.Fatal("expected hasHeader")
	}
}

type noHdr struct {
	A string `csv:"0"`
	B string `csv:"1"`
}

func TestBuildTypePlan_NoHeaderIndexes(t *testing.T) {
	p, err := buildTypePlan(reflect.TypeOf(noHdr{}), joinOptions([]Options{NoHeader(true)}))
	if err != nil {
		t.Fatal(err)
	}
	if p.hasHeader {
		t.Fatal("expected no header plan")
	}
	if p.fields[0].colIndex != 0 || p.fields[1].colIndex != 1 {
		t.Fatalf("%+v", p.fields)
	}
}

type bad struct {
	X any `csv:"x"`
}

func TestBuildTypePlan_RejectsAnyField(t *testing.T) {
	_, err := buildTypePlan(reflect.TypeOf(bad{}), joinOptions(nil))
	if err == nil {
		t.Fatal("expected error")
	}
	var sem *SemanticError
	if !errors.As(err, &sem) {
		t.Fatalf("expected *SemanticError, got %T", err)
	}
}

type badPtrAny struct {
	X *any `csv:"x"`
}

func TestBuildTypePlan_RejectsPointerToAnyField(t *testing.T) {
	_, err := buildTypePlan(reflect.TypeOf(badPtrAny{}), joinOptions(nil))
	if err == nil {
		t.Fatal("expected error")
	}
	var sem *SemanticError
	if !errors.As(err, &sem) {
		t.Fatalf("expected *SemanticError, got %T", err)
	}
}

type planInner struct {
	A int `csv:"a"`
}

type planOuter struct {
	planInner `csv:"."`
	B         int `csv:"b"`
}

func TestBuildTypePlan_Inline(t *testing.T) {
	p, err := buildTypePlan(reflect.TypeOf(planOuter{}), joinOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.fields) != 2 {
		t.Fatalf("fields=%d %+v", len(p.fields), p.fields)
	}
	if p.fields[0].name != "a" || p.fields[1].name != "b" {
		t.Fatalf("%+v", p.fields)
	}
}

func TestBuildTypePlan_Cache(t *testing.T) {
	typ := reflect.TypeOf(sample{})
	o := joinOptions(nil)
	p1, err := buildTypePlan(typ, o)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := buildTypePlan(typ, o)
	if err != nil {
		t.Fatal(err)
	}
	if p1 != p2 {
		t.Fatal("expected cached plan pointer")
	}
}
