//go:build !redact

package sanity_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/sessaidi/sanity"
)

func TestTypedErrors_Is_As_Verbose(t *testing.T) {
	testCases := []struct {
		name     string
		function func() interface{}
		expected interface{}
	}{
		{
			name: "NotNilError matches ErrNotNil and exposes Field",
			function: func() interface{} {
				err := sanity.NotNilError{Field: "client"}
				okIs := errors.Is(err, sanity.ErrNotNil)
				var fe sanity.FieldError
				okAs := errors.As(err, &fe) && fe.FieldName() == "client"
				return okIs && okAs
			},
			expected: true,
		},
		{
			name: "OutOfRangeError matches ErrOutOfRange and exposes bounds/value",
			function: func() interface{} {
				err := sanity.OutOfRangeError[int]{Field: "port", Min: 1, Max: 10, Got: 0}
				okIs := errors.Is(err, sanity.ErrOutOfRange)
				var re sanity.RangeError
				okAs := errors.As(err, &re)
				min, max := re.Bounds()
				got := re.Value()
				return okIs && okAs && min == 1 && max == 10 && got == 0
			},
			expected: true,
		},
		{
			name: "LenAtLeastError verbose contains 'got'",
			function: func() interface{} {
				err := sanity.LenAtLeastError{Field: "name", Want: 3, Got: 1}
				return strings.Contains(err.Error(), "got 1")
			},
			expected: true,
		},
		{
			name: "OutOfRangeError verbose contains 'got'",
			function: func() interface{} {
				err := sanity.OutOfRangeError[int]{Field: "n", Min: 1, Max: 3, Got: 0}
				return strings.Contains(err.Error(), "got 0")
			},
			expected: true,
		},
		{
			name: "NotNilError verbose string",
			function: func() interface{} {
				return sanity.NotNilError{Field: "client"}.Error()
			},
			expected: "client: must not be nil",
		},
		{
			name: "NonZeroError verbose string",
			function: func() interface{} {
				return sanity.NonZeroError{Field: "count"}.Error()
			},
			expected: "count: must be non-zero",
		},
		{
			name: "NonEmptyError verbose string",
			function: func() interface{} {
				return sanity.NonEmptyError{Field: "name"}.Error()
			},
			expected: "name: must be non-empty",
		},
		{
			name: "NotInSetError verbose string",
			function: func() interface{} {
				return sanity.NotInSetError{Field: "mode"}.Error()
			},
			expected: "mode: invalid value",
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
