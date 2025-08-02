package util

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

// Abs returns the absolute value of a number
func Abs(number any) float64 {
	switch v := number.(type) {
	case int:
		return math.Abs(float64(v))
	case int64:
		return math.Abs(float64(v))
	case float32:
		return math.Abs(float64(v))
	case float64:
		return math.Abs(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return math.Abs(f)
		}
	}
	return 0
}

// Ceil returns the next highest integer value
func Ceil(value float64) float64 {
	return math.Ceil(value)
}

// Floor returns the next lowest integer value
func Floor(value float64) float64 {
	return math.Floor(value)
}

// Round returns the rounded value
func Round(val float64, precision ...int) float64 {
	prec := 0
	if len(precision) > 0 {
		prec = precision[0]
	}

	if prec == 0 {
		return math.Round(val)
	}

	shift := math.Pow(10, float64(prec))
	return math.Round(val*shift) / shift
}

// Min returns the lowest value
func Min(values ...any) any {
	if len(values) == 0 {
		return nil
	}

	minVal := values[0]
	minFloat, _ := convertToFloat64(minVal)

	for _, v := range values[1:] {
		if f, ok := convertToFloat64(v); ok && f < minFloat {
			minVal = v
			minFloat = f
		}
	}

	return minVal
}

// Max returns the highest value
func Max(values ...any) any {
	if len(values) == 0 {
		return nil
	}

	maxVal := values[0]
	maxFloat, _ := convertToFloat64(maxVal)

	for _, v := range values[1:] {
		if f, ok := convertToFloat64(v); ok && f > maxFloat {
			maxVal = v
			maxFloat = f
		}
	}

	return maxVal
}

// Pow returns base raised to the power of exp
func Pow(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// Sqrt returns the square root
func Sqrt(arg float64) float64 {
	return math.Sqrt(arg)
}

// Sin returns the sine
func Sin(arg float64) float64 {
	return math.Sin(arg)
}

// Cos returns the cosine
func Cos(arg float64) float64 {
	return math.Cos(arg)
}

// Tan returns the tangent
func Tan(arg float64) float64 {
	return math.Tan(arg)
}

// Asin returns the arc sine
func Asin(arg float64) float64 {
	return math.Asin(arg)
}

// Acos returns the arc cosine
func Acos(arg float64) float64 {
	return math.Acos(arg)
}

// Atan returns the arc tangent
func Atan(arg float64) float64 {
	return math.Atan(arg)
}

// Atan2 returns the arc tangent of y/x
func Atan2(y, x float64) float64 {
	return math.Atan2(y, x)
}

// Log returns the natural logarithm
func Log(arg float64, base ...float64) float64 {
	if len(base) > 0 {
		return math.Log(arg) / math.Log(base[0])
	}
	return math.Log(arg)
}

// Log10 returns the base-10 logarithm
func Log10(arg float64) float64 {
	return math.Log10(arg)
}

// Exp returns e raised to the power of arg
func Exp(arg float64) float64 {
	return math.Exp(arg)
}

// Deg2rad converts degrees to radians
func Deg2rad(number float64) float64 {
	return number * math.Pi / 180
}

// Rad2deg converts radians to degrees
func Rad2deg(number float64) float64 {
	return number * 180 / math.Pi
}

// Pi returns the value of pi
func Pi() float64 {
	return math.Pi
}

// Rand generates a random integer
func Rand(min ...int) int {
	if len(min) == 0 {
		return rand.Intn(math.MaxInt32)
	}
	if len(min) == 1 {
		return rand.Intn(min[0])
	}
	return min[0] + rand.Intn(min[1]-min[0]+1)
}

// Mt_rand generates a random value via the Mersenne Twister
func MtRand(min ...int) int {
	// Use the same implementation as Rand for simplicity
	return Rand(min...)
}

// Srand seeds the random number generator
func Srand(seed ...int64) {
	if len(seed) > 0 {
		rand.Seed(seed[0])
	} else {
		rand.Seed(time.Now().UnixNano())
	}
}

// Mt_srand seeds the Mersenne Twister
func MtSrand(seed ...int64) {
	Srand(seed...)
}

// Getrandmax returns the largest possible random value
func Getrandmax() int {
	return math.MaxInt32
}

// Mt_getrandmax returns the largest possible random value
func MtGetrandmax() int {
	return math.MaxInt32
}

// Lcg_value returns a pseudo random number
func LcgValue() float64 {
	return rand.Float64()
}

// Fmod returns the floating point remainder of x/y
func Fmod(x, y float64) float64 {
	return math.Mod(x, y)
}

// Hypot returns sqrt(x*x + y*y)
func Hypot(x, y float64) float64 {
	return math.Hypot(x, y)
}

// IsFinite finds whether a value is a legal finite number
func IsFinite(val float64) bool {
	return !math.IsInf(val, 0) && !math.IsNaN(val)
}

// IsInfinite finds whether a value is infinite
func IsInfinite(val float64) bool {
	return math.IsInf(val, 0)
}

// IsNan finds whether a value is not a number
func IsNan(val float64) bool {
	return math.IsNaN(val)
}

// Intval gets the integer value of a variable
func Intval(val any, base ...int) int {
	baseVal := 10
	if len(base) > 0 {
		baseVal = base[0]
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.ParseInt(v, baseVal, 64); err == nil {
			return int(i)
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int(f)
		}
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return 0
}

// Floatval gets the float value of a variable
func Floatval(val any) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	}
	return 0.0
}

