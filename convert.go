package csvx

import (
	"reflect"
	"strconv"
)

func parseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

func formatBuiltin(v reflect.Value) (string, error) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits()), nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return string(v.Bytes()), nil
		}
	case reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, v.Len())
			reflect.Copy(reflect.ValueOf(b), v)
			return string(b), nil
		}
	case reflect.Interface:
		return "", &SemanticError{Msg: "interface types are not supported"}
	case reflect.Complex64, reflect.Complex128:
		return "", &SemanticError{Msg: "complex types are not supported"}
	}
	return "", &SemanticError{Msg: "unsupported type " + v.Type().String()}
}

func parseBuiltin(dst reflect.Value, cell string, allocEmptyPointers bool) error {
	if dst.Kind() == reflect.Interface {
		return &SemanticError{Msg: "interface types are not supported"}
	}
	if dst.Kind() == reflect.Pointer {
		if cell == "" && !allocEmptyPointers {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if cell == "" && allocEmptyPointers {
			if dst.IsNil() {
				dst.Set(reflect.New(dst.Type().Elem()))
			}
			return parseBuiltin(dst.Elem(), cell, allocEmptyPointers)
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return parseBuiltin(dst.Elem(), cell, allocEmptyPointers)
	}
	if cell == "" && allocEmptyPointers {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(cell)
		return nil
	case reflect.Bool:
		b, err := parseBool(cell)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(cell, 10, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(cell, 10, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(cell, dst.Type().Bits())
		if err != nil {
			return err
		}
		dst.SetFloat(n)
		return nil
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			dst.SetBytes([]byte(cell))
			return nil
		}
	case reflect.Array:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			if dst.Len() != len(cell) {
				return &SemanticError{Msg: "byte array length mismatch"}
			}
			reflect.Copy(dst, reflect.ValueOf([]byte(cell)))
			return nil
		}
	case reflect.Complex64, reflect.Complex128:
		return &SemanticError{Msg: "complex types are not supported"}
	}
	return &SemanticError{Msg: "unsupported type " + dst.Type().String()}
}
