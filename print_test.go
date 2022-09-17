package deepequal

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/sirkon/deepequal/internal/diff"
	"github.com/sirkon/deepequal/internal/testdata"
)

func TestPrint(t *testing.T) {
	type testStruct struct {
		a string
		b *testdata.Sample
	}

	printItem(t, 1, nil)
	printItem(t, "Hello!", nil)
	printItem(t, []bool{true, true, false}, nil)
	printItem(t, map[string]int{
		"1": 1,
		"a": 10,
	}, nil)
	printItem(t, testStruct{
		a: "123",
		b: &testdata.Sample{
			Str: "string",
			Sub: &testdata.Sub{
				Val: 12,
			},
		},
	}, nil)
	printItem(t, &testStruct{
		a: "123",
		b: &testdata.Sample{
			Str: "string",
			Sub: &testdata.Sub{
				Val: 12,
			},
		},
	}, nil)
	printItem(t, proto.Message(&testdata.Sample{
		Str: "string",
		Sub: &testdata.Sub{
			Val: 13,
		},
	}), nil)
	printItem(t, (*string)(nil), nil)
	printItem(t, []int(nil), nil)
	printItem(t, map[int]string(nil), nil)
	printItem(t, []int{}, nil)
	printItem(t, map[int]string{}, nil)
	printItem(t, struct{}{}, nil)
	printItem(t, []int{1, 2, 3}, &diff.Indices{
		Right: map[int]diff.Diff{
			1: &diff.Missing{},
		},
	})
	printItem(
		t,
		map[int]string{
			1: "Hello",
			2: "World!",
		},
		&diff.Keys{
			Right: map[any]diff.Diff{
				2: &diff.Missing{},
			},
		},
	)
	printItem(
		t,
		&testStruct{
			a: "abcd",
			b: &testdata.Sample{
				Str: "str",
				Sub: &testdata.Sub{
					Val: 14,
				},
			},
		},
		&diff.Fields{
			Fields: map[string]diff.Diff{
				"b": &diff.Fields{
					Fields: map[string]diff.Diff{
						"Sub": &diff.Value{},
					},
				},
			},
		},
	)
	printItem(
		t,
		map[string]any{
			"a": "one",
			"b": 1,
			"c": true,
		},
		&diff.Keys{
			Right: map[any]diff.Diff{
				"a": &diff.Missing{},
			},
		},
	)
}

func TestSunHighlight(t *testing.T) {
	type testStruct struct {
		a string
		b *testdata.Sample
	}

	printItem(
		t,
		&testStruct{
			a: "abcd",
			b: &testdata.Sample{
				Str: "str",
				Sub: &testdata.Sub{
					Val: 14,
				},
			},
		},
		&diff.Fields{
			Fields: map[string]diff.Diff{
				"b": &diff.Fields{
					Fields: map[string]diff.Diff{
						"Sub": &diff.Value{},
					},
				},
			},
		},
	)
}

func printItem(t *testing.T, v any, d diff.Diff) {
	p := &printer{
		buf:         &bytes.Buffer{},
		formatDepth: 0,
		isLeft:      false,
	}
	p.printValue("", reflect.ValueOf(v), d, false, false)
	t.Log("\r", p.buf.String())
}
