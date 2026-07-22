package csvx

import (
	"io"
	"reflect"
	"sort"
)

type Encoder struct {
	writer        recordWriter
	o             options
	headerWritten bool
	plan          *typePlan
}

func NewEncoder(w io.Writer, opts ...Options) *Encoder {
	o := joinOptions(opts)
	return &Encoder{
		writer: newStdlibWriter(w, o),
		o:      o,
	}
}

func (e *Encoder) Encode(v any) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return &SemanticError{Msg: "encode value must be a non-nil struct"}
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return &SemanticError{Msg: "encode value must be a struct"}
	}

	plan, err := buildTypePlan(rv.Type(), e.o)
	if err != nil {
		return err
	}
	e.plan = plan

	if !e.o.noHeader && !e.headerWritten {
		header := make([]string, len(plan.fields))
		for i, fp := range plan.fields {
			header[i] = fp.name
		}
		if err := e.writer.Write(header); err != nil {
			return err
		}
		e.headerWritten = true
	}

	fields := append([]fieldPlan(nil), plan.fields...)
	if e.o.noHeader {
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].colIndex < fields[j].colIndex
		})
	}

	record := make([]string, len(fields))
	for j, fp := range fields {
		fv := rv.FieldByIndex(fp.index)
		cell, err := marshalFieldCell(fv, fp, e.o)
		if err != nil {
			return &FieldError{Row: 1, Column: columnNameForField(fp, e.o), Err: err}
		}
		record[j] = cell
	}
	if err := e.writer.Write(record); err != nil {
		return err
	}
	return e.writer.Flush()
}
