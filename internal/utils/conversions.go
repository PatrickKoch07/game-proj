package utils

func StringToUint8(s *string) *uint8 {
	b := []byte(*s)
	return &b[0]
}
