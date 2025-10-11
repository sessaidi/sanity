//go:build redact

package sanity

import "fmt"

func (e NotNilError) Error() string {
	return e.FieldName() + ": must not be nil"
}

func (e NonZeroError) Error() string {
	return e.FieldName() + ": must be non-zero"
}

func (e NonEmptyError) Error() string {
	return e.FieldName() + ": must be non-empty"
}

func (e NotInSetError) Error() string {
	return e.FieldName() + ": invalid value"
}

func (e LenAtLeastError) Error() string {
	return fmt.Sprintf("%s: len must be >= %d", e.FieldName(), e.Want)
}

func (e OutOfRangeError[T]) Error() string {
	return fmt.Sprintf("%s: must be in [%v,%v]", e.FieldName(), e.Min, e.Max)
}
