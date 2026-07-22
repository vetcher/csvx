package csvx

import (
	"io"
	"reflect"
)

// UnmarshalRecords binds already-parsed CSV rows into v.
// Destination must be a pointer to a slice (same targets as UnmarshalRead).
// Header mode (default): records[0] is the header row; remaining rows are data.
// With NoHeader(true): every row is data (positional / index tags).
func UnmarshalRecords(records [][]string, v any, opts ...Options) error {
	o := joinOptions(opts)
	return unmarshalFromRecordReader(&sliceRecordReader{records: records}, v, o)
}

// UnmarshalRecord binds one already-parsed CSV row into a single struct pointed to by v.
// Requires NoHeader(true) (one row has no header names).
func UnmarshalRecord(record []string, v any, opts ...Options) error {
	o := joinOptions(opts)
	if !o.noHeader {
		return &SemanticError{Msg: "UnmarshalRecord requires NoHeader"}
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &SemanticError{Msg: "unmarshal destination must be a non-nil pointer"}
	}
	rv = rv.Elem()
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return &SemanticError{Msg: "UnmarshalRecord destination must be pointer to struct"}
	}

	plan, err := buildTypePlan(rv.Type(), o)
	if err != nil {
		return err
	}
	maxCol := -1
	for _, fp := range plan.fields {
		if fp.colIndex > maxCol {
			maxCol = fp.colIndex
		}
	}
	if err := validateNoHeaderTrailingCells(record, maxCol, o); err != nil {
		return err
	}
	for _, fp := range plan.fields {
		cell := ""
		if fp.colIndex < len(record) {
			cell = record[fp.colIndex]
		}
		dst := rv.FieldByIndex(fp.index)
		if err := unmarshalField(dst, cell, fp, o); err != nil {
			return &FieldError{Row: 1, Column: noHeaderColumnName(fp), Err: err}
		}
	}
	return nil
}

type sliceRecordReader struct {
	records [][]string
	i       int
}

func (s *sliceRecordReader) Read() ([]string, error) {
	if s.i >= len(s.records) {
		return nil, io.EOF
	}
	rec := s.records[s.i]
	s.i++
	return rec, nil
}

func unmarshalFromRecordReader(reader recordReader, v any, o options) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &SemanticError{Msg: "unmarshal destination must be a non-nil pointer"}
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Slice {
		return &SemanticError{Msg: "unmarshal destination must be pointer to slice"}
	}

	elemType := rv.Type().Elem()
	inner := elemType
	if inner.Kind() == reflect.Ptr {
		inner = inner.Elem()
	}

	switch {
	case isMapStringString(inner):
		if o.noHeader {
			return &SemanticError{Msg: "unmarshal into map[string]string requires header row"}
		}
		return unmarshalMapStringSlice(rv, reader, inner, o)
	case inner.Kind() == reflect.Slice && inner.Elem().Kind() == reflect.String:
		return unmarshalStringMatrix(rv, reader, o)
	case o.noHeader:
		return unmarshalStructSliceNoHeader(rv, reader, elemType, o)
	default:
		return unmarshalStructSliceHeader(rv, reader, elemType, o)
	}
}
