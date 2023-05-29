package deepequal

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/sirkon/deepequal/internal/diff"
)

const (
	formatBold  = "\033[1m"
	formatGreen = "\033[32m"
	formatRed   = "\033[31m"
)

type printer struct {
	buf         *bytes.Buffer
	formatDepth int
	isLeft      bool
}

func newPrinter(isLeft bool) *printer {
	return &printer{
		buf:         &bytes.Buffer{},
		formatDepth: 0,
		isLeft:      isLeft,
	}
}

func (p *printer) printValue(
	offset string,
	v reflect.Value,
	d diff.Diff,
	isProto bool,
	showType bool,
	stack map[uintptr]struct{},
) {
	p.setFormatOn(d)
	defer p.setFormatOff(d)
	noff := offset + "  "

	p.setColorOn()
	defer p.setColorOff()

	t := v.Type()
	switch t.Kind() {
	case
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.UnsafePointer,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:

		if _, ok := v.Interface().(fmt.Stringer); ok {
			if showType {
				_, _ = fmt.Fprintf(p.buf, "%s(%s)", v.Type().String(), v.Interface())
			} else {
				_, _ = fmt.Fprintf(p.buf, "%s", v.Interface())
			}
		} else {
			if showType {
				_, _ = fmt.Fprintf(p.buf, "%s(%v)", v.Type().String(), v.Interface())
			} else {
				_, _ = fmt.Fprintf(p.buf, "%v", v.Interface())
			}
		}

	case reflect.Bool:
		if t == reflect.TypeOf(true) && !showType {
			_, _ = fmt.Fprintf(p.buf, "%v", v.Interface())
		} else {
			_, _ = fmt.Fprintf(p.buf, "%s(%v)", v.Type().String(), v.Interface())
		}

	case reflect.String:
		if t == reflect.TypeOf("") {
			_, _ = fmt.Fprintf(p.buf, "%q", v.Interface())
		} else {
			_, _ = fmt.Fprintf(p.buf, "%s(%q)", v.Type().String(), v.Interface())
		}

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			if v.IsNil() {
				_, _ = fmt.Fprintf(p.buf, "%s(nil)", v.Type().String())
				return
			}

			_, _ = fmt.Fprintf(p.buf, "%s{}", v.Type().String())
			return
		}

		_, _ = fmt.Fprintf(p.buf, "%s{\n\r", v.Type().String())
		ds := p.sliceDiff(d)

		for i := 0; i < v.Len(); i++ {
			vi := v.Index(i)
			p.buf.WriteString(noff)

			if vdiff := ds[i]; vdiff != nil {
				p.printValue(noff, vi, vdiff, false, false, stack)
			} else {
				p.printValue(noff, vi, nil, false, false, stack)
			}

			p.buf.WriteString(",\n\r")
		}
		p.buf.WriteString(offset)
		p.buf.WriteString("}")

	case reflect.Map:
		if v.Len() == 0 {
			if v.IsNil() {
				_, _ = fmt.Fprintf(p.buf, "%s(nil)", v.Type().String())
				return
			}

			_, _ = fmt.Fprintf(p.buf, "%s{}", v.Type().String())
			return
		}

		_, _ = fmt.Fprintf(p.buf, "%s{\n\r", v.Type().String())
		ds := reflect.ValueOf(p.mapDiff(d))
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return compareReflectValues(keys[i], keys[j])
		})

		for _, key := range keys {
			p.buf.WriteString(noff)

			dm := ds.MapIndex(key)
			if dm.IsValid() {
				p.printValue(noff, key, dm.Interface().(diff.Diff), false, false, stack)
				p.buf.WriteString(": ")
				p.printValue(noff, v.MapIndex(key), dm.Interface().(diff.Diff), false, false, stack)
			} else {
				p.printValue(noff, key, nil, false, false, stack)
				p.buf.WriteString(": ")
				p.printValue(noff, v.MapIndex(key), nil, false, false, stack)
			}

			p.buf.WriteString(",\n\r")
		}

		p.buf.WriteString(offset)
		p.buf.WriteString("}")

	case reflect.Struct:
		if v.NumField() == 0 {
			_, _ = fmt.Fprintf(p.buf, "%s{}", v.Type().String())
			return
		}

		_, _ = fmt.Fprintf(p.buf, "%s{\n\r", v.Type().String())
		ds := p.structDiff(d)

		for i := 0; i < v.NumField(); i++ {
			if isProto && !t.Field(i).IsExported() {
				continue
			}

			fieldName := t.Field(i).Name
			vs := ds[fieldName]
			p.buf.WriteString(noff)

			if vs != nil {
				p.setFormatOn(vs)
				p.setColorOn()
				p.buf.WriteString(fieldName)
				p.buf.WriteString("\033[0m: ")
				p.printValue(noff, getField(v, i), vs, false, false, stack)
				p.setFormatOff(vs)
				p.setColorOff()
			} else {
				p.setColorOn()
				p.buf.WriteString(fieldName)
				p.setColorOff()
				p.buf.WriteString(": ")
				p.printValue(noff, getField(v, i), nil, false, false, stack)
			}

			p.buf.WriteString(",\n\r")
		}

		p.buf.WriteString(offset)
		p.setColorOn()
		p.buf.WriteString("}")
		p.setColorOff()

	case reflect.Pointer:
		if v.IsNil() {
			p.buf.WriteString(t.String())
			p.buf.WriteString("(nil)")
			return
		}

		addr := v.Pointer()
		if _, ok := stack[addr]; ok {
			p.buf.WriteByte('(')
			p.buf.WriteString(t.String())
			p.buf.WriteByte(')')
			_, _ = fmt.Fprintf(p.buf, "(%x)", addr)
			return
		}
		stack[addr] = struct{}{}

		p.buf.WriteByte('&')
		// _, _ = fmt.Fprintf(p.buf, "(%x)", addr)
		_, ip := v.Interface().(proto.Message)
		p.printValue(offset, v.Elem(), d, ip, false, stack)

	case reflect.Interface:
		if v.IsNil() {
			p.buf.WriteString(t.String())
			p.buf.WriteString("(nil)")
			return
		}

		var xxx proto.Message
		p.printValue(offset, v.Elem(), d, t == reflect.TypeOf(xxx), true, stack)

	default:
		panic(fmt.Errorf("type %s is not supported for printing", t.String()))
	}
}

