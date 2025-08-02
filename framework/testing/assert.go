package testing

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Assert 断言工具
type Assert struct {
	t *testing.T
}

// NewAssert 创建断言工具
func NewAssert(t *testing.T) *Assert {
	return &Assert{t: t}
}

// ============= 基础断言 =============

// True 断言为真
func (a *Assert) True(condition bool, msgAndArgs ...any) {
	if !condition {
		a.t.Helper()
		a.fail("Expected true", msgAndArgs...)
	}
}

// False 断言为假
func (a *Assert) False(condition bool, msgAndArgs ...any) {
	if condition {
		a.t.Helper()
		a.fail("Expected false", msgAndArgs...)
	}
}

// Equal 断言相等
func (a *Assert) Equal(expected, actual any, msgAndArgs ...any) {
	if !isEqual(expected, actual) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %#v, got %#v", expected, actual), msgAndArgs...)
	}
}

// NotEqual 断言不相等
func (a *Assert) NotEqual(expected, actual any, msgAndArgs ...any) {
	if isEqual(expected, actual) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected values to be different, but both were %#v", expected), msgAndArgs...)
	}
}

// Nil 断言为nil
func (a *Assert) Nil(value any, msgAndArgs ...any) {
	if !isNil(value) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected nil, got %#v", value), msgAndArgs...)
	}
}

// NotNil 断言不为nil
func (a *Assert) NotNil(value any, msgAndArgs ...any) {
	if isNil(value) {
		a.t.Helper()
		a.fail("Expected non-nil value", msgAndArgs...)
	}
}

// Empty 断言为空
func (a *Assert) Empty(value any, msgAndArgs ...any) {
	if !isEmpty(value) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected empty, got %#v", value), msgAndArgs...)
	}
}

// NotEmpty 断言不为空
func (a *Assert) NotEmpty(value any, msgAndArgs ...any) {
	if isEmpty(value) {
		a.t.Helper()
		a.fail("Expected non-empty value", msgAndArgs...)
	}
}

// Zero 断言为零值
func (a *Assert) Zero(value any, msgAndArgs ...any) {
	if !isZero(value) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected zero value, got %#v", value), msgAndArgs...)
	}
}

// NotZero 断言不为零值
func (a *Assert) NotZero(value any, msgAndArgs ...any) {
	if isZero(value) {
		a.t.Helper()
		a.fail("Expected non-zero value", msgAndArgs...)
	}
}

// ============= 字符串断言 =============

// Contains 断言包含子字符串
func (a *Assert) Contains(haystack, needle string, msgAndArgs ...any) {
	if !strings.Contains(haystack, needle) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected '%s' to contain '%s'", haystack, needle), msgAndArgs...)
	}
}

// NotContains 断言不包含子字符串
func (a *Assert) NotContains(haystack, needle string, msgAndArgs ...any) {
	if strings.Contains(haystack, needle) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected '%s' to not contain '%s'", haystack, needle), msgAndArgs...)
	}
}

// HasPrefix 断言有前缀
func (a *Assert) HasPrefix(str, prefix string, msgAndArgs ...any) {
	if !strings.HasPrefix(str, prefix) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected '%s' to have prefix '%s'", str, prefix), msgAndArgs...)
	}
}

// HasSuffix 断言有后缀
func (a *Assert) HasSuffix(str, suffix string, msgAndArgs ...any) {
	if !strings.HasSuffix(str, suffix) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected '%s' to have suffix '%s'", str, suffix), msgAndArgs...)
	}
}

// Matches 断言匹配正则表达式
func (a *Assert) Matches(pattern, str string, msgAndArgs ...any) {
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		a.t.Helper()
		a.fail(fmt.Sprintf("Invalid regex pattern '%s': %v", pattern, err), msgAndArgs...)
		return
	}

	if !matched {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected '%s' to match pattern '%s'", str, pattern), msgAndArgs...)
	}
}

// NotMatches 断言不匹配正则表达式
func (a *Assert) NotMatches(pattern, str string, msgAndArgs ...any) {
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		a.t.Helper()
		a.fail(fmt.Sprintf("Invalid regex pattern '%s': %v", pattern, err), msgAndArgs...)
		return
	}

	if matched {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected '%s' to not match pattern '%s'", str, pattern), msgAndArgs...)
	}
}

// ============= 数值断言 =============

// Greater 断言大于
func (a *Assert) Greater(actual, expected any, msgAndArgs ...any) {
	if !isGreater(actual, expected) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to be greater than %v", actual, expected), msgAndArgs...)
	}
}

