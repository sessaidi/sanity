package sanity_test

import (
	"github.com/sessaidi/sanity"
	"testing"
)

func TestPointerFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		function func() interface{}
		expected interface{}
	}{
		{
			"P with non-nil string",
			func() interface{} {
				s := "hello"
				return sanity.P(&s)
			},
			"hello",
		},
		{
			"P with nil string",
			func() interface{} {
				var s *string
				return sanity.P(s)
			},
			"",
		},
		{
			"P with non-nil int",
			func() interface{} {
				i := 42
				return sanity.P(&i)
			},
			42,
		},
		{
			"P with nil int",
			func() interface{} {
				var i *int
				return sanity.P(i)
			},
			0,
		},
		{
			"Ptr to string",
			func() interface{} {
				s := "hello"
				return *sanity.Ptr(s)
			},
			"hello",
		},
		{
			"Ptr to int",
			func() interface{} {
				i := 42
				return *sanity.Ptr(i)
			},
			42,
		},
		{
			"POrDefault with non-nil string",
			func() interface{} {
				s := "hello"
				return sanity.POrDefault(&s, "default")
			},
			"hello",
		},
		{
			"POrDefault with nil string",
			func() interface{} {
				var s *string
				return sanity.POrDefault(s, "default")
			},
			"default",
		},
		{
			"POrDefault with non-nil int",
			func() interface{} {
				i := 42
				return sanity.POrDefault(&i, 100)
			},
			42,
		},
		{
			"POrDefault with nil int",
			func() interface{} {
				var i *int
				return sanity.POrDefault(i, 100)
			},
			100,
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
