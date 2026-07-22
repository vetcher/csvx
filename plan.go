package csvx

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type fieldPlan struct {
	index              []int
	name               string
	colIndex           int
	omitEmpty, omitZero bool
	inline             bool
}

type typePlan struct {
	fields    []fieldPlan
	hasHeader bool
}

type planKey struct {
	typ reflect.Type
	fp  uintptr
}

var planCache sync.Map

func optionsPlanFingerprint(o options) uintptr {
	var fp uintptr
	if o.noHeader {
		fp = 1
	}
	return fp
}

func buildTypePlan(t reflect.Type, o options) (*typePlan, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	key := planKey{typ: t, fp: optionsPlanFingerprint(o)}
	if cached, ok := planCache.Load(key); ok {
		return cached.(*typePlan), nil
	}
	p, err := buildTypePlanUncached(t, o)
	if err != nil {
		return nil, err
	}
	actual, _ := planCache.LoadOrStore(key, p)
	return actual.(*typePlan), nil
}

func buildTypePlanUncached(t reflect.Type, o options) (*typePlan, error) {
	if t.Kind() != reflect.Struct {
		return nil, &SemanticError{Msg: "plan requires struct type"}
	}
	var fields []fieldPlan
	col := 0
	if err := walkStructFields(t, nil, o, &fields, &col); err != nil {
		return nil, err
	}
	return &typePlan{
		fields:    fields,
		hasHeader: !o.noHeader,
	}, nil
}

func walkStructFields(st reflect.Type, indexPrefix []int, o options, out *[]fieldPlan, col *int) error {
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}
		idx := append(append([]int(nil), indexPrefix...), i)

		name, omitEmpty, omitZero, skip, inline := parseCSVTag(sf)
		if skip {
			continue
		}

		ft := sf.Type
		if inline {
			for ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() != reflect.Struct {
				return &SemanticError{Msg: "inline field must be a struct"}
			}
			if err := walkStructFields(ft, idx, o, out, col); err != nil {
				return err
			}
			continue
		}

		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Interface {
			return &SemanticError{Msg: "interface types are not supported"}
		}

		fp := fieldPlan{
			index:     idx,
			omitEmpty: omitEmpty,
			omitZero:  omitZero,
		}

		if o.noHeader {
			if n, err := strconv.Atoi(name); err == nil && name != "" {
				fp.colIndex = n
			} else {
				fp.colIndex = *col
				*col++
			}
		} else {
			if name == "" {
				name = sf.Name
			}
			fp.name = name
			fp.colIndex = *col
			*col++
		}

		*out = append(*out, fp)
	}
	return nil
}

func parseCSVTag(sf reflect.StructField) (name string, omitEmpty, omitZero, skip, inline bool) {
	tag := sf.Tag.Get("csv")
	if tag == "" {
		return "", false, false, false, false
	}
	if tag == "-" {
		return "", false, false, true, false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	for _, p := range parts[1:] {
		switch strings.TrimSpace(p) {
		case "omitempty":
			omitEmpty = true
		case "omitzero":
			omitZero = true
		}
	}
	if name == "." {
		inline = true
		name = ""
	}
	return name, omitEmpty, omitZero, skip, inline
}
