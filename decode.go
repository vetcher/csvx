package csvx

import (
	"io"
	"reflect"
)

type Decoder struct {
	reader       recordReader
	o            options
	header       []string
	headerDone   bool
	colIndex     map[string]int
	plan         *typePlan
	planElemType reflect.Type
	exhausted    bool
	dataRow      int
}

func NewDecoder(r io.Reader, opts ...Options) *Decoder {
	o := joinOptions(opts)
	return &Decoder{
		reader: newStdlibReader(r, o),
		o:      o,
	}
}

func (d *Decoder) Decode(v any) error {
	if d.exhausted {
		return io.EOF
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &SemanticError{Msg: "decode destination must be a non-nil pointer"}
	}
	rv = rv.Elem()

	if rv.Kind() == reflect.Slice {
		return d.decodeSlice(rv)
	}
	return d.decodeOne(rv)
}

func (d *Decoder) decodeOne(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Struct:
		return d.decodeOneStruct(rv, rv.Type())
	case reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return d.decodeOne(rv.Elem())
	case reflect.Map:
		if !isMapStringString(rv.Type()) {
			return &SemanticError{Msg: "decode destination must be struct, map[string]string, or slice"}
		}
		return d.decodeOneMap(rv)
	default:
		return &SemanticError{Msg: "decode destination must be struct, map[string]string, or slice"}
	}
}

func (d *Decoder) decodeOneStruct(rv reflect.Value, structType reflect.Type) error {
	if err := d.ensureHeader(structType); err != nil {
		return err
	}

	record, err := d.reader.Read()
	if err == io.EOF {
		return io.EOF
	}
	if err != nil {
		return err
	}
	record = append([]string(nil), record...)
	d.dataRow++

	if err := d.fillStructFromRecord(rv, record); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) decodeOneMap(rv reflect.Value) error {
	mapType := rv.Type()
	if err := d.ensureHeaderForMap(); err != nil {
		return err
	}

	record, err := d.reader.Read()
	if err == io.EOF {
		return io.EOF
	}
	if err != nil {
		return err
	}
	record = append([]string(nil), record...)
	d.dataRow++

	m := reflect.MakeMap(mapType)
	for i, col := range d.header {
		val := ""
		if i < len(record) {
			val = record[i]
		}
		m.SetMapIndex(reflect.ValueOf(col), reflect.ValueOf(val))
	}
	rv.Set(m)
	return nil
}

func (d *Decoder) decodeSlice(rv reflect.Value) error {
	elemType := rv.Type().Elem()
	inner := elemType
	if inner.Kind() == reflect.Ptr {
		inner = inner.Elem()
	}

	rv.SetLen(0)

	var n int
	var err error
	switch {
	case isMapStringString(inner):
		n, err = d.decodeMapSlice(rv, elemType, inner)
	case inner.Kind() == reflect.Struct:
		n, err = d.decodeStructSlice(rv, elemType, inner)
	default:
		return &SemanticError{Msg: "decode slice element must be struct, pointer to struct, or map[string]string"}
	}
	if err != nil {
		return err
	}

	d.exhausted = true
	if n == 0 {
		return io.EOF
	}
	return nil
}

func (d *Decoder) decodeStructSlice(rv reflect.Value, elemType, inner reflect.Type) (int, error) {
	if err := d.ensureHeader(inner); err != nil {
		return 0, err
	}

	var n int
	for {
		record, err := d.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return n, err
		}
		record = append([]string(nil), record...)
		d.dataRow++

		elem, err := newSliceElement(elemType)
		if err != nil {
			return n, err
		}

		dst := elem
		if dst.Kind() == reflect.Ptr {
			dst = dst.Elem()
		}
		if err := d.fillStructFromRecord(dst, record); err != nil {
			return n, err
		}

		rv.Set(reflect.Append(rv, elem))
		n++
	}
	return n, nil
}

func (d *Decoder) decodeMapSlice(rv reflect.Value, elemType, mapType reflect.Type) (int, error) {
	if err := d.ensureHeaderForMap(); err != nil {
		return 0, err
	}

	var n int
	for {
		record, err := d.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return n, err
		}
		record = append([]string(nil), record...)
		d.dataRow++

		m := reflect.MakeMap(mapType)
		for i, col := range d.header {
			val := ""
			if i < len(record) {
				val = record[i]
			}
			m.SetMapIndex(reflect.ValueOf(col), reflect.ValueOf(val))
		}

		elem := m
		if elemType.Kind() == reflect.Ptr {
			ptr := reflect.New(mapType)
			ptr.Elem().Set(m)
			elem = ptr
		}
		rv.Set(reflect.Append(rv, elem))
		n++
	}
	return n, nil
}

func (d *Decoder) ensureHeaderForMap() error {
	if d.o.noHeader {
		return &SemanticError{Msg: "decode into map[string]string requires header row"}
	}
	if d.headerDone {
		return nil
	}
	header, err := d.reader.Read()
	if err == io.EOF {
		return io.EOF
	}
	if err != nil {
		return err
	}
	d.header = append([]string(nil), header...)
	d.headerDone = true
	return nil
}

func (d *Decoder) ensureHeader(structType reflect.Type) error {
	plan, err := d.ensurePlan(structType)
	if err != nil {
		return err
	}
	if d.o.noHeader {
		return nil
	}
	if d.headerDone {
		return nil
	}
	header, err := d.reader.Read()
	if err == io.EOF {
		return io.EOF
	}
	if err != nil {
		return err
	}
	d.header = append([]string(nil), header...)
	d.headerDone = true

	if err := validateHeaderColumns(d.header, plan, d.o); err != nil {
		return err
	}
	d.colIndex = mapPlanColumns(d.header, plan, d.o)
	return nil
}

func (d *Decoder) ensurePlan(structType reflect.Type) (*typePlan, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	if d.plan != nil && d.planElemType == structType {
		return d.plan, nil
	}
	plan, err := buildTypePlan(structType, d.o)
	if err != nil {
		return nil, err
	}
	d.plan = plan
	d.planElemType = structType
	if d.headerDone && !d.o.noHeader {
		if err := validateHeaderColumns(d.header, plan, d.o); err != nil {
			return nil, err
		}
		d.colIndex = mapPlanColumns(d.header, plan, d.o)
	}
	return plan, nil
}

func (d *Decoder) fillStructFromRecord(elem reflect.Value, record []string) error {
	plan := d.plan
	if d.o.noHeader {
		if err := validateNoHeaderTrailingCells(record, maxColIndex(plan), d.o); err != nil {
			return err
		}
	}
	for _, fp := range plan.fields {
		cell := ""
		if d.o.noHeader {
			if fp.colIndex < len(record) {
				cell = record[fp.colIndex]
			}
		} else {
			idx := d.colIndex[fp.name]
			if idx >= 0 && idx < len(record) {
				cell = record[idx]
			}
		}
		dst := elem.FieldByIndex(fp.index)
		col := fp.name
		if d.o.noHeader {
			col = noHeaderColumnName(fp)
		}
		if err := unmarshalField(dst, cell, fp, d.o); err != nil {
			return &FieldError{Row: d.dataRow, Column: col, Err: err}
		}
	}
	return nil
}

func maxColIndex(plan *typePlan) int {
	maxCol := -1
	for _, fp := range plan.fields {
		if fp.colIndex > maxCol {
			maxCol = fp.colIndex
		}
	}
	return maxCol
}
