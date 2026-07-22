package csvx

import (
	"encoding"
	"reflect"
	"sync"
)

// Marshaler converts a concrete type to a single CSV cell string.
type Marshaler interface {
	MarshalCSV() (string, error)
}

// Unmarshaler parses a single CSV cell string into a concrete type.
type Unmarshaler interface {
	UnmarshalCSV(string) error
}

type marshalFn func(reflect.Value) (string, error)
type unmarshalFn func(reflect.Value, string) error

// Marshalers holds per-type marshal hooks for local Options.
type Marshalers struct {
	byType map[reflect.Type]marshalFn
}

// Unmarshalers holds per-type unmarshal hooks for local Options.
type Unmarshalers struct {
	byType map[reflect.Type]unmarshalFn
}

// MarshalFunc registers a marshal hook for type T in a Marshalers set.
func MarshalFunc[T any](fn func(T) (string, error)) *Marshalers {
	var zero T
	typ := reflect.TypeOf(zero)
	m := &Marshalers{byType: make(map[reflect.Type]marshalFn)}
	m.byType[typ] = func(v reflect.Value) (string, error) {
		tv, ok := valueAs[T](v)
		if !ok {
			return "", &SemanticError{Msg: "marshaler type mismatch for " + typ.String()}
		}
		return fn(tv)
	}
	return m
}

// UnmarshalFunc registers an unmarshal hook for type T in an Unmarshalers set.
func UnmarshalFunc[T any](fn func(string) (T, error)) *Unmarshalers {
	var zero T
	typ := reflect.TypeOf(zero)
	u := &Unmarshalers{byType: make(map[reflect.Type]unmarshalFn)}
	u.byType[typ] = func(dst reflect.Value, cell string) error {
		val, err := fn(cell)
		if err != nil {
			return err
		}
		if !dst.IsValid() {
			return &SemanticError{Msg: "invalid unmarshal destination"}
		}
		v := reflect.ValueOf(val)
		if !v.Type().AssignableTo(dst.Type()) {
			return &SemanticError{Msg: "unmarshaler type mismatch for " + typ.String()}
		}
		dst.Set(v)
		return nil
	}
	return u
}

// JoinMarshalers merges marshal hook sets; later entries override the same type.
func JoinMarshalers(ms ...*Marshalers) *Marshalers {
	out := &Marshalers{byType: make(map[reflect.Type]marshalFn)}
	for _, m := range ms {
		if m == nil {
			continue
		}
		for k, fn := range m.byType {
			out.byType[k] = fn
		}
	}
	return out
}

// JoinUnmarshalers merges unmarshal hook sets; later entries override the same type.
func JoinUnmarshalers(us ...*Unmarshalers) *Unmarshalers {
	out := &Unmarshalers{byType: make(map[reflect.Type]unmarshalFn)}
	for _, u := range us {
		if u == nil {
			continue
		}
		for k, fn := range u.byType {
			out.byType[k] = fn
		}
	}
	return out
}

func WithMarshalers(m *Marshalers) Options   { return Options{marshalers: m} }
func WithUnmarshalers(u *Unmarshalers) Options { return Options{unmarshalers: u} }

func marshalCell(v reflect.Value, o options) (string, error) {
	if fn, ok := lookupMarshalFn(o.marshalers, v); ok {
		return fn(v)
	}
	if s, ok, err := marshalViaCSVInterface(v); ok {
		return s, err
	}
	if fn, ok := lookupGlobalMarshal(v); ok {
		return fn(v)
	}
	if s, ok, err := marshalViaText(v); ok {
		return s, err
	}
	return formatBuiltin(v)
}

func unmarshalCell(dst reflect.Value, cell string, o options) error {
	if fn, ok := lookupUnmarshalFn(o.unmarshalers, dst); ok {
		return fn(dst, cell)
	}
	if ok, err := unmarshalViaCSVInterface(dst, cell); ok {
		return err
	}
	if fn, ok := lookupGlobalUnmarshal(dst); ok {
		return fn(dst, cell)
	}
	if ok, err := unmarshalViaText(dst, cell); ok {
		return err
	}
	return parseBuiltin(dst, cell, o.allocEmptyPointers)
}

func lookupMarshalFn(m *Marshalers, v reflect.Value) (marshalFn, bool) {
	if m == nil {
		return nil, false
	}
	return lookupMarshalMap(m.byType, v)
}

func lookupUnmarshalFn(u *Unmarshalers, dst reflect.Value) (unmarshalFn, bool) {
	if u == nil {
		return nil, false
	}
	return lookupUnmarshalMap(u.byType, dst)
}

