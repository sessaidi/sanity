package sanity

import "math"

type Float interface {
	~float32 | ~float64
}

func ZeroIfNaN[T Float](v T) T {
	if math.IsNaN(float64(v)) {
		var z T
		return z
	}
	return v
}

func DefaultIfNaN[T Float](v, def T) T {
	if math.IsNaN(float64(v)) {
		return def
	}
	return v
}

func ClampFinite[T Float](p *T, def T) {
	f := float64(*p)
	if math.IsNaN(f) || math.IsInf(f, 0) {
		*p = def
	}
}
