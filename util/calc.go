package util

import (
	"math"
	"strconv"
	"strings"
)

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func CountDecimalPlaces(value float64) int {
	strValue := strconv.FormatFloat(value, 'f', -1, 64)

	parts := strings.Split(strValue, ".")

	if len(parts) < 2 {
		return 0
	}

	return len(parts[1])
}

func CountDecimalPlacesF32(value float32) int {
	strValue := strconv.FormatFloat(float64(value), 'f', -1, 64)

	parts := strings.Split(strValue, ".")

	if len(parts) < 2 {
		return 0
	}

	return len(parts[1])
}

func TruncateFloat64(val float64, precision int) float64 {
	prec := math.Pow(10, float64(precision))
	valInt := int64(val * prec)
	val = float64(valInt) / prec
	return val
}

func TruncateFloat32(val float32, precision int) float32 {
	prec := math.Pow(10, float64(precision))
	valInt := int64(float64(val) * prec)
	val = float32(float64(valInt) / prec)
	return val
}
