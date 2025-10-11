package sanity

// P dereferences the pointer to a value of type T. It returns the value pointed to
// if the pointer is not nil. If the pointer is nil, it returns the zero value of type T.
func P[T any](ptr *T) T {
	if ptr != nil {

		return *ptr
	}
	var zero T
	return zero
}

// Ptr creates and returns a pointer to a value of type T.
func Ptr[T any](value T) *T {
	return &value
}

// POrDefault dereferences the pointer to a value of type T. If the pointer is nil,
// it returns a specified default value instead of the zero value.
func POrDefault[T any](ptr *T, defaultVal T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
