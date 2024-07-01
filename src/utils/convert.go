package utils

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func BoolToFloat(b bool) float64 {
	if b {
		return 1.
	}
	return 0.
}
