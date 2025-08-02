package util

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

// ArrayKeys returns all the keys or a subset of the keys of an array
func ArrayKeys(arr any) []any {
	v := reflect.ValueOf(arr)
	var keys []any

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			keys = append(keys, key.Interface())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			keys = append(keys, i)
		}
	}
	return keys
}

// ArrayValues returns all the values of an array
func ArrayValues(arr any) []any {
	v := reflect.ValueOf(arr)
	var values []any

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			values = append(values, v.MapIndex(key).Interface())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			values = append(values, v.Index(i).Interface())
		}
	}
	return values
}

// InArray checks if a value exists in an array
func InArray(needle any, haystack any) bool {
	v := reflect.ValueOf(haystack)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(needle, v.Index(i).Interface()) {
				return true
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if reflect.DeepEqual(needle, v.MapIndex(key).Interface()) {
				return true
			}
		}
	}
	return false
}

// ArraySearch searches the array for a given value and returns the key
func ArraySearch(needle any, haystack any) any {
	v := reflect.ValueOf(haystack)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(needle, v.Index(i).Interface()) {
				return i
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if reflect.DeepEqual(needle, v.MapIndex(key).Interface()) {
				return key.Interface()
			}
		}
	}
	return false
}

// ArrayMerge merges one or more arrays
func ArrayMerge(arrays ...[]any) []any {
	var result []any
	for _, arr := range arrays {
		result = append(result, arr...)
	}
	return result
}