// IsNumeric finds whether a variable is a number or a numeric string
func IsNumeric(val any) bool {
	switch v := val.(type) {
	case int, int64, float32, float64:
		return true
	case string:
		_, err1 := strconv.ParseInt(v, 10, 64)
		_, err2 := strconv.ParseFloat(v, 64)
		return err1 == nil || err2 == nil
	}
	return false
}

// Hexdec converts hexadecimal to decimal
func Hexdec(hexString string) int64 {
	result, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		return 0
	}
	return result
}

// Dechex converts decimal to hexadecimal
func Dechex(number int64) string {
	return strconv.FormatInt(number, 16)
}

// Octdec converts octal to decimal
func Octdec(octString string) int64 {
	result, err := strconv.ParseInt(octString, 8, 64)
	if err != nil {
		return 0
	}
	return result
}

// Decoct converts decimal to octal
func Decoct(number int64) string {
	return strconv.FormatInt(number, 8)
}

// Bindec converts binary to decimal
func Bindec(binaryString string) int64 {
	result, err := strconv.ParseInt(binaryString, 2, 64)
	if err != nil {
		return 0
	}
	return result
}

// Decbin converts decimal to binary
func Decbin(number int64) string {
	return strconv.FormatInt(number, 2)
}

// BaseConvert converts a number between arbitrary bases
func BaseConvert(number string, frombase, tobase int) string {
	// First convert to decimal
	decimal, err := strconv.ParseInt(number, frombase, 64)
	if err != nil {
		return "0"
	}

	// Then convert to target base
	return strconv.FormatInt(decimal, tobase)
}

// Gmp functions (simplified implementations)

// GmpAdd adds two numbers
func GmpAdd(a, b string) string {
	numA, errA := strconv.ParseInt(a, 10, 64)
	numB, errB := strconv.ParseInt(b, 10, 64)
	if errA != nil || errB != nil {
		return "0"
	}
	return strconv.FormatInt(numA+numB, 10)
}

// GmpSub subtracts two numbers
func GmpSub(a, b string) string {
	numA, errA := strconv.ParseInt(a, 10, 64)
	numB, errB := strconv.ParseInt(b, 10, 64)
	if errA != nil || errB != nil {
		return "0"
	}
	return strconv.FormatInt(numA-numB, 10)
}

// GmpMul multiplies two numbers
func GmpMul(a, b string) string {
	numA, errA := strconv.ParseInt(a, 10, 64)
	numB, errB := strconv.ParseInt(b, 10, 64)
	if errA != nil || errB != nil {
		return "0"
	}
	return strconv.FormatInt(numA*numB, 10)
}

// GmpDiv divides two numbers
func GmpDiv(a, b string) string {
	numA, errA := strconv.ParseInt(a, 10, 64)
	numB, errB := strconv.ParseInt(b, 10, 64)
	if errA != nil || errB != nil || numB == 0 {
		return "0"
	}
	return strconv.FormatInt(numA/numB, 10)
}

// GmpMod calculates modulo
func GmpMod(a, b string) string {
	numA, errA := strconv.ParseInt(a, 10, 64)
	numB, errB := strconv.ParseInt(b, 10, 64)
	if errA != nil || errB != nil || numB == 0 {
		return "0"
	}
	return strconv.FormatInt(numA%numB, 10)
}

// GmpPow raises number into power
func GmpPow(base string, exp int) string {
	numBase, err := strconv.ParseInt(base, 10, 64)
	if err != nil {
		return "0"
	}
	result := int64(1)
	for i := 0; i < exp; i++ {
		result *= numBase
	}
	return strconv.FormatInt(result, 10)
}

// Statistical functions

// ArraySum calculates the sum of values in an array (implemented in array.go but also useful here)
func ArraySumFloat(arr []float64) float64 {
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum
}

// ArrayMean calculates the mean (average) of values in an array
func ArrayMean(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	return ArraySumFloat(arr) / float64(len(arr))
}

// ArrayMedian calculates the median of values in an array
func ArrayMedian(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}

	// Sort the array first
	sorted := make([]float64, len(arr))
	copy(sorted, arr)

	// Simple bubble sort for demonstration
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// ArrayMode calculates the mode (most frequent value) of values in an array
func ArrayMode(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}

	frequency := make(map[float64]int)
	for _, v := range arr {
		frequency[v]++
	}

	var mode float64
	maxCount := 0
	for value, count := range frequency {
		if count > maxCount {
			maxCount = count
			mode = value
		}
	}

	return mode
}