// GreaterOrEqual 断言大于等于
func (a *Assert) GreaterOrEqual(actual, expected any, msgAndArgs ...any) {
	if !isGreaterOrEqual(actual, expected) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to be greater than or equal to %v", actual, expected), msgAndArgs...)
	}
}

// Less 断言小于
func (a *Assert) Less(actual, expected any, msgAndArgs ...any) {
	if !isLess(actual, expected) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to be less than %v", actual, expected), msgAndArgs...)
	}
}

// LessOrEqual 断言小于等于
func (a *Assert) LessOrEqual(actual, expected any, msgAndArgs ...any) {
	if !isLessOrEqual(actual, expected) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to be less than or equal to %v", actual, expected), msgAndArgs...)
	}
}

// InRange 断言在范围内
func (a *Assert) InRange(value, min, max any, msgAndArgs ...any) {
	if !isGreaterOrEqual(value, min) || !isLessOrEqual(value, max) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to be in range [%v, %v]", value, min, max), msgAndArgs...)
	}
}

// NotInRange 断言不在范围内
func (a *Assert) NotInRange(value, min, max any, msgAndArgs ...any) {
	if isGreaterOrEqual(value, min) && isLessOrEqual(value, max) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to not be in range [%v, %v]", value, min, max), msgAndArgs...)
	}
}

// InDelta 断言在误差范围内
func (a *Assert) InDelta(expected, actual, delta float64, msgAndArgs ...any) {
	if math.Abs(expected-actual) > delta {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %f to be within %f of %f", actual, delta, expected), msgAndArgs...)
	}
}

// ============= 集合断言 =============

// Len 断言长度
func (a *Assert) Len(object any, expectedLen int, msgAndArgs ...any) {
	actualLen := getLen(object)
	if actualLen != expectedLen {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected length %d, got %d", expectedLen, actualLen), msgAndArgs...)
	}
}

// ElementsMatch 断言元素匹配（忽略顺序）
func (a *Assert) ElementsMatch(expected, actual any, msgAndArgs ...any) {
	if !elementsMatch(expected, actual) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected elements to match: %#v vs %#v", expected, actual), msgAndArgs...)
	}
}

// ContainsElement 断言包含元素
func (a *Assert) ContainsElement(collection, element any, msgAndArgs ...any) {
	if !containsElement(collection, element) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected collection %#v to contain element %#v", collection, element), msgAndArgs...)
	}
}

// NotContainsElement 断言不包含元素
func (a *Assert) NotContainsElement(collection, element any, msgAndArgs ...any) {
	if containsElement(collection, element) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected collection %#v to not contain element %#v", collection, element), msgAndArgs...)
	}
}

// Subset 断言是子集
func (a *Assert) Subset(subset, set any, msgAndArgs ...any) {
	if !isSubset(subset, set) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %#v to be a subset of %#v", subset, set), msgAndArgs...)
	}
}

// ============= 类型断言 =============

// IsType 断言类型
func (a *Assert) IsType(expectedType, actual any, msgAndArgs ...any) {
	expectedT := reflect.TypeOf(expectedType)
	actualT := reflect.TypeOf(actual)

	if expectedT != actualT {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected type %v, got %v", expectedT, actualT), msgAndArgs...)
	}
}

// Implements 断言实现接口
func (a *Assert) Implements(interfaceType, object any, msgAndArgs ...any) {
	interfaceT := reflect.TypeOf(interfaceType).Elem()
	objectT := reflect.TypeOf(object)

	if objectT == nil || !objectT.Implements(interfaceT) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to implement %v", objectT, interfaceT), msgAndArgs...)
	}
}

// ============= 时间断言 =============

// WithinDuration 断言在时间范围内
func (a *Assert) WithinDuration(expected, actual time.Time, delta time.Duration, msgAndArgs ...any) {
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}

	if diff > delta {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected time %v to be within %v of %v", actual, delta, expected), msgAndArgs...)
	}
}

// Before 断言时间在之前
func (a *Assert) Before(t1, t2 time.Time, msgAndArgs ...any) {
	if !t1.Before(t2) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to be before %v", t1, t2), msgAndArgs...)
	}
}

// After 断言时间在之后
func (a *Assert) After(t1, t2 time.Time, msgAndArgs ...any) {
	if !t1.After(t2) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected %v to be after %v", t1, t2), msgAndArgs...)
	}
}

// ============= JSON断言 =============