func lookupMarshalMap(byType map[reflect.Type]marshalFn, v reflect.Value) (marshalFn, bool) {
	if fn, ok := byType[v.Type()]; ok {
		return fn, true
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return func(reflect.Value) (string, error) { return "", nil }, true
		}
		if fn, ok := byType[v.Type().Elem()]; ok {
			return func(v reflect.Value) (string, error) { return fn(v.Elem()) }, true
		}
	}
	if v.CanAddr() {
		if fn, ok := byType[reflect.PointerTo(v.Type())]; ok {
			return func(v reflect.Value) (string, error) { return fn(v.Addr()) }, true
		}
	}
	return nil, false
}

func lookupUnmarshalMap(byType map[reflect.Type]unmarshalFn, dst reflect.Value) (unmarshalFn, bool) {
	if fn, ok := byType[dst.Type()]; ok {
		return fn, true
	}
	if dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		if fn, ok := byType[dst.Type().Elem()]; ok {
			return func(dst reflect.Value, cell string) error { return fn(dst.Elem(), cell) }, true
		}
	}
	if dst.CanAddr() {
		if fn, ok := byType[reflect.PointerTo(dst.Type())]; ok {
			return func(dst reflect.Value, cell string) error { return fn(dst.Addr(), cell) }, true
		}
	}
	return nil, false
}

func marshalViaCSVInterface(v reflect.Value) (string, bool, error) {
	if v.CanInterface() {
		if m, ok := v.Interface().(Marshaler); ok {
			s, err := m.MarshalCSV()
			return s, true, err
		}
	}
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(Marshaler); ok {
			s, err := m.MarshalCSV()
			return s, true, err
		}
	}
	return "", false, nil
}

func unmarshalViaCSVInterface(dst reflect.Value, cell string) (bool, error) {
	if dst.CanAddr() {
		if u, ok := dst.Addr().Interface().(Unmarshaler); ok {
			return true, u.UnmarshalCSV(cell)
		}
	}
	if dst.CanInterface() {
		if u, ok := dst.Interface().(Unmarshaler); ok {
			return true, u.UnmarshalCSV(cell)
		}
	}
	return false, nil
}

func marshalViaText(v reflect.Value) (string, bool, error) {
	if v.CanInterface() {
		if m, ok := v.Interface().(encoding.TextMarshaler); ok {
			b, err := m.MarshalText()
			return string(b), true, err
		}
	}
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(encoding.TextMarshaler); ok {
			b, err := m.MarshalText()
			return string(b), true, err
		}
	}
	return "", false, nil
}

func unmarshalViaText(dst reflect.Value, cell string) (bool, error) {
	if dst.CanAddr() {
		if u, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return true, u.UnmarshalText([]byte(cell))
		}
	}
	if dst.CanInterface() {
		if u, ok := dst.Interface().(encoding.TextUnmarshaler); ok {
			return true, u.UnmarshalText([]byte(cell))
		}
	}
	return false, nil
}

func valueAs[T any](v reflect.Value) (T, bool) {
	var zero T
	typ := reflect.TypeOf(zero)
	if v.Type() == typ {
		return v.Interface().(T), true
	}
	if v.Kind() == reflect.Pointer && !v.IsNil() && v.Type().Elem() == typ {
		return v.Elem().Interface().(T), true
	}
	if v.Kind() != reflect.Pointer && v.CanAddr() && reflect.PointerTo(v.Type()) == typ {
		return v.Addr().Interface().(T), true
	}
	return zero, false
}

var (
	registryMu        sync.RWMutex
	globalMarshalFns  = make(map[reflect.Type]marshalFn)
	globalUnmarshalFns = make(map[reflect.Type]unmarshalFn)
)

// RegisterMarshaler installs a process-wide marshal hook for type T.
func RegisterMarshaler[T any](fn func(T) (string, error)) {
	m := MarshalFunc(fn)
	var zero T
	typ := reflect.TypeOf(zero)
	registryMu.Lock()
	globalMarshalFns[typ] = m.byType[typ]
	registryMu.Unlock()
}

// RegisterUnmarshaler installs a process-wide unmarshal hook for type T.
func RegisterUnmarshaler[T any](fn func(string) (T, error)) {
	u := UnmarshalFunc(fn)
	var zero T
	typ := reflect.TypeOf(zero)
	registryMu.Lock()
	globalUnmarshalFns[typ] = u.byType[typ]
	registryMu.Unlock()
}

// ClearRegistry removes all global marshal and unmarshal hooks. Intended for tests.
func ClearRegistry() {
	registryMu.Lock()
	clear(globalMarshalFns)
	clear(globalUnmarshalFns)
	registryMu.Unlock()
}

func lookupGlobalMarshal(v reflect.Value) (marshalFn, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return lookupMarshalMap(globalMarshalFns, v)
}

func lookupGlobalUnmarshal(dst reflect.Value) (unmarshalFn, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return lookupUnmarshalMap(globalUnmarshalFns, dst)
}
