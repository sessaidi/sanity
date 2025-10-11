package sanity_test

import (
	"github.com/sessaidi/sanity"
	"testing"
)

func TestSetIfZero(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "int zero -> set default",
			fn: func() interface{} {
				v := 0
				sanity.SetIfZero(&v, 42)
				return v
			},
			expected: 42,
		},
		{
			name: "int non-zero -> unchanged",
			fn: func() interface{} {
				v := 7
				sanity.SetIfZero(&v, 42)
				return v
			},
			expected: 7,
		},
		{
			name: "string zero -> set default",
			fn: func() interface{} {
				s := ""
				sanity.SetIfZero(&s, "def")
				return s
			},
			expected: "def",
		},
		// Note: allowed but generally avoid for bool unless false truly means "unset".
		{
			name: "bool false -> set default true",
			fn: func() interface{} {
				b := false
				sanity.SetIfZero(&b, true)
				return b
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestSetIfNil(t *testing.T) {
	type C struct{ N int }

	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "nil -> set default pointer",
			fn: func() interface{} {
				var p *C
				def := &C{N: 3}
				sanity.SetIfNil(&p, def)
				return p.N
			},
			expected: 3,
		},
		{
			name: "non-nil -> unchanged",
			fn: func() interface{} {
				p := &C{N: 7}
				def := &C{N: 3}
				sanity.SetIfNil(&p, def)
				return p.N
			},
			expected: 7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestSetIfComparators(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "SetIfLE: <= limit -> default",
			fn: func() interface{} {
				x := 0
				sanity.SetIfLE(&x, 0, 10)
				return x
			},
			expected: 10,
		},
		{
			name: "SetIfLE: > limit -> unchanged",
			fn: func() interface{} {
				y := 5
				sanity.SetIfLE(&y, 0, 10)
				return y
			},
			expected: 5,
		},
		{
			name: "SetIfLT: < limit -> default",
			fn: func() interface{} {
				v := 5
				sanity.SetIfLT(&v, 10, 20)
				return v
			},
			expected: 20,
		},
		{
			name: "SetIfGT: > limit -> default",
			fn: func() interface{} {
				v := 100
				sanity.SetIfGT(&v, 50, 7)
				return v
			},
			expected: 7,
		},
		{
			name: "SetIfGE: >= limit -> default",
			fn: func() interface{} {
				v := 9
				sanity.SetIfGE(&v, 9, 1)
				return v
			},
			expected: 1,
		},
		{
			name: "SetIfGE: < limit -> unchanged",
			fn: func() interface{} {
				v := 8
				sanity.SetIfGE(&v, 9, 1)
				return v
			},
			expected: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "in-range -> unchanged",
			fn: func() interface{} {
				x := 5
				sanity.Clamp(&x, 1, 10)
				return x
			},
			expected: 5,
		},
		{
			name: "below min -> clamped to min",
			fn: func() interface{} {
				x := -3
				sanity.Clamp(&x, 1, 10)
				return x
			},
			expected: 1,
		},
		{
			name: "above max -> clamped to max",
			fn: func() interface{} {
				x := 99
				sanity.Clamp(&x, 1, 10)
				return x
			},
			expected: 10,
		},
		{
			name: "swapped bounds (min>max) -> swap then clamp up",
			fn: func() interface{} {
				x := -3
				sanity.Clamp(&x, 10, 1) // becomes [1,10]
				return x
			},
			expected: 1,
		},
		{
			name: "float clamp in-range",
			fn: func() interface{} {
				f := 1.5
				sanity.Clamp(&f, 1.0, 2.0)
				return f
			},
			expected: 1.5,
		},
		{
			name: "float clamp above max",
			fn: func() interface{} {
				f := 3.14
				sanity.Clamp(&f, 0.0, 1.0)
				return f
			},
			expected: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestSetIfZeroThenClamp(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "zero -> default then clamp (def in-range)",
			fn: func() interface{} {
				v := 0
				sanity.SetIfZeroThenClamp(&v, 5, 1, 10)
				return v
			},
			expected: 5,
		},
		{
			name: "zero -> default then clamp (def above max -> clamped to max)",
			fn: func() interface{} {
				v := 0
				sanity.SetIfZeroThenClamp(&v, 99, 1, 10)
				return v
			},
			expected: 10,
		},
		{
			name: "non-zero below min -> clamped to min",
			fn: func() interface{} {
				v := -5
				sanity.SetIfZeroThenClamp(&v, 100, 1, 10)
				return v
			},
			expected: 1,
		},
		{
			name: "non-zero in-range -> unchanged",
			fn: func() interface{} {
				v := 7
				sanity.SetIfZeroThenClamp(&v, 100, 1, 10)
				return v
			},
			expected: 7,
		},
		{
			name: "swapped bounds -> swap then apply (zero -> def -> clamp)",
			fn: func() interface{} {
				v := 0
				sanity.SetIfZeroThenClamp(&v, 100, 10, 1) // swap -> [1,10]
				return v
			},
			expected: 10,
		},
		{
			name: "float case: zero -> default then clamp",
			fn: func() interface{} {
				f := 0.0
				sanity.SetIfZeroThenClamp(&f, 2.5, 1.0, 3.0)
				return f
			},
			expected: 2.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestDefaultIf(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "int zero -> default",
			fn: func() interface{} {
				return sanity.DefaultIf(0, 42)
			},
			expected: 42,
		},
		{
			name: "int non-zero -> unchanged",
			fn: func() interface{} {
				return sanity.DefaultIf(7, 42)
			},
			expected: 7,
		},
		{
			name: "string zero -> default",
			fn: func() interface{} {
				return sanity.DefaultIf("", "def")
			},
			expected: "def",
		},
		{
			name: "string non-zero -> unchanged",
			fn: func() interface{} {
				return sanity.DefaultIf("ok", "def")
			},
			expected: "ok",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestDefaultIfClamp(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "zero -> default then clamp to [1,3] -> 3",
			fn: func() interface{} {
				return sanity.DefaultIfClamp(0, 5, 1, 3)
			},
			expected: 3,
		},
		{
			name: "in-range -> unchanged",
			fn: func() interface{} {
				return sanity.DefaultIfClamp(2, 5, 1, 3)
			},
			expected: 2,
		},
		{
			name: "above max -> clamped to max",
			fn: func() interface{} {
				return sanity.DefaultIfClamp(9, 5, 1, 3)
			},
			expected: 3,
		},
		{
			name: "swapped bounds -> swap then apply",
			fn: func() interface{} {
				return sanity.DefaultIfClamp(0, 100, 10, 1)
			},
			expected: 10,
		},
		{
			name: "float case: zero -> default in-range",
			fn: func() interface{} {
				return sanity.DefaultIfClamp(0.0, 2.5, 1.0, 3.0)
			},
			expected: 2.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestInRange(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "in-range",
			fn: func() interface{} {
				return sanity.InRange(5, 1, 10)
			},
			expected: true,
		},
		{
			name: "below min",
			fn: func() interface{} {
				return sanity.InRange(-1, 1, 10)
			},
			expected: false,
		},
		{
			name: "swap bounds then in-range",
			fn: func() interface{} {
				return sanity.InRange(5, 10, 1)
			},
			expected: true,
		},
		{
			name: "inclusive min",
			fn: func() interface{} {
				return sanity.InRange(1, 1, 10)
			},
			expected: true,
		},
		{
			name: "inclusive max",
			fn: func() interface{} {
				return sanity.InRange(10, 1, 10)
			},
			expected: true,
		},
		{
			name: "float in-range",
			fn: func() interface{} {
				return sanity.InRange(1.5, 1.0, 2.0)
			},
			expected: true,
		},
		{
			name: "float out-of-range",
			fn: func() interface{} {
				return sanity.InRange(2.5, 1.0, 2.0)
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestDefaultIfBlank(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func() interface{}
		expected interface{}
	}{
		{
			name: "empty string -> default",
			fn: func() interface{} {
				return sanity.DefaultIfBlank("", "def")
			},
			expected: "def",
		},
		{
			name: "whitespace string -> default",
			fn: func() interface{} {
				return sanity.DefaultIfBlank(" \t\n", "def")
			},
			expected: "def",
		},
		{
			name: "non-blank string -> unchanged",
			fn: func() interface{} {
				return sanity.DefaultIfBlank(" ok ", "def")
			},
			expected: " ok ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
