package utils

import (
	"strconv"
)

func StringToUint8(s *string) *uint8 {
	b := []byte(*s)
	return &b[0]
}

func Float32SliceToString(floatArray []float32) string {
	outString := ""	
	for _, float := range(floatArray) {
		outString += strconv.FormatFloat(float64(float), 'f', 3, 32)
	}
	return outString
}
