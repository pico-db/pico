package utils

import "math"

// Rounds a floating point number to a precision/number of decimal places
func Round(num float64, precision int) float64 {
	coeff := math.Pow(10, float64(precision))
	return math.Round(num*coeff) / coeff // numeric truncation
}