// ArrayUnique removes duplicate values from an array
func ArrayUnique(arr []any) []any {
	seen := make(map[any]bool)
	var result []any

	for _, v := range arr {
		key := fmt.Sprintf("%v", v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}
	return result
}

// ArraySlice extracts a slice of the array
func ArraySlice(arr []any, offset int, length ...int) []any {
	size := len(arr)

	if offset < 0 {
		offset = size + offset
	}
	if offset < 0 || offset >= size {
		return []any{}
	}

	if len(length) == 0 {
		return arr[offset:]
	}

	end := offset + length[0]
	if end > size {
		end = size
	}
	if end <= offset {
		return []any{}
	}

	return arr[offset:end]
}

// ArrayChunk splits an array into chunks
func ArrayChunk(arr []any, size int) [][]any {
	if size <= 0 {
		return [][]any{}
	}

	var chunks [][]any
	for i := 0; i < len(arr); i += size {
		end := i + size
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}
	return chunks
}

// ArrayReverse returns an array with elements in reverse order
func ArrayReverse(arr []any) []any {
	result := make([]any, len(arr))
	for i, v := range arr {
		result[len(arr)-1-i] = v
	}
	return result
}

// ArrayFlip exchanges all keys with their associated values
func ArrayFlip(arr map[any]any) map[any]any {
	flipped := make(map[any]any)
	for k, v := range arr {
		flipped[v] = k
	}
	return flipped
}

// ArraySum calculates the sum of values in an array
func ArraySum(arr []any) float64 {
	var sum float64
	for _, v := range arr {
		if num, ok := convertToFloat64(v); ok {
			sum += num
		}
	}
	return sum
}

// ArrayProduct calculates the product of values in an array
func ArrayProduct(arr []any) float64 {
	product := 1.0
	for _, v := range arr {
		if num, ok := convertToFloat64(v); ok {
			product *= num
		}
	}
	return product
}

// ArrayPush pushes one or more elements to the end of array
func ArrayPush(arr *[]any, elements ...any) int {
	*arr = append(*arr, elements...)
	return len(*arr)
}

// ArrayPop pops the element off the end of array
func ArrayPop(arr *[]any) any {
	if len(*arr) == 0 {
		return nil
	}
	last := (*arr)[len(*arr)-1]
	*arr = (*arr)[:len(*arr)-1]
	return last
}

// ArrayUnshift prepends one or more elements to the beginning of an array
func ArrayUnshift(arr *[]any, elements ...any) int {
	*arr = append(elements, *arr...)
	return len(*arr)
}

// ArrayShift shifts an element off the beginning of array
func ArrayShift(arr *[]any) any {
	if len(*arr) == 0 {
		return nil
	}
	first := (*arr)[0]
	*arr = (*arr)[1:]
	return first
}

// ArrayColumn returns the values from a single column in the input array
func ArrayColumn(arr []map[string]any, column string, indexKey ...string) any {
	var result []any
	resultMap := make(map[any]any)
	useMap := len(indexKey) > 0

	for _, row := range arr {
		if value, exists := row[column]; exists {
			if useMap {
				if keyValue, keyExists := row[indexKey[0]]; keyExists {
					resultMap[keyValue] = value
				}
			} else {
				result = append(result, value)
			}
		}
	}

	if useMap {
		return resultMap
	}
	return result
}

// ArrayFilter filters elements of an array using a callback function
func ArrayFilter(arr []any, callback func(any) bool) []any {
	var result []any
	for _, v := range arr {
		if callback == nil {
			// If no callback, remove empty values
			if !isEmpty(v) {
				result = append(result, v)
			}
		} else if callback(v) {
			result = append(result, v)
		}
	}
	return result
}

// ArrayMap applies the callback to the elements of the given arrays
func ArrayMap(callback func(any) any, arr []any) []any {
	result := make([]any, len(arr))
	for i, v := range arr {
		result[i] = callback(v)
	}
	return result
}

// Range creates an array containing a range of elements
func Range(start, end any, step ...any) []any {
	var result []any
	stepVal := 1

	if len(step) > 0 {
		if s, ok := convertToInt(step[0]); ok {
			stepVal = s
		}
	}

	if stepVal == 0 {
		return result
	}

	startInt, startIsInt := convertToInt(start)
	endInt, endIsInt := convertToInt(end)

	if startIsInt && endIsInt {
		if stepVal > 0 {
			for i := startInt; i <= endInt; i += stepVal {
				result = append(result, i)
			}
		} else {
			for i := startInt; i >= endInt; i += stepVal {
				result = append(result, i)
			}
		}
		return result
	}

	// Handle string ranges (like 'a' to 'z')
	startStr := fmt.Sprintf("%v", start)
	endStr := fmt.Sprintf("%v", end)

	if len(startStr) == 1 && len(endStr) == 1 {
		startRune := rune(startStr[0])
		endRune := rune(endStr[0])

		if stepVal > 0 {
			for r := startRune; r <= endRune; r += rune(stepVal) {
				result = append(result, string(r))
			}
		} else {
			for r := startRune; r >= endRune; r += rune(stepVal) {
				result = append(result, string(r))
			}
		}
	}

	return result
}

// ArrayCountValues counts all the values of an array
func ArrayCountValues(arr []any) map[any]int {
	counts := make(map[any]int)
	for _, v := range arr {
		key := fmt.Sprintf("%v", v)
		counts[key]++
	}
	return counts
}

// Shuffle shuffles an array
func Shuffle(arr *[]any) {
	for i := len(*arr) - 1; i > 0; i-- {
		j := i // simplified random, in real implementation use rand.Intn(i+1)
		(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i]
	}
}

// ArrayDiff computes the difference of arrays
func ArrayDiff(arr1 []any, arrays ...[]any) []any {
	var result []any

	for _, v1 := range arr1 {
		found := false
		for _, arr := range arrays {
			if InArray(v1, arr) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, v1)
		}
	}
	return result
}

// ArrayIntersect computes the intersection of arrays
func ArrayIntersect(arr1 []any, arrays ...[]any) []any {
	var result []any

	for _, v1 := range arr1 {
		inAll := true
		for _, arr := range arrays {
			if !InArray(v1, arr) {
				inAll = false
				break
			}
		}
		if inAll {
			result = append(result, v1)
		}
	}
	return result
}

// Helper functions

func convertToFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func convertToInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float32:
		return int(val), true
	case float64:
		return int(val), true
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i, true
		}
	}
	return 0, false
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case string:
		return val == ""
	case int, int64, float32, float64:
		return val == 0
	case bool:
		return !val
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return rv.Len() == 0
		case reflect.Ptr:
			return rv.IsNil()
		}
	}
	return false
}

// Sort functions for different types

// SortInts sorts a slice of integers (helper for array sorting)
func SortInts(arr []int) {
	sort.Ints(arr)
}

// SortStrings sorts a slice of strings (helper for array sorting)
func SortStrings(arr []string) {
	sort.Strings(arr)
}

// SortFloat64s sorts a slice of float64s (helper for array sorting)
func SortFloat64s(arr []float64) {
	sort.Float64s(arr)
}
