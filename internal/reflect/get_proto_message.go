package reflect

import (
	"google.golang.org/protobuf/proto"
	"reflect"
)

func getProtoMessage(msg interface{}) (proto.Message, bool) {
	v, ok := msg.(proto.Message)
	if !ok {
		return nil, false
	}

	// if this is not either a struct or a pointer to it then it is all ok, must be an enum
	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Ptr {
		return v, true
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// proto generated structure can't have embeds
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			// включение, точно не генерированная структура
			return nil, false
		}
	}

	return v, true
}
