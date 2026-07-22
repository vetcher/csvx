package csvx

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(data []byte, v any, opts ...Options) error {
	return UnmarshalRead(bytes.NewReader(data), v, opts...)
}

func UnmarshalRead(r io.Reader, v any, opts ...Options) error {
	o := joinOptions(opts)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &SemanticError{Msg: "unmarshal destination must be a non-nil pointer"}
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Slice {
		return &SemanticError{Msg: "unmarshal destination must be pointer to slice"}
	}

	elemType := rv.Type().Elem()
	reader := newStdlibReader(r, o)

	inner := elemType
	if inner.Kind() == reflect.Ptr {
		inner = inner.Elem()
	}

	switch {
	case isMapStringString(inner):
		return unmarshalMapStringSlice(rv, reader, inner, o)
	case inner.Kind() == reflect.Slice && inner.Elem().Kind() == reflect.String:
		return unmarshalStringMatrix(rv, reader, o)
	case o.noHeader:
		return unmarshalStructSliceNoHeader(rv, reader, elemType, o)
	default:
		return unmarshalStructSliceHeader(rv, reader, elemType, o)
	}
}

func isMapStringString(t reflect.Type) bool {
	return t.Kind() == reflect.Map &&
		t.Key().Kind() == reflect.String &&
		t.Elem().Kind() == reflect.String
}

func unmarshalMapStringSlice(rv reflect.Value, reader recordReader, mapType reflect.Type, o options) error {
	header, err := reader.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	header = append([]string(nil), header...)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		record = append([]string(nil), record...)

		m := reflect.MakeMap(mapType)
		for i, col := range header {
			val := ""
			if i < len(record) {
				val = record[i]
			}
			m.SetMapIndex(reflect.ValueOf(col), reflect.ValueOf(val))
		}

		elem := m
		if rv.Type().Elem().Kind() == reflect.Ptr {
			ptr := reflect.New(mapType)
			ptr.Elem().Set(m)
			elem = ptr
		}
		rv.Set(reflect.Append(rv, elem))
	}
	return nil
}

func unmarshalStringMatrix(rv reflect.Value, reader recordReader, o options) error {
	if !o.noHeader {
		return &SemanticError{Msg: "unmarshal into [][]string requires NoHeader"}
	}
	rowType := rv.Type().Elem()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		record = append([]string(nil), record...)

		row := reflect.MakeSlice(rowType, len(record), len(record))
		for i, cell := range record {
			row.Index(i).SetString(cell)
		}
		rv.Set(reflect.Append(rv, row))
	}
	return nil
}

func unmarshalStructSliceNoHeader(rv reflect.Value, reader recordReader, elemType reflect.Type, o options) error {
	plan, err := buildTypePlan(elemType, o)
	if err != nil {
		return err
	}

	maxCol := -1
	for _, fp := range plan.fields {
		if fp.colIndex > maxCol {
			maxCol = fp.colIndex
		}
	}

	dataRow := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		record = append([]string(nil), record...)
		dataRow++

		if err := validateNoHeaderTrailingCells(record, maxCol, o); err != nil {
			return err
		}

		elem, err := newSliceElement(elemType)
		if err != nil {
			return err
		}

		for _, fp := range plan.fields {
			cell := ""
			if fp.colIndex < len(record) {
				cell = record[fp.colIndex]
			}
			dst := elem
			if dst.Kind() == reflect.Ptr {
				dst = dst.Elem()
			}
			dst = dst.FieldByIndex(fp.index)
			if err := unmarshalCell(dst, cell, o); err != nil {
				return &FieldError{Row: dataRow, Column: noHeaderColumnName(fp), Err: err}
			}
		}

		rv.Set(reflect.Append(rv, elem))
	}
	return nil
}

func noHeaderColumnName(fp fieldPlan) string {
	if fp.name != "" {
		return fp.name
	}
	return strconv.Itoa(fp.colIndex)
}

func validateNoHeaderTrailingCells(record []string, maxCol int, o options) error {
	if o.allowUnknownColumns || maxCol < 0 {
		return nil
	}
	if len(record) > maxCol+1 {
		return &SemanticError{Msg: "unknown column " + strconvQuote(record[maxCol+1])}
	}
	return nil
}

func unmarshalStructSliceHeader(rv reflect.Value, reader recordReader, elemType reflect.Type, o options) error {
	header, err := reader.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	header = append([]string(nil), header...)

	plan, err := buildTypePlan(elemType, o)
	if err != nil {
		return err
	}

	if err := validateHeaderColumns(header, plan, o); err != nil {
		return err
	}

	colIndex := mapPlanColumns(header, plan, o)

	dataRow := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		record = append([]string(nil), record...)
		dataRow++

		elem, err := newSliceElement(elemType)
		if err != nil {
			return err
		}

		for _, fp := range plan.fields {
			idx := colIndex[fp.name]
			cell := ""
			if idx < len(record) {
				cell = record[idx]
			}
			dst := elem
			if dst.Kind() == reflect.Ptr {
				dst = dst.Elem()
			}
			dst = dst.FieldByIndex(fp.index)
			if err := unmarshalCell(dst, cell, o); err != nil {
				return &FieldError{Row: dataRow, Column: fp.name, Err: err}
			}
		}

		rv.Set(reflect.Append(rv, elem))
	}
	return nil
}

func newSliceElement(elemType reflect.Type) (reflect.Value, error) {
	if elemType.Kind() == reflect.Ptr {
		return reflect.New(elemType.Elem()), nil
	}
	return reflect.New(elemType).Elem(), nil
}

func headerNameEqual(a, b string, insensitive bool) bool {
	if insensitive {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func validateHeaderColumns(header []string, plan *typePlan, o options) error {
	known := make(map[string]struct{}, len(plan.fields))
	for _, fp := range plan.fields {
		known[fp.name] = struct{}{}
	}

	if !o.allowUnknownColumns {
		for _, col := range header {
			if !headerColumnKnown(col, known, o.matchCaseInsensitiveNames) {
				return &SemanticError{Msg: "unknown column " + strconvQuote(col)}
			}
		}
	}

	for _, fp := range plan.fields {
		if findHeaderIndex(header, fp.name, o.matchCaseInsensitiveNames) < 0 {
			return &SemanticError{Msg: "missing column " + strconvQuote(fp.name)}
		}
	}
	return nil
}

func strconvQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func headerColumnKnown(col string, known map[string]struct{}, insensitive bool) bool {
	if insensitive {
		for name := range known {
			if strings.EqualFold(col, name) {
				return true
			}
		}
		return false
	}
	_, ok := known[col]
	return ok
}

func findHeaderIndex(header []string, name string, insensitive bool) int {
	for i, col := range header {
		if headerNameEqual(col, name, insensitive) {
			return i
		}
	}
	return -1
}

func mapPlanColumns(header []string, plan *typePlan, o options) map[string]int {
	m := make(map[string]int, len(plan.fields))
	for _, fp := range plan.fields {
		m[fp.name] = findHeaderIndex(header, fp.name, o.matchCaseInsensitiveNames)
	}
	return m
}
