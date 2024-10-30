package goiforest

import (
	"fmt"
	"testing"
)

func TestKurtosis(t *testing.T) {
	cases := []struct {
		values   []float64
		expected float64
	}{
		{[]float64{1, 2, 3, 4, 5}, -1.2},
		{[]float64{100, 102, 98, 101, 99, 97, 103, 101, 98, 100}, -0.87},
		{[]float64{1, 2, 2, 3, 3, 3, 3, 4, 4, 5}, 0},
	}

	withinTolerance := func(a, b, tolerance float64) bool {
		return a > b-tolerance && a < b+tolerance
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.values), func(t *testing.T) {
			actual := kurtosis(c.values)
			if !withinTolerance(actual, c.expected, 0.1) {
				t.Errorf("Expected %f, but got %f", c.expected, actual)
			}
		})
	}
}
