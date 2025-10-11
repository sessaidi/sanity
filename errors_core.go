package sanity

import "errors"

// NotNilError indicates a nil pointer or nilable value.
type NotNilError struct {
	Field string
}

// NonZeroError indicates a zero-value where non-zero is required.
type NonZeroError struct {
	Field string
}

// NonEmptyError indicates an empty string where non-empty is required.
type NonEmptyError struct {
	Field string
}

// LenAtLeastError indicates len(value) < Want.
type LenAtLeastError struct {
	Field string
	Want  int
	Got   int
}

// OutOfRangeError indicates v ∉ [Min,Max] (inclusive).
type OutOfRangeError[T any] struct {
	Field    string
	Min, Max T
	Got      T
}

// NotInSetError indicates v ∉ allowed set.
type NotInSetError struct {
	Field string
}

// ---- Category sentinels (for errors.Is) ----
var (
	ErrNotNil     = errors.New("sanity:not_nil")
	ErrNonZero    = errors.New("sanity:non_zero")
	ErrNonEmpty   = errors.New("sanity:non_empty")
	ErrLenAtLeast = errors.New("sanity:len_at_least")
	ErrOutOfRange = errors.New("sanity:out_of_range")
	ErrNotInSet   = errors.New("sanity:not_in_set")
)

// ---- Introspection interfaces (for errors.As) ----

// FieldError exposes the logical field name causing the error.
type FieldError interface {
	error
	FieldName() string
}

// RangeError exposes bounds and offending value in a type-agnostic way.
type RangeError interface {
	error
	FieldName() string
	Bounds() (min any, max any)
	Value() any
}

// ---- Unwrap to category sentinels ----

func (e NotNilError) Unwrap() error {
	return ErrNotNil
}

func (e NonZeroError) Unwrap() error {
	return ErrNonZero
}

func (e NonEmptyError) Unwrap() error {
	return ErrNonEmpty
}

func (e LenAtLeastError) Unwrap() error {
	return ErrLenAtLeast
}

func (e NotInSetError) Unwrap() error {
	return ErrNotInSet
}

func (e OutOfRangeError[T]) Unwrap() error {
	return ErrOutOfRange
}

// ---- Field names ----

func (e NotNilError) FieldName() string {
	return e.Field
}

func (e NonZeroError) FieldName() string {
	return e.Field
}

func (e NonEmptyError) FieldName() string {
	return e.Field
}

func (e LenAtLeastError) FieldName() string {
	return e.Field
}

func (e NotInSetError) FieldName() string {
	return e.Field
}

func (e OutOfRangeError[T]) FieldName() string {
	return e.Field
}

// ---- Range details ----

func (e OutOfRangeError[T]) Bounds() (any, any) {
	return e.Min, e.Max
}

func (e OutOfRangeError[T]) Value() any {
	return e.Got
}
