package csvx

import "testing"

func TestJoinOptions_LastWriteWins(t *testing.T) {
	o := joinOptions([]Options{NoHeader(false), NoHeader(true), AllowUnknownColumns(true)})
	if !o.noHeader {
		t.Fatalf("noHeader=%v, want true", o.noHeader)
	}
	if !o.allowUnknownColumns {
		t.Fatalf("allowUnknownColumns=%v, want true", o.allowUnknownColumns)
	}
	if o.comma != 0 {
		t.Fatalf("comma=%q, want 0 (unset)", o.comma)
	}
}

func TestCommaOption(t *testing.T) {
	o := joinOptions([]Options{Comma(';')})
	if o.comma != ';' {
		t.Fatalf("comma=%q, want ';'", o.comma)
	}
}
