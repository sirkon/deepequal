package deepequal

import (
	"reflect"
	"unsafe"
)

func getField(value reflect.Value, i int) reflect.Value {
	field := value.Field(i)
	if !field.CanAddr() {
		value2 := reflect.New(value.Type()).Elem()
		value2.Set(value)
		field = value2.Field(i)
	}

	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}
