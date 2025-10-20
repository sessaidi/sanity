package sanity

import (
	"math"
	"strings"
	"time"
)

func NotNilPtr[T any](name string, p *T) error {
	if p == nil {
		return NotNilError{Field: name}
	}
	return nil
}

func NonZero[T comparable](name string, v T) error {
	var zero T
	if v == zero {
		return NonZeroError{Field: name}
	}
	return nil
}

func NonEmpty(name, s string) error {
	if s == "" {
		return NonEmptyError{Field: name}
	}
	return nil
}

func NonBlank(name, s string) error {
	if len(strings.TrimSpace(s)) == 0 {
		return NonEmptyError{Field: name}
	}
	return nil
}

func StrLenAtLeast(name string, s string, n int) error {
	if len(s) < n {
		return LenAtLeastError{Field: name, Want: n, Got: len(s)}
	}
	return nil
}

func SliceLenAtLeast[T any](name string, s []T, n int) error {
	if len(s) < n {
		return LenAtLeastError{Field: name, Want: n, Got: len(s)}
	}
	return nil
}

func MapLenAtLeast[K comparable, V any](name string, m map[K]V, n int) error {
	if len(m) < n {
		return LenAtLeastError{Field: name, Want: n, Got: len(m)}
	}
	return nil
}

func InSet[T comparable](name string, v T, set map[T]struct{}) error {
	if _, ok := set[v]; !ok {
		return NotInSetError{Field: name}
	}
	return nil
}

func InRangeString(name, v, min, max string) error {
	if min > max {
		min, max = max, min
	}
	if v < min || v > max {
		return OutOfRangeError[string]{Field: name, Min: min, Max: max, Got: v}
	}
	return nil
}

func InRangeNum[T Numeric](name string, v, min, max T) error {
	if min > max {
		min, max = max, min
	}
	if v < min || v > max {
		return OutOfRangeError[T]{Field: name, Min: min, Max: max, Got: v}
	}
	return nil
}

func InRangeFloat64(name string, v, min, max float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < min || v > max {
		return OutOfRangeError[float64]{Field: name, Min: min, Max: max, Got: v}
	}
	return nil
}

func FiniteFloat64(name string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return OutOfRangeError[float64]{Field: name, Min: v, Max: v, Got: v}
	}
	return nil
}

func InRangeDuration(name string, v, min, max time.Duration) error {
	if min > max {
		min, max = max, min
	}
	if v < min || v > max {
		return OutOfRangeError[time.Duration]{Field: name, Min: min, Max: max, Got: v}
	}
	return nil
}
