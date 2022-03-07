package deepequal_test

import (
	"github.com/sirkon/deepequal"
	"github.com/sirkon/deepequal/internal/testdata"
	"testing"
)

func TestEqual(t *testing.T) {
	type test struct {
		name string
		x    any
		y    any
		want bool
	}

	type hasProtoMessage struct {
		sample *testdata.Sample
	}

	tests := []test{
		{
			name: "simple match",
			x:    map[string]int{},
			y:    map[string]int{},
			want: true,
		},
		{
			name: "simple mismatch",
			x:    map[string]string{},
			y:    map[string]int{},
			want: false,
		},
		{
			name: "direct proto match",
			x: &testdata.Sample{
				Str: "str",
				Sub: &testdata.Sub{
					Val: 1,
				},
			},
			y: &testdata.Sample{
				Str: "str",
				Sub: &testdata.Sub{
					Val: 1,
				},
			},
			want: true,
		},
		{
			name: "direct proto mismatch",
			x: &testdata.Sample{
				Str: "str",
				Sub: nil,
			},
			y: testdata.Sample{
				Str: "str",
				Sub: &testdata.Sub{
					Val: 2,
				},
			},
			want: false,
		},
		{
			name: "proto field match",
			x: hasProtoMessage{
				sample: &testdata.Sample{
					Str: "123",
				},
			},
			y: hasProtoMessage{
				sample: &testdata.Sample{
					Str: "123",
				},
			},
			want: true,
		},
		{
			name: "proto field mismatch",
			x: hasProtoMessage{
				sample: &testdata.Sample{
					Str: "123",
				},
			},
			y: hasProtoMessage{
				sample: &testdata.Sample{
					Str: "1234",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if deepequal.Equal(tt.x, tt.y) != tt.want {
				t.Errorf("unexpected mismatch between %#v and %#v", tt.x, tt.y)
			}
		})
	}
}