// JSONEq 断言JSON相等
func (a *Assert) JSONEq(expected, actual string, msgAndArgs ...any) {
	var expectedJSON, actualJSON any

	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		a.t.Helper()
		a.fail(fmt.Sprintf("Invalid expected JSON: %v", err), msgAndArgs...)
		return
	}

	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		a.t.Helper()
		a.fail(fmt.Sprintf("Invalid actual JSON: %v", err), msgAndArgs...)
		return
	}

	if !isEqual(expectedJSON, actualJSON) {
		a.t.Helper()
		a.fail(fmt.Sprintf("JSON not equal:\nExpected: %s\nActual: %s", expected, actual), msgAndArgs...)
	}
}

// ValidJSON 断言为有效JSON
func (a *Assert) ValidJSON(data string, msgAndArgs ...any) {
	var v any
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		a.t.Helper()
		a.fail(fmt.Sprintf("Invalid JSON: %v", err), msgAndArgs...)
	}
}

// ============= 异常断言 =============

// Panics 断言会panic
func (a *Assert) Panics(fn func(), msgAndArgs ...any) {
	defer func() {
		if r := recover(); r == nil {
			a.t.Helper()
			a.fail("Expected function to panic", msgAndArgs...)
		}
	}()
	fn()
}

// NotPanics 断言不会panic
func (a *Assert) NotPanics(fn func(), msgAndArgs ...any) {
	defer func() {
		if r := recover(); r != nil {
			a.t.Helper()
			a.fail(fmt.Sprintf("Function panicked: %v", r), msgAndArgs...)
		}
	}()
	fn()
}

// PanicsWithValue 断言panic的值
func (a *Assert) PanicsWithValue(expected any, fn func(), msgAndArgs ...any) {
	var actual any
	func() {
		defer func() {
			actual = recover()
		}()
		fn()
	}()

	if actual == nil {
		a.t.Helper()
		a.fail("Expected function to panic", msgAndArgs...)
		return
	}

	if !isEqual(expected, actual) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected panic value %#v, got %#v", expected, actual), msgAndArgs...)
	}
}

// ============= 错误断言 =============

// NoError 断言无错误
func (a *Assert) NoError(err error, msgAndArgs ...any) {
	if err != nil {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected no error, got %v", err), msgAndArgs...)
	}
}

// Error 断言有错误
func (a *Assert) Error(err error, msgAndArgs ...any) {
	if err == nil {
		a.t.Helper()
		a.fail("Expected error, got nil", msgAndArgs...)
	}
}

// ErrorContains 断言错误包含指定文本
func (a *Assert) ErrorContains(err error, contains string, msgAndArgs ...any) {
	if err == nil {
		a.t.Helper()
		a.fail("Expected error, got nil", msgAndArgs...)
		return
	}

	if !strings.Contains(err.Error(), contains) {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected error to contain '%s', got '%s'", contains, err.Error()), msgAndArgs...)
	}
}

// ErrorIs 断言错误类型
func (a *Assert) ErrorIs(err, target error, msgAndArgs ...any) {
	if err == nil {
		a.t.Helper()
		a.fail("Expected error, got nil", msgAndArgs...)
		return
	}

	// 简单的错误比较
	if err.Error() != target.Error() {
		a.t.Helper()
		a.fail(fmt.Sprintf("Expected error '%v', got '%v'", target, err), msgAndArgs...)
	}
}

// ============= 辅助方法 =============

// fail 断言失败
func (a *Assert) fail(message string, msgAndArgs ...any) {
	if len(msgAndArgs) > 0 {
		if msg, ok := msgAndArgs[0].(string); ok {
			if len(msgAndArgs) > 1 {
				message = fmt.Sprintf(msg, msgAndArgs[1:]...) + " - " + message
			} else {
				message = msg + " - " + message
			}
		}
	}

	a.t.Error(message)
}

// ============= 辅助函数 =============

// isEqual 判断是否相等
func isEqual(expected, actual any) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	expectedValue := reflect.ValueOf(expected)
	actualValue := reflect.ValueOf(actual)

	if expectedValue.Type() != actualValue.Type() {
		return false
	}

	return reflect.DeepEqual(expected, actual)
}

// isNil 判断是否为nil
func isNil(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}

	return false
}

// isEmpty 判断是否为空
func isEmpty(value any) bool {
	if isNil(value) {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isEmpty(v.Elem().Interface())
	default:
		return false
	}
}

// isZero 判断是否为零值
func isZero(value any) bool {
	if isNil(value) {
		return true
	}

	v := reflect.ValueOf(value)
	return v.IsZero()
}

// getLen 获取长度
func getLen(object any) int {
	if object == nil {
		return 0
	}

	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len()
	default:
		return 0
	}
}