func (p *printer) setColorOn() {
	if p.formatDepth == 0 {
		return
	}

	color := formatRed
	if p.isLeft {
		color = formatGreen
	}
	p.buf.WriteString(color)
}

func (p *printer) setColorOff() {
	if p.formatDepth == 0 {
		return
	}

	p.buf.WriteString("\033[0n")
}

func (p *printer) setFormatOn(d diff.Diff) {
	if d == nil {
		return
	}

	switch d.(type) {
	case *diff.Indices, *diff.Fields, *diff.Keys:
		return
	}

	if p.formatDepth == 0 {
		color := formatRed
		if p.isLeft {
			color = formatGreen
		}
		p.buf.WriteString(color)
	}
	p.formatDepth++
}

func (p *printer) setFormatOff(d diff.Diff) {
	if d == nil {
		return
	}

	switch d.(type) {
	case *diff.Indices, *diff.Fields, *diff.Keys:
		return
	}

	p.formatDepth--
	if p.formatDepth == 0 {
		p.buf.WriteString("\033[0m")
	}
}

func (p *printer) sliceDiff(v diff.Diff) map[int]diff.Diff {
	vv, ok := v.(*diff.Indices)
	if !ok {
		return nil
	}

	if p.isLeft {
		return vv.Left
	}

	return vv.Right
}

func (p *printer) mapDiff(v diff.Diff) map[any]diff.Diff {
	vv, ok := v.(*diff.Keys)
	if !ok {
		return nil
	}

	if p.isLeft {
		return vv.Left
	}

	return vv.Right
}

func (p *printer) structDiff(v diff.Diff) map[string]diff.Diff {
	vv, ok := v.(*diff.Fields)
	if !ok {
		return nil
	}

	return vv.Fields
}

func compareReflectValues(a, b reflect.Value) bool {
	if a.Type().Kind() != b.Type().Kind() {
		return true
	}

	switch a.Type().Kind() {
	case reflect.Int:
		return getReflectValue[int](a) < getReflectValue[int](b)
	case reflect.Int8:
		return getReflectValue[int8](a) < getReflectValue[int8](b)
	case reflect.Int16:
		return getReflectValue[int16](a) < getReflectValue[int16](b)
	case reflect.Int32:
		return getReflectValue[int32](a) < getReflectValue[int32](b)
	case reflect.Int64:
		return getReflectValue[int64](a) < getReflectValue[int64](b)
	case reflect.Uint:
		return getReflectValue[uint](a) < getReflectValue[uint](b)
	case reflect.Uint8:
		return getReflectValue[uint8](a) < getReflectValue[uint8](b)
	case reflect.Uint16:
		return getReflectValue[uint16](a) < getReflectValue[uint16](b)
	case reflect.Uint32:
		return getReflectValue[uint32](a) < getReflectValue[uint32](b)
	case reflect.Uint64:
		return getReflectValue[uint64](a) < getReflectValue[uint64](b)
	case reflect.Uintptr:
		return getReflectValue[uintptr](a) < getReflectValue[uintptr](b)
	case reflect.Float32:
		return getReflectValue[float32](a) < getReflectValue[float32](b)
	case reflect.Float64:
		return getReflectValue[float64](a) < getReflectValue[float64](b)
	case reflect.String:
		return getReflectValue[string](a) < getReflectValue[string](b)
	default:
		return true
	}
}

func getReflectValue[T any](v reflect.Value) T {
	var tmp T
	vv := reflect.ValueOf(&tmp)
	vv.Elem().Set(v)

	return vv.Elem().Interface().(T)
}
