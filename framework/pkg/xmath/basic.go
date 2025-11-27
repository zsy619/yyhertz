package xmath

import (
	"math"
	"strconv"
	"strings"
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

// convertToFloat64 helper function
func convertToFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// Statistical functions
func ArraySumFloat(arr []float64) float64 {
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum
}

func ArrayMean(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	return ArraySumFloat(arr) / float64(len(arr))
}

// 数学运算函数
func Add(a, b any) any {
	return ToFloat64(a) + ToFloat64(b)
}

func Sub(a, b any) any {
	return ToFloat64(a) - ToFloat64(b)
}

func Mul(a, b any) any {
	return ToFloat64(a) * ToFloat64(b)
}

func Div(a, b any) any {
	bVal := ToFloat64(b)
	if bVal == 0 {
		return 0
	}
	return ToFloat64(a) / bVal
}

func Mod(a, b any) any {
	return int(ToFloat64(a)) % int(ToFloat64(b))
}

func ToInt(v any) int {
	return int(ToFloat64(v))
}

// 辅助函数
func ToFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

// ToInt64 转换为int64
func ToInt64(v any) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

// ToBool 转换为bool
func ToBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
		return val != ""
	}
	return false
}

// NumberFormat formats a number with grouped thousands
func NumberFormat(number float64, decimals int, decPoint, thousandsSep string) string {
	formatted := strconv.FormatFloat(number, 'f', decimals, 64)
	parts := strings.Split(formatted, ".")

	// Add thousands separator
	intPart := parts[0]
	if len(intPart) > 3 {
		var result []rune
		for i, r := range []rune(intPart) {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result = append(result, []rune(thousandsSep)...)
			}
			result = append(result, r)
		}
		intPart = string(result)
	}

	if decimals > 0 && len(parts) > 1 {
		return intPart + decPoint + parts[1]
	}
	return intPart
}
