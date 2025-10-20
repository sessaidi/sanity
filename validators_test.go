package sanity_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/sessaidi/sanity"
)

func TestValidators(t *testing.T) {
	testCases := []struct {
		name     string
		function func() interface{}
		expected interface{}
	}{
		// presence / content
		{
			name: "NotNilPtr nil -> ErrNotNil",
			function: func() interface{} {
				return errors.Is(sanity.NotNilPtr[int]("p", nil), sanity.ErrNotNil)
			},
			expected: true,
		},
		{
			name: "NotNilPtr valid -> nil",
			function: func() interface{} {
				p := new(int)
				return errors.Is(sanity.NotNilPtr[int]("p", p), sanity.ErrNotNil)
			},
			expected: false,
		},
		{
			name: "NonZero zero -> ErrNonZero",
			function: func() interface{} {
				return errors.Is(sanity.NonZero("n", 0), sanity.ErrNonZero)
			},
			expected: true,
		},
		{
			name: "NonZero valid -> nil",
			function: func() interface{} {
				return errors.Is(sanity.NonZero("n", 1), sanity.ErrNonZero)
			},
			expected: false,
		},
		{
			name: "NonEmpty empty -> ErrNonEmpty",
			function: func() interface{} {
				return errors.Is(sanity.NonEmpty("s", ""), sanity.ErrNonEmpty)
			},
			expected: true,
		},
		{
			name: "NonEmpty valid -> nil",
			function: func() interface{} {
				return errors.Is(sanity.NonEmpty("s", "valid"), sanity.ErrNonEmpty)
			},
			expected: false,
		},
		{
			name: "NonBlank blanks -> ErrNonEmpty",
			function: func() interface{} {
				return errors.Is(sanity.NonBlank("s", "   "), sanity.ErrNonEmpty)
			},
			expected: true,
		},
		{
			name: "NonBlank valid -> nil",
			function: func() interface{} {
				return errors.Is(sanity.NonBlank("s", "   valid"), sanity.ErrNonEmpty)
			},
			expected: false,
		},
		{
			name: "NonEmpty ok -> nil",
			function: func() interface{} {
				return sanity.NonEmpty("s", "x") == nil
			},
			expected: true,
		},
		{
			name: "StrLenAtLeast short -> ErrLenAtLeast",
			function: func() interface{} {
				return errors.Is(sanity.StrLenAtLeast("name", "ab", 3), sanity.ErrLenAtLeast)
			},
			expected: true,
		},
		{
			name: "SliceLenAtLeast short -> ErrLenAtLeast",
			function: func() interface{} {
				return errors.Is(sanity.SliceLenAtLeast("xs", []int{1}, 2), sanity.ErrLenAtLeast)
			},
			expected: true,
		},
		{
			name: "MapLenAtLeast short -> ErrLenAtLeast",
			function: func() interface{} {
				return errors.Is(sanity.MapLenAtLeast("m", map[int]int{1: 1}, 2), sanity.ErrLenAtLeast)
			},
			expected: true,
		},
		{
			name: "InSet miss -> ErrNotInSet",
			function: func() interface{} {
				set := map[string]struct{}{"auto": {}, "manual": {}}
				return errors.Is(sanity.InSet("mode", "x", set), sanity.ErrNotInSet)
			},
			expected: true,
		},
		{
			name: "InSet hit -> nil",
			function: func() interface{} {
				set := map[string]struct{}{"auto": {}, "manual": {}}
				return sanity.InSet("mode", "auto", set) == nil
			},
			expected: true,
		},
		{
			name: "InRangeNum below -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.InRangeNum("n", 0, 1, 10), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeNum swap -> ok",
			function: func() interface{} {
				return sanity.InRangeNum("n", 5, 10, 1) == nil
			},
			expected: true,
		},
		{
			name: "InRange string lexicographic ok",
			function: func() interface{} {
				return sanity.InRangeString("s", "b", "a", "c") == nil
			},
			expected: true,
		},
		{
			name: "InRange string lexicographic miss",
			function: func() interface{} {
				return errors.Is(sanity.InRangeString("s", "z", "a", "c"), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeFloat64 ok in [0,1]",
			function: func() interface{} {
				return sanity.InRangeFloat64("x", 0.75, 0.0, 1.0) == nil
			},
			expected: true,
		},
		{
			name: "InRangeFloat64 below min -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.InRangeFloat64("x", -0.1, 0.0, 1.0), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeFloat64 above max -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.InRangeFloat64("x", 1.1, 0.0, 1.0), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeFloat64 NaN -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.InRangeFloat64("x", math.NaN(), 0.0, 1.0), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeFloat64 +Inf -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.InRangeFloat64("x", math.Inf(+1), 0.0, 1.0), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeFloat64 -Inf -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.InRangeFloat64("x", math.Inf(-1), 0.0, 1.0), sanity.ErrOutOfRange)
			},
			expected: true,
		},

		{
			name: "FiniteFloat64 finite -> nil",
			function: func() interface{} {
				return sanity.FiniteFloat64("x", 123.0) == nil
			},
			expected: true,
		},
		{
			name: "FiniteFloat64 NaN -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.FiniteFloat64("x", math.NaN()), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "FiniteFloat64 +Inf -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.FiniteFloat64("x", math.Inf(+1)), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "FiniteFloat64 -Inf -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.FiniteFloat64("x", math.Inf(-1)), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeDuration below -> ErrOutOfRange",
			function: func() interface{} {
				return errors.Is(sanity.InRangeDuration("d", 500*time.Millisecond, time.Second, 2*time.Second), sanity.ErrOutOfRange)
			},
			expected: true,
		},
		{
			name: "InRangeDuration ok",
			function: func() interface{} {
				return sanity.InRangeDuration("d", 1500*time.Millisecond, time.Second, 2*time.Second) == nil
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.function()
			if got != tc.expected {
				t.Errorf("Failed %s: expected %v, got %v", tc.name, tc.expected, got)
			}
		})
	}
}
