package sanity

import "errors"

// GroupAsSlice appends underlying errors into dst and returns the result.
func GroupAsSlice(err error, dst []error) []error {
	if err == nil {
		return dst
	}
	var eg ErrorGroup
	if errors.As(err, &eg) {
		eg.Iter(func(e error) bool {
			dst = append(dst, e)
			return true
		})
		return dst
	}
	return append(dst, err)
}

// GroupLen reports the number of underlying errors if err is this package's group.
func GroupLen(err error) (int, bool) {
	type hasLen interface{ Len() int }
	if hg, ok := err.(hasLen); ok {
		return hg.Len(), true
	}
	return 0, false
}
