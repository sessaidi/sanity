package sanity

import (
	"strings"
)

func SetIfZero[T comparable](p *T, def T) {
	var zero T
	if *p == zero {
		*p = def
	}
}

func SetIfNil[T any](p **T, def *T) {
	if *p == nil {
		*p = def
	}
}

func SetIfLE[T Numeric](p *T, limit, def T) {
	if *p <= limit {
		*p = def
	}
}

func SetIfLT[T Numeric](p *T, limit, def T) {
	if *p < limit {
		*p = def
	}
}

func SetIfGT[T Numeric](p *T, limit, def T) {
	if *p > limit {
		*p = def
	}
}

func SetIfGE[T Numeric](p *T, limit, def T) {
	if *p >= limit {
		*p = def
	}
}

func SetIfZeroThenClamp[T Numeric](p *T, def, min, max T) {
	if min > max {
		min, max = max, min
	}
	var zero T
	if *p == zero {
		*p = def
	}
	v := *p
	if v < min {
		*p = min
	} else if v > max {
		*p = max
	}
}

func Clamp[T Numeric](p *T, min, max T) {
	if min > max {
		min, max = max, min
	}
	v := *p
	if v < min {
		*p = min
	} else if v > max {
		*p = max
	}
}

func DefaultIf[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

func DefaultIfClamp[T Numeric](v, def, min, max T) T {
	if min > max {
		min, max = max, min
	}
	var zero T
	if v == zero {
		v = def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func InRange[T Numeric](v, min, max T) bool {
	if min > max {
		min, max = max, min
	}
	return v >= min && v <= max
}

func DefaultIfBlank(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
