package utils

import (
	"math"

	"golang.org/x/exp/constraints"
)

func Clamp[T constraints.Ordered](value T, mini T, maxi T) T {
	return min(max(value, mini), maxi)
}

func Ceil32(value float32) float32 {
	return float32(math.Ceil(float64(value)))
}
