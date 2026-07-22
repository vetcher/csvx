package csvx

import "fmt"

type SyntaxError struct {
	Msg string
}

func (e *SyntaxError) Error() string { return "csvx: syntax: " + e.Msg }

type SemanticError struct {
	Msg string
}

func (e *SemanticError) Error() string { return "csvx: semantic: " + e.Msg }

type FieldError struct {
	Row    int
	Column string
	Index  int
	Err    error
}

func (e *FieldError) Error() string {
	if e.Column != "" {
		return fmt.Sprintf("csvx: row %d column %q: %v", e.Row, e.Column, e.Err)
	}
	return fmt.Sprintf("csvx: row %d index %d: %v", e.Row, e.Index, e.Err)
}

func (e *FieldError) Unwrap() error { return e.Err }
