package sanity_test

import (
	"testing"
	"time"

	"github.com/sessaidi/sanity"
)

func TestDurationFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		function func() interface{}
		expected interface{}
	}{
		{
			name: "ClampDuration in-range stays same",
			function: func() interface{} {
				d := 1500 * time.Millisecond
				sanity.ClampDuration(&d, 1*time.Second, 2*time.Second)
				return d
			},
			expected: 1500 * time.Millisecond,
		},
		{
			name: "ClampDuration below min -> clamped to min",
			function: func() interface{} {
				d := 200 * time.Millisecond
				sanity.ClampDuration(&d, 1*time.Second, 2*time.Second)
				return d
			},
			expected: 1 * time.Second,
		},
		{
			name: "ClampDuration above max -> clamped to max",
			function: func() interface{} {
				d := 3 * time.Second
				sanity.ClampDuration(&d, 1*time.Second, 2*time.Second)
				return d
			},
			expected: 2 * time.Second,
		},
		{
			name: "ClampDuration with swapped bounds (min>max) still clamps correctly",
			function: func() interface{} {
				d := 500 * time.Millisecond
				sanity.ClampDuration(&d, 2*time.Second, 1*time.Second) // swapped internally to [1s,2s]
				return d
			},
			expected: 1 * time.Second,
		},
		{
			name: "DefaultDurationClamp zero -> default (in-range)",
			function: func() interface{} {
				return sanity.DefaultDurationClamp(
					0,
					2*time.Second,
					500*time.Millisecond,
					10*time.Second,
				)
			},
			expected: 2 * time.Second,
		},
		{
			name: "DefaultDurationClamp non-zero in-range stays same",
			function: func() interface{} {
				return sanity.DefaultDurationClamp(
					1500*time.Millisecond,
					2*time.Second,
					1*time.Second,
					2*time.Second,
				)
			},
			expected: 1500 * time.Millisecond,
		},
		{
			name: "DefaultDurationClamp non-zero below min -> clamped to min",
			function: func() interface{} {
				return sanity.DefaultDurationClamp(
					200*time.Millisecond,
					2*time.Second,
					1*time.Second,
					2*time.Second,
				)
			},
			expected: 1 * time.Second,
		},
		{
			name: "DefaultDurationClamp with swapped bounds (min>max) zero -> default then checked in [max,min]",
			function: func() interface{} {
				// bounds swapped to [50ms, 200ms]; zero -> default (100ms) which is in-range
				return sanity.DefaultDurationClamp(
					0,
					100*time.Millisecond,
					200*time.Millisecond,
					50*time.Millisecond,
				)
			},
			expected: 100 * time.Millisecond,
		},
		{
			name: "DefaultDurationClamp non-zero above max -> clamped to max",
			function: func() interface{} {
				return sanity.DefaultDurationClamp(
					5*time.Second,
					2*time.Second,
					500*time.Millisecond,
					2*time.Second,
				)
			},
			expected: 2 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.function()
			if result != tc.expected {
				t.Errorf("Failed %s: expected %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}
