package deepequal_test

import (
	"fmt"
	"testing"

	"github.com/sirkon/deepequal"
	"github.com/sirkon/deepequal/internal/testdata"
)

func TestSideBySide(t *testing.T) {
	deepequal.SideBySide(quasiTesting{}, "integers", 1, 2)
	deepequal.SideBySide(
		quasiTesting{},
		"slices",
		[]int{1, 2, 3, 4, 5, 6},
		[]int{3, 6, 7},
	)
	deepequal.SideBySide(
		quasiTesting{},
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
		quasiTesting{},
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

type quasiTesting struct{}

func (q quasiTesting) Helper() {
	return
}

func (q quasiTesting) Log(a ...any) {
	fmt.Print(a...)
}

func (q quasiTesting) Error(a ...any) {
	fmt.Print(a...)
}
