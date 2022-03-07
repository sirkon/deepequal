package deepequal

import (
	"bytes"
	"google.golang.org/protobuf/proto"
	"reflect"
	"unsafe"
)

// Equal makes x deep equality comparison of given entities with x special care of types generated by
// protoc-gen-go. This is not a drop-in replacement for reflect.DeepEqual as some functionality
// is missing. Should be enough in most cases though.
func Equal(x, y any) bool {
	if x == nil && y == nil {
		return x == y
	}

	xv := reflect.ValueOf(x)
	yv := reflect.ValueOf(y)
	if xv.Type() != yv.Type() {
		return false
	}

	return deepEqual(xv, yv, map[visit]bool{})
}

func deepEqual(x reflect.Value, y reflect.Value, visited map[visit]bool) bool {
	if !x.IsValid() || !y.IsValid() {
		return x.IsValid() == y.IsValid()
	}

	if pbx, ok := getProtoMessage(exportableValue(x).Interface()); ok {
		if pby, ok := getProtoMessage(exportableValue(y).Interface()); ok {
			switch {
			case pbx == nil && pby == nil:
				return true
			case pbx == nil && pby != nil:
				return false
			case pbx != nil && pby == nil:
				return false
			}

			return proto.Equal(pbx, pby)
		}
	}

	// We want to avoid putting more in the visited map than we need to.
	// For any possible reference cycle that might be encountered,
	// hard(x, y) needs to return true for at least one of the types in the cycle,
	// and it's safe and valid to get Value's internal pointer.
	hard := func(v1, v2 reflect.Value) bool {
		switch v1.Kind() {
		case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
			// Nil pointers cannot be cyclic. Avoid putting them in the visited map.
			return !v1.IsNil() && !v2.IsNil()
		}
		return false
	}

	if hard(x, y) {
		// take ptr of reflect.Value
		addr1 := reflectValuePtr(x)
		addr2 := reflectValuePtr(y)

		if uintptr(addr1) > uintptr(addr2) {
			addr1, addr2 = addr2, addr1
		}
		v := visit{
			a1:  addr1,
			a2:  addr2,
			typ: x.Type(),
		}

		if visited[v] {
			return true
		}

		visited[v] = true
	}

	switch x.Kind() {
	case reflect.Array:
		for i := 0; i < x.Len(); i++ {
			if !deepEqual(x.Index(i), x.Index(i), visited) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if x.IsNil() != y.IsNil() {
			return false
		}
		if x.Len() != y.Len() {
			return false
		}
		if x.UnsafePointer() == y.UnsafePointer() {
			return true
		}
		// Special case for []byte, which is common.
		if x.Type().Elem().Kind() == reflect.Uint8 {
			return bytes.Equal(x.Bytes(), y.Bytes())
		}
		for i := 0; i < x.Len(); i++ {
			if !deepEqual(x.Index(i), y.Index(i), visited) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if x.IsNil() || y.IsNil() {
			return x.IsNil() == y.IsNil()
		}
		return deepEqual(x.Elem(), y.Elem(), visited)
	case reflect.Pointer:
		if x.UnsafePointer() == y.UnsafePointer() {
			return true
		}
		return deepEqual(x.Elem(), y.Elem(), visited)
	case reflect.Struct:
		for i, n := 0, x.NumField(); i < n; i++ {
			if !deepEqual(x.Field(i), y.Field(i), visited) {
				return false
			}
		}
		return true
	case reflect.Map:
		if x.IsNil() != y.IsNil() {
			return false
		}
		if x.Len() != y.Len() {
			return false
		}
		if x.UnsafePointer() == y.UnsafePointer() {
			return true
		}
		for _, k := range x.MapKeys() {
			val1 := x.MapIndex(k)
			val2 := y.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() || !deepEqual(val1, val2, visited) {
				return false
			}
		}
		return true
	case reflect.Func:
		if x.IsNil() && y.IsNil() {
			return true
		}
		// Can't do better than this:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return x.Int() == y.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return x.Uint() == y.Uint()
	case reflect.String:
		return x.String() == y.String()
	case reflect.Bool:
		return x.Bool() == y.Bool()
	case reflect.Float32, reflect.Float64:
		return x.Float() == y.Float()
	case reflect.Complex64, reflect.Complex128:
		return x.Complex() == y.Complex()
	default:
		// Don't want to replicate reflect magic
		return false
	}
}

type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ reflect.Type
}

func getProtoMessage(msg interface{}) (proto.Message, bool) {
	if msg == nil {
		return nil, false
	}

	v, ok := msg.(proto.Message)
	if !ok {
		return nil, false
	}

	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
		// This proto.Message can be enum only and generic comparison is
		// sufficient for it.
		return nil, false
	}

	// proto.Message can be satisfied by embedding, but this is not real one.
	for i := 0; i < t.Elem().NumField(); i++ {
		f := t.Elem().Field(i)
		if f.Anonymous {
			return nil, false
		}
	}

	return v, true
}

// reflectValuePtr takes ptr field value of the given reflect.Value
// beware, the field name may change in the future
func reflectValuePtr(v reflect.Value) unsafe.Pointer {
	return reflect.ValueOf(v).FieldByName("ptr").UnsafePointer()
}

// exportableValue returns a value which is a copy of the v minus RO flags
// what are turned off in the new value.
// beware, this may change in the future
func exportableValue(v reflect.Value) reflect.Value {
	type valuePlaceholder struct {
		typ  *struct{}
		ptr  unsafe.Pointer
		flag uintptr
	}

	value := *(*valuePlaceholder)(unsafe.Pointer(&v))
	unsetFlag := ^((1 << 5) | (1 << 6))
	value.flag = uintptr(int(value.flag) & unsetFlag)

	v = *(*reflect.Value)(unsafe.Pointer(&value))
	return v
}
