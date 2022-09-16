package deepequal

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/sirkon/deepequal/internal/diff"
)

// Builds a difference between left and right values.
func difference(l, r reflect.Value, isProto bool) diff.Diff {
	if !l.IsValid() {
		panic(fmt.Errorf("left is %s", l.String()))
	}
	if !r.IsValid() {
		panic(fmt.Errorf("right is %s", r.String()))
	}

	if Equal(l.Interface(), r.Interface()) {
		return nil
	}

	if l.Kind() != r.Kind() {
		return &diff.Type{
			Left:  l.Type().String(),
			Right: r.Type().String(),
		}
	}

	switch l.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Uintptr, reflect.UnsafePointer:
		return &diff.Value{}
	case reflect.Slice, reflect.Array:
		if l.IsZero() || r.IsZero() {
			// They can't be zero both, would be equal then.
			return &diff.Value{}
		}
		if l.Len() == 0 || r.Len() == 0 {
			// Same, they can't have zero length both for the same reason as above.
			return &diff.Value{}
		}

		common := lcs(l, r)
		return &diff.Indices{
			Left:  findUncommon(l, common),
			Right: findUncommon(r, common),
		}

	case reflect.Map:
		if l.IsZero() || r.IsZero() {
			// They can't be zero both, would be equal then.
			return &diff.Value{}
		}

		res := &diff.Keys{
			Left:  map[any]diff.Diff{},
			Right: map[any]diff.Diff{},
		}

		for _, key := range l.MapKeys() {
			rv := r.MapIndex(key)
			if !rv.IsValid() {
				res.Left[key.Interface()] = &diff.Missing{}
				continue
			}

			if v := difference(l.MapIndex(key), rv, isProto); v != nil {
				res.Left[key.Interface()] = v
			}
		}

		for _, key := range r.MapKeys() {
			lv := l.MapIndex(key)
			if !lv.IsValid() {
				res.Right[key.Interface()] = &diff.Missing{}
				continue
			}

			if v := difference(lv, r.MapIndex(key), isProto); v != nil {
				res.Right[key.Interface()] = v
			}
		}

		return res

	case reflect.Struct:

		res := &diff.Fields{
			Fields: map[string]diff.Diff{},
		}

		for i := 0; i < l.NumField(); i++ {
			if isProto && !l.Type().Field(i).IsExported() {
				// Pass unexported fields in proto message.
				continue
			}
			fl := getField(l, i)
			fr := getField(r, i)

			if v := difference(fl, fr, isProto); v != nil {
				res.Fields[l.Type().Field(i).Name] = v
			}
		}

		return res
	case reflect.Pointer, reflect.Interface:
		if l.IsZero() || r.IsZero() {
			return &diff.Value{}
		}

		_, isProto = l.Interface().(proto.Message)
		return difference(l.Elem(), r.Elem(), isProto)

	default:
		panic(fmt.Errorf("cannot diff values of %T", l.Interface()))
	}
}

func lcs(x, y reflect.Value) reflect.Value {
	if x.Len() == 0 || y.Len() == 0 {
		return reflect.Zero(x.Type())
	}

	xlast := x.Index(x.Len() - 1)
	ylast := y.Index(y.Len() - 1)

	if Equal(xlast.Interface(), ylast.Interface()) {
		res := lcs(x.Slice(0, x.Len()-1), y.Slice(0, y.Len()-1))
		return reflect.Append(res, xlast)
	}

	first := lcs(x.Slice(0, x.Len()-1), y)
	second := lcs(x, y.Slice(0, y.Len()-1))
	if first.Len() < second.Len() {
		return second
	}

	return first
}

func findUncommon(src, known reflect.Value) map[int]diff.Diff {
	info := map[int]diff.Diff{}

	var j int
	for i := 0; i < src.Len(); i++ {
		if j >= known.Len() {
			info[i] = &diff.Missing{}
			continue
		}

		if Equal(src.Index(i).Interface(), known.Index(j).Interface()) {
			j++
			continue
		}

		info[i] = &diff.Missing{}
	}

	return info
}