// containsElement 判断集合是否包含元素
func containsElement(collection, element any) bool {
	if collection == nil {
		return false
	}

	v := reflect.ValueOf(collection)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if isEqual(v.Index(i).Interface(), element) {
				return true
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if isEqual(v.MapIndex(key).Interface(), element) {
				return true
			}
		}
	case reflect.String:
		if str, ok := element.(string); ok {
			return strings.Contains(v.String(), str)
		}
	}

	return false
}

// elementsMatch 判断元素是否匹配（忽略顺序）
func elementsMatch(expected, actual any) bool {
	if expected == nil && actual == nil {
		return true
	}

	if expected == nil || actual == nil {
		return false
	}

	expectedV := reflect.ValueOf(expected)
	actualV := reflect.ValueOf(actual)

	if expectedV.Kind() != actualV.Kind() {
		return false
	}

	switch expectedV.Kind() {
	case reflect.Array, reflect.Slice:
		if expectedV.Len() != actualV.Len() {
			return false
		}

		// 创建映射来计数每个元素
		expectedCount := make(map[any]int)
		actualCount := make(map[any]int)

		for i := 0; i < expectedV.Len(); i++ {
			elem := expectedV.Index(i).Interface()
			expectedCount[elem]++
		}

		for i := 0; i < actualV.Len(); i++ {
			elem := actualV.Index(i).Interface()
			actualCount[elem]++
		}

		return reflect.DeepEqual(expectedCount, actualCount)
	}

	return false
}

// isSubset 判断是否为子集
func isSubset(subset, set any) bool {
	if subset == nil {
		return true
	}

	if set == nil {
		return false
	}

	subsetV := reflect.ValueOf(subset)

	switch subsetV.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < subsetV.Len(); i++ {
			if !containsElement(set, subsetV.Index(i).Interface()) {
				return false
			}
		}
		return true
	}

	return false
}

// 数值比较函数
func isGreater(actual, expected any) bool {
	return compareNumbers(actual, expected) > 0
}

func isGreaterOrEqual(actual, expected any) bool {
	return compareNumbers(actual, expected) >= 0
}

func isLess(actual, expected any) bool {
	return compareNumbers(actual, expected) < 0
}

func isLessOrEqual(actual, expected any) bool {
	return compareNumbers(actual, expected) <= 0
}

// compareNumbers 比较数值
func compareNumbers(a, b any) int {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// 转换为float64进行比较
	var aFloat, bFloat float64

	switch aVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		aFloat = float64(aVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		aFloat = float64(aVal.Uint())
	case reflect.Float32, reflect.Float64:
		aFloat = aVal.Float()
	default:
		return 0
	}

	switch bVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bFloat = float64(bVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bFloat = float64(bVal.Uint())
	case reflect.Float32, reflect.Float64:
		bFloat = bVal.Float()
	default:
		return 0
	}

	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

// ============= 全局断言函数 =============

// 为了方便使用，提供全局断言函数

func True(t *testing.T, condition bool, msgAndArgs ...any) {
	NewAssert(t).True(condition, msgAndArgs...)
}

func False(t *testing.T, condition bool, msgAndArgs ...any) {
	NewAssert(t).False(condition, msgAndArgs...)
}

func Equal(t *testing.T, expected, actual any, msgAndArgs ...any) {
	NewAssert(t).Equal(expected, actual, msgAndArgs...)
}

func NotEqual(t *testing.T, expected, actual any, msgAndArgs ...any) {
	NewAssert(t).NotEqual(expected, actual, msgAndArgs...)
}

func Nil(t *testing.T, value any, msgAndArgs ...any) {
	NewAssert(t).Nil(value, msgAndArgs...)
}

func NotNil(t *testing.T, value any, msgAndArgs ...any) {
	NewAssert(t).NotNil(value, msgAndArgs...)
}

func NoError(t *testing.T, err error, msgAndArgs ...any) {
	NewAssert(t).NoError(err, msgAndArgs...)
}

func Error(t *testing.T, err error, msgAndArgs ...any) {
	NewAssert(t).Error(err, msgAndArgs...)
}

func Contains(t *testing.T, haystack, needle string, msgAndArgs ...any) {
	NewAssert(t).Contains(haystack, needle, msgAndArgs...)
}

func Len(t *testing.T, object any, expectedLen int, msgAndArgs ...any) {
	NewAssert(t).Len(object, expectedLen, msgAndArgs...)
}

func Panics(t *testing.T, fn func(), msgAndArgs ...any) {
	NewAssert(t).Panics(fn, msgAndArgs...)
}

func NotPanics(t *testing.T, fn func(), msgAndArgs ...any) {
	NewAssert(t).NotPanics(fn, msgAndArgs...)
}
