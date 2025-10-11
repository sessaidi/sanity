//go:build redact

package sanity_test

import (
	"strings"
	"testing"

	"github.com/sessaidi/sanity"
)

func TestTypedErrors_RedactedStrings(t *testing.T) {
	testCases := []struct {
		name     string
		function func() interface{}
		expected interface{}
	}{
		{
			name: "LenAtLeastError redacted omits 'got'",
			function: func() interface{} {
				err := sanity.LenAtLeastError{Field: "name", Want: 3, Got: 1}
				return !strings.Contains(err.Error(), "got")
			},
			expected: true,
		},
		{
			name: "OutOfRangeError redacted omits 'got'",
			function: func() interface{} {
				err := sanity.OutOfRangeError[int]{Field: "n", Min: 1, Max: 3, Got: 0}
				return !strings.Contains(err.Error(), "got")
			},
			expected: true,
		},
		{
			name: "NotNilError redacted string",
			function: func() interface{} {
				return sanity.NotNilError{Field: "client"}.Error()
			},
			expected: "client: must not be nil",
		},
		{
			name: "NonZeroError redacted string",
			function: func() interface{} {
				return sanity.NonZeroError{Field: "count"}.Error()
			},
			expected: "count: must be non-zero",
		},
		{
			name: "NonEmptyError redacted string",
			function: func() interface{} {
				return sanity.NonEmptyError{Field: "name"}.Error()
			},
			expected: "name: must be non-empty",
		},
		{
			name: "NotInSetError redacted string",
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
