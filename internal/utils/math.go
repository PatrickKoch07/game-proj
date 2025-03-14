package utils

import "golang.org/x/exp/constraints"

func Clamp[T constraints.Ordered](value T, mini T, maxi T) T {
	return min(max(value, mini), maxi)
}
