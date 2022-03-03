package reflect

// digValue returns field value even of an unexported one
func digValue(field Value) interface{} {
	return field.Interface()
}
