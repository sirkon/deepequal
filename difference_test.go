package deepequal

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/sirkon/deepequal/internal/diff"
	"github.com/sirkon/deepequal/internal/testdata"
)

func TestDifference(t *testing.T) {
	type test struct {
		name  string
		a     any
		b     any
		want  diff.Diff
		panic bool
	}
	type testSet struct {
		name  string
		tests []test
	}
	type sampleStruct struct {
		a int
		b string
		s *sampleStruct
	}
	proto1 := &testdata.Sample{
		Str: "str",
		Sub: &testdata.Sub{
			Val: 12,
		},
	}
	proto2 := &testdata.Sample{
		Str: "str",
		Sub: &testdata.Sub{
			Val: 13,
		},
	}
	if _, err := proto.Marshal(proto2); err != nil {
		t.Error(fmt.Errorf("marshal proto: %w", err))
		return
	}

	tests := []testSet{
		{
			name: "types handling",
			tests: []test{
				{
					name: "different types",
					a:    1,
					b:    "2",
					want: &diff.Type{
						Left:  "int",
						Right: "string",
					},
				},
				{
					name:  "unsupported type",
					a:     make(chan struct{}),
					b:     make(chan struct{}),
					want:  nil,
					panic: true,
				},
				{
					name:  "invalid a type",
					a:     nil,
					b:     1,
					want:  nil,
					panic: true,
				},
				{
					name:  "invalid b type",
					a:     1,
					b:     nil,
					want:  nil,
					panic: true,
				},
				{
					name: "proto",
					a:    proto1,
					b:    proto2,
					want: &diff.Fields{
						Fields: map[string]diff.Diff{
							"Sub": &diff.Fields{
								Fields: map[string]diff.Diff{
									"Val": &diff.Value{},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "raw values",
			tests: []test{
				{
					name: "integers equal",
					a:    15,
					b:    15,
					want: nil,
				},
				{
					name: "integers not equal",
					a:    1,
					b:    2,
					want: &diff.Value{},
				},
				{
					name: "strings equal",
					a:    "1",
					b:    "1",
					want: nil,
				},
				{
					name: "strings not equal",
					a:    "1",
					b:    "2",
					want: &diff.Value{},
				},
			},
		},
		{
			name: "slices",
			tests: []test{
				{
					name: "slices empty",
					a:    []bool{},
					b:    []bool{},
					want: nil,
				},
				{
					name: "emtpy and nil slice",
					a:    []int{},
					b:    *new([]int),
					want: &diff.Value{},
				},
				{
					name: "slices equal",
					a:    []int{1, 2, 3},
					b:    []int{1, 2, 3},
					want: nil,
				},
				{
					name: "slices different length",
					a:    []int{1, 2, 3, 4},
					b:    []int{1, 2, 3},
					want: &diff.Indices{
						Left: map[int]diff.Diff{
							3: &diff.Missing{},
						},
						Right: map[int]diff.Diff{},
					},
				},
				{
					name: "slices not equal",
					a:    []bool{true, false, true},
					b:    []bool{false, false, false},
					want: &diff.Indices{
						Left: map[int]diff.Diff{
							0: &diff.Missing{},
							2: &diff.Missing{},
						},
						Right: map[int]diff.Diff{
							1: &diff.Missing{},
							2: &diff.Missing{},
						},
					},
				},
				{
					name: "empty and not empty slice",
					a:    []int{},
					b:    []int{1, 2, 3},
					want: &diff.Value{},
				},
				{
					name: "to cover lcs algorithm",
					a:    []int{1, 2, 3},
					b:    []int{1, 2, 3, 4},
					want: &diff.Indices{
						Left: map[int]diff.Diff{},
						Right: map[int]diff.Diff{
							3: &diff.Missing{},
						},
					},
				},
			},
		},
		{
			name: "maps",
			tests: []test{
				{
					name: "empty maps",
					a:    map[int]bool{},
					b:    map[int]bool{},
					want: nil,
				},
				{
					name: "nil and not nil map",
					a:    (map[int]bool)(nil),
					b:    map[int]bool{},
					want: &diff.Value{},
				},
				{
					name: "maps equal",
					a: map[int]bool{
						0: false,
						1: true,
					},
					b: map[int]bool{
						0: false,
						1: true,
					},
					want: nil,
				},
				{
					name: "empty and nil map",
					a:    map[int]bool{},
					b:    *new(map[int]bool),
					want: &diff.Value{},
				},
				{
					name: "maps not equal",
					a: map[int]bool{
						0: true,
						1: true,
						2: true,
					},
					b: map[int]bool{
						0: false,
						1: true,
						3: false,
					},
					want: &diff.Keys{
						Left: map[any]diff.Diff{
							0: &diff.Value{},
							2: &diff.Missing{},
						},
						Right: map[any]diff.Diff{
							0: &diff.Value{},
							3: &diff.Missing{},
						},
					},
				},
			},
		},
		{
			name: "pointers",
			tests: []test{
				{
					name: "pointer nils",
					a:    new(int),
					b:    new(int),
					want: nil,
				},
				{
					name: "pointers equals values",
					a:    ptr(2),
					b:    ptr(2),
					want: nil,
				},
				{
					name: "pointers different values",
					a:    ptr(1),
					b:    ptr(2),
					want: &diff.Value{},
				},
			},
		},
		{
			name: "structs",
			tests: []test{
				{
					name: "equal structs",
					a: test{
						name: "1",
						a:    2,
						b:    3,
						want: &diff.Value{},
					},
					b: test{
						name: "1",
						a:    2,
						b:    3,
						want: &diff.Value{},
					},
					want: nil,
				},
				{
					name: "structs not equal",
					a: sampleStruct{
						a: 1,
						b: "2",
						s: &sampleStruct{},
					},
					b: sampleStruct{
						a: 1,
						b: "2",
						s: &sampleStruct{
							a: 1,
							b: "2",
						},
					},
					want: &diff.Fields{
						Fields: map[string]diff.Diff{
							"s": &diff.Fields{
								Fields: map[string]diff.Diff{
									"a": &diff.Value{},
									"b": &diff.Value{},
								},
							},
						},
					},
				},
				{
					name: "nil and not nil pointer to a struct",
					a:    (*testSet)(nil),
					b: &testSet{
						name:  "",
						tests: nil,
					},
					want: &diff.Value{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, ttt := range tt.tests {
				t.Run(ttt.name, func(t *testing.T) {
					if ttt.panic {
						defer func() {
							if r := recover(); r != nil {
								t.Log("expected panic: \033[1m", r, "\033[0m")
								return
							}

							t.Error("panic was expected")
						}()
					}
					got := difference(reflect.ValueOf(ttt.a), reflect.ValueOf(ttt.b), false, walkSet{})
					if !reflect.DeepEqual(got, ttt.want) {
						t.Error("want\n", spew.Sdump(ttt.want), "\ngot\n", spew.Sdump(got))
					}
				})

			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
