package csvx

import (
	"bytes"
	"io"
	"reflect"
	"sort"
)

func Marshal(v any, opts ...Options) ([]byte, error) {
	var buf bytes.Buffer
	if err := MarshalWrite(&buf, v, opts...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MarshalWrite(w io.Writer, v any, opts ...Options) error {
	o := joinOptions(opts)

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return &SemanticError{Msg: "marshal value must be a non-nil slice"}
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return &SemanticError{Msg: "marshal value must be a slice"}
	}

	elemType := rv.Type().Elem()
	inner := elemType
	if inner.Kind() == reflect.Ptr {
		inner = inner.Elem()
	}
	if inner.Kind() != reflect.Struct {
		return &SemanticError{Msg: "marshal slice element must be struct or pointer to struct"}
	}

	plan, err := buildTypePlan(elemType, o)
	if err != nil {
		return err
	}

	writer := newStdlibWriter(w, o)

	if !o.noHeader {
		header := make([]string, len(plan.fields))
		for i, fp := range plan.fields {
			header[i] = fp.name
		}
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	fields := append([]fieldPlan(nil), plan.fields...)
	if o.noHeader {
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].colIndex < fields[j].colIndex
		})
	}

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				return &SemanticError{Msg: "marshal slice contains nil pointer element"}
			}
			elem = elem.Elem()
		}

		record := make([]string, len(fields))
		for j, fp := range fields {
			fv := elem.FieldByIndex(fp.index)
			cell, err := marshalFieldCell(fv, fp, o)
			if err != nil {
				return &FieldError{Row: i + 1, Column: columnNameForField(fp, o), Err: err}
			}
			record[j] = cell
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func columnNameForField(fp fieldPlan, o options) string {
	if o.noHeader {
		return noHeaderColumnName(fp)
	}
	return fp.name
}

func marshalFieldCell(fv reflect.Value, fp fieldPlan, o options) (string, error) {
	if fp.omitZero && fv.IsZero() {
		return "", nil
	}
	if fp.omitEmpty && isEmptyForOmit(fv) {
		return "", nil
	}
	return marshalCell(fv, o)
}

func isEmptyForOmit(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}
