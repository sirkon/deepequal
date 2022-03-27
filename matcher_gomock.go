package deepequal

import (
	"fmt"
	"reflect"
)

// NewEqMatcher creates equality matcher.
func NewEqMatcher(v any) EqMatcher {
	return EqMatcher{v: v}
}

// EqMatcher equality matcher for gomock. Implements gomock.Matcher.
type EqMatcher struct {
	v any
}

// Matches to satisfy gomock.Matcher
func (e EqMatcher) Matches(x any) bool {
	if e.v == nil || x == nil {
		return Equal(e.v, x)
	}

	a := reflect.ValueOf(e.v)
	b := reflect.ValueOf(x)

	if a.Type().AssignableTo(b.Type()) {
		aConv := a.Convert(b.Type())
		return Equal(aConv.Interface(), b.Interface())
	}

	return false
}

// String to satisfy gomock.Matcher
func (e EqMatcher) String() string {
	return fmt.Sprintf("%v", e.v)
}
