package deepequal_test

import (
	"testing"

	"github.com/sirkon/deepequal"
	"github.com/sirkon/deepequal/internal/testdata"
)

func TestSideBySide(t *testing.T) {
	deepequal.SideBySide(t, "integers", 1, 2)
	deepequal.SideBySide(t, "slices", []int{1, 2, 3, 4, 5, 6}, []int{3, 6, 7})
	deepequal.SideBySide(
		t,
		"maps",
		map[string]int{
			"one": 1,
			"two": 2,
		},
		map[string]int{
			"one":   2,
			"two":   2,
			"three": 3,
		},
	)
	deepequal.SideBySide(
		t,
		"structs",
		&testdata.Sample{
			Str: "abcdef",
			Sub: &testdata.Sub{
				Val: 12,
			},
		},
		&testdata.Sample{
			Str: "abcdef",
			Sub: &testdata.Sub{
				Val: 14,
			},
		},
	)
}
