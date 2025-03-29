package utils

import "slices"

func AnyOverlap[S ~[]E, E comparable](s1 S, s2 S) bool {
	return slices.ContainsFunc(
		s1,
		func(e E) bool {
			return slices.Contains(s2, e)
		},
	)
}
