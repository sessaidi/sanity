package sanity_test

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"

	"github.com/sessaidi/sanity"
)

func TestFloatFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		function func() interface{}
		expected interface{}
	}{
		{
			name: "ZeroIfNaN(float64) -> 0",
			function: func() interface{} {
				return sanity.ZeroIfNaN[float64](math.NaN())
			},
			expected: float64(0),
		},
		{
			name: "ZeroIfNaN(float32) -> 0",
			function: func() interface{} {
				return sanity.ZeroIfNaN[float32](float32(math.NaN()))
			},
			expected: float32(0),
		},
		{
			name: "ZeroIfNaN with finite leaves value",
			function: func() interface{} {
				return sanity.ZeroIfNaN[float64](3.1415)
			},
			expected: 3.1415,
		},
		{
			name: "DefaultIfNaN uses default",
			function: func() interface{} {
				return sanity.DefaultIfNaN[float64](math.NaN(), 7)
			},
			expected: float64(7),
		},
		{
			name: "DefaultIfNaN leaves finite value",
			function: func() interface{} {
				return sanity.DefaultIfNaN[float64](2.5, 7)
			},
			expected: 2.5,
		},
		{
			name: "ClampFinite replaces NaN with default",
			function: func() interface{} {
				v := math.NaN()
				sanity.ClampFinite(&v, 0)
				return v
			},
			expected: float64(0),
		},
		{
			name: "ClampFinite replaces +Inf with default",
			function: func() interface{} {
				v := math.Inf(1)
				sanity.ClampFinite(&v, 0)
				return v
			},
			expected: float64(0),
		},
		{
			name: "ClampFinite replaces -Inf with default",
			function: func() interface{} {
				v := math.Inf(-1)
				sanity.ClampFinite(&v, 0)
				return v
			},
			expected: float64(0),
		},
		{
			name: "ClampFinite leaves finite value unchanged",
			function: func() interface{} {
				v := 1.25
				sanity.ClampFinite(&v, 0)
				return v
			},
			expected: 1.25,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.function()
			assert.Equal(t, tc.expected, got, "Failed %s: expected %v, got %v", tc.name, tc.expected, got)
		})
	}
}
