package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Isset determines if a variable is declared and is different than null
func Isset(vars ...any) bool {
	for _, v := range vars {
		if v == nil {
			return false
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			return false
		}
		if rv.Kind() == reflect.Interface && rv.IsNil() {
			return false
		}
	}
	return true
}

// Empty determines whether a variable is empty
func Empty(val any) bool {
	if val == nil {
		return true
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.String:
		return rv.String() == "" || rv.String() == "0"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0.0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Chan:
		return rv.IsNil()
	case reflect.Func:
		return rv.IsNil()
	}
	return false
}

// IsNull finds whether a variable is null
func IsNull(val any) bool {
	if val == nil {
		return true
	}
	rv := reflect.ValueOf(val)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

// IsArray finds whether a variable is an array
func IsArray(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	return rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice
}

// IsObject finds whether a variable is an object
func IsObject(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	return rv.Kind() == reflect.Struct || (rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct)
}

// IsString finds whether a variable is a string
func IsString(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	return rv.Kind() == reflect.String
}

// IsInt finds whether a variable is an integer
func IsInt(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

// IsFloat finds whether a variable is a float
func IsFloat(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	return rv.Kind() == reflect.Float32 || rv.Kind() == reflect.Float64
}

// IsBool finds whether a variable is a boolean
func IsBool(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	return rv.Kind() == reflect.Bool
}

// IsResource finds whether a variable is a resource
func IsResource(val any) bool {
	// Go doesn't have resources like PHP, but we can check for file handles, etc.
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	// Check if it's a pointer to a struct that might represent a resource
	return rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct
}

// IsCallable finds whether a variable is callable
func IsCallable(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	return rv.Kind() == reflect.Func
}

// IsScalar finds whether a variable is a scalar
func IsScalar(val any) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	}
	return false
}

// GetType returns the type of a variable
func GetType(val any) string {
	if val == nil {
		return "NULL"
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "double"
	case reflect.String:
		return "string"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Map:
		return "array"
	case reflect.Struct:
		return "object"
	case reflect.Ptr:
		if rv.IsNil() {
			return "NULL"
		}
		return "object"
	case reflect.Func:
		return "resource"
	default:
		return "unknown type"
	}
}

// VarDump dumps information about a variable
func VarDump(vars ...any) string {
	var result strings.Builder

	for i, v := range vars {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(varDumpValue(v, 0))
	}

	return result.String()
}

// VarExport outputs or returns a parsable string representation of a variable
func VarExport(val any, returnResult ...bool) string {
	result := varExportValue(val, 0)
	if len(returnResult) > 0 && returnResult[0] {
		return result
	}
	fmt.Print(result)
	return ""
}

// PrintR prints human-readable information about a variable
func PrintR(val any, returnResult ...bool) string {
	result := printRValue(val, 0)
	if len(returnResult) > 0 && returnResult[0] {
		return result
	}
	fmt.Print(result)
	return ""
}

// Serialize generates a storable representation of a value
func Serialize(val any) string {
	return serializeValue(val)
}

// Unserialize creates a PHP value from a stored representation
func Unserialize(data string) (any, error) {
	return unserializeValue(data)
}

// Strval returns the string value of a variable
func Strval(val any) string {
	if val == nil {
		return ""
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Bool:
		if rv.Bool() {
			return "1"
		}
		return ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.Array, reflect.Slice:
		return "Array"
	case reflect.Map:
		return "Array"
	case reflect.Struct, reflect.Ptr:
		return "Object"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Boolval returns the boolean value of a variable
func Boolval(val any) bool {
	if val == nil {
		return false
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		s := rv.String()
		return s != "" && s != "0"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0.0
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !rv.IsNil()
	default:
		return true
	}
}

// Count counts all elements in an array or object
func Count(val any) int {
	if val == nil {
		return 0
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return rv.Len()
	case reflect.Ptr:
		if !rv.IsNil() {
			return Count(rv.Elem().Interface())
		}
		return 0
	default:
		return 1
	}
}

// Sizeof alias for Count
func Sizeof(val any) int {
	return Count(val)
}

// Unset unsets a given variable (simplified - Go doesn't support this directly)
func Unset(val any) {
	// In Go, we can't actually unset variables like in PHP
	// This is just a placeholder function
	// In practice, you would set the variable to nil or zero value
}

// Helper functions for var_dump, print_r, etc.

func varDumpValue(val any, depth int) string {
	if val == nil {
		return "NULL"
	}

	rv := reflect.ValueOf(val)
	rt := reflect.TypeOf(val)

	indent := strings.Repeat("  ", depth)

	switch rv.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("bool(%t)", rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("int(%d)", rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("int(%d)", rv.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("float(%g)", rv.Float())
	case reflect.String:
		return fmt.Sprintf("string(%d) \"%s\"", rv.Len(), rv.String())
	case reflect.Array, reflect.Slice:
		var result strings.Builder
		result.WriteString(fmt.Sprintf("array(%d) {\n", rv.Len()))
		for i := 0; i < rv.Len(); i++ {
			result.WriteString(fmt.Sprintf("%s  [%d]=>\n", indent, i))
			result.WriteString(fmt.Sprintf("%s  %s\n", indent, varDumpValue(rv.Index(i).Interface(), depth+1)))
		}
		result.WriteString(indent + "}")
		return result.String()
	case reflect.Map:
		var result strings.Builder
		result.WriteString(fmt.Sprintf("array(%d) {\n", rv.Len()))
		for _, key := range rv.MapKeys() {
			result.WriteString(fmt.Sprintf("%s  [%v]=>\n", indent, key.Interface()))
			result.WriteString(fmt.Sprintf("%s  %s\n", indent, varDumpValue(rv.MapIndex(key).Interface(), depth+1)))
		}
		result.WriteString(indent + "}")
		return result.String()
	case reflect.Struct:
		var result strings.Builder
		result.WriteString(fmt.Sprintf("object(%s) {\n", rt.Name()))
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.IsExported() {
				result.WriteString(fmt.Sprintf("%s  [\"%s\"]=>\n", indent, field.Name))
				result.WriteString(fmt.Sprintf("%s  %s\n", indent, varDumpValue(rv.Field(i).Interface(), depth+1)))
			}
		}
		result.WriteString(indent + "}")
		return result.String()
	case reflect.Ptr:
		if rv.IsNil() {
			return "NULL"
		}
		return varDumpValue(rv.Elem().Interface(), depth)
	default:
		return fmt.Sprintf("resource(%v)", val)
	}
}

func printRValue(val any, depth int) string {
	if val == nil {
		return ""
	}

	rv := reflect.ValueOf(val)
	indent := strings.Repeat("    ", depth)

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		var result strings.Builder
		result.WriteString("Array\n" + indent + "(\n")
		for i := 0; i < rv.Len(); i++ {
			result.WriteString(fmt.Sprintf("%s    [%d] => ", indent, i))
			result.WriteString(printRValue(rv.Index(i).Interface(), depth+1))
			result.WriteString("\n")
		}
		result.WriteString(indent + ")")
		return result.String()
	case reflect.Map:
		var result strings.Builder
		result.WriteString("Array\n" + indent + "(\n")
		for _, key := range rv.MapKeys() {
			result.WriteString(fmt.Sprintf("%s    [%v] => ", indent, key.Interface()))
			result.WriteString(printRValue(rv.MapIndex(key).Interface(), depth+1))
			result.WriteString("\n")
		}
		result.WriteString(indent + ")")
		return result.String()
	case reflect.Struct:
		var result strings.Builder
		rt := reflect.TypeOf(val)
		result.WriteString(fmt.Sprintf("%s Object\n%s(\n", rt.Name(), indent))
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.IsExported() {
				result.WriteString(fmt.Sprintf("%s    [%s] => ", indent, field.Name))
				result.WriteString(printRValue(rv.Field(i).Interface(), depth+1))
				result.WriteString("\n")
			}
		}
		result.WriteString(indent + ")")
		return result.String()
	case reflect.Ptr:
		if rv.IsNil() {
			return ""
		}
		return printRValue(rv.Elem().Interface(), depth)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func varExportValue(val any, depth int) string {
	if val == nil {
		return "NULL"
	}

	rv := reflect.ValueOf(val)

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return "true"
		}
		return "false"
	case reflect.String:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(rv.String(), "'", "\\'"))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.Array, reflect.Slice:
		var result strings.Builder
		result.WriteString("array(\n")
		for i := 0; i < rv.Len(); i++ {
			result.WriteString(strings.Repeat("  ", depth+1))
			result.WriteString(fmt.Sprintf("%d => %s,\n", i, varExportValue(rv.Index(i).Interface(), depth+1)))
		}
		result.WriteString(strings.Repeat("  ", depth) + ")")
		return result.String()
	case reflect.Map:
		var result strings.Builder
		result.WriteString("array(\n")
		for _, key := range rv.MapKeys() {
			result.WriteString(strings.Repeat("  ", depth+1))
			keyStr := varExportValue(key.Interface(), depth+1)
			valStr := varExportValue(rv.MapIndex(key).Interface(), depth+1)
			result.WriteString(fmt.Sprintf("%s => %s,\n", keyStr, valStr))
		}
		result.WriteString(strings.Repeat("  ", depth) + ")")
		return result.String()
	case reflect.Ptr:
		if rv.IsNil() {
			return "NULL"
		}
		return varExportValue(rv.Elem().Interface(), depth)
	default:
		return fmt.Sprintf("'%v'", val)
	}
}

func serializeValue(val any) string {
	if val == nil {
		return "N;"
	}

	rv := reflect.ValueOf(val)

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return "b:1;"
		}
		return "b:0;"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("i:%d;", rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("i:%d;", rv.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("d:%g;", rv.Float())
	case reflect.String:
		s := rv.String()
		return fmt.Sprintf("s:%d:\"%s\";", len(s), s)
	case reflect.Array, reflect.Slice:
		var result strings.Builder
		result.WriteString(fmt.Sprintf("a:%d:{", rv.Len()))
		for i := 0; i < rv.Len(); i++ {
			result.WriteString(fmt.Sprintf("i:%d;", i))
			result.WriteString(serializeValue(rv.Index(i).Interface()))
		}
		result.WriteString("}")
		return result.String()
	case reflect.Ptr:
		if rv.IsNil() {
			return "N;"
		}
		return serializeValue(rv.Elem().Interface())
	default:
		return fmt.Sprintf("s:%d:\"%v\";", len(fmt.Sprintf("%v", val)), val)
	}
}

func unserializeValue(data string) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	switch data[0] {
	case 'N':
		return nil, nil
	case 'b':
		if strings.HasPrefix(data, "b:1;") {
			return true, nil
		}
		return false, nil
	case 'i':
		parts := strings.Split(data, ":")
		if len(parts) >= 2 {
			numStr := strings.TrimSuffix(parts[1], ";")
			return strconv.ParseInt(numStr, 10, 64)
		}
	case 'd':
		parts := strings.Split(data, ":")
		if len(parts) >= 2 {
			numStr := strings.TrimSuffix(parts[1], ";")
			return strconv.ParseFloat(numStr, 64)
		}
	case 's':
		// Simple string parsing - s:length:"content";
		parts := strings.SplitN(data, ":", 3)
		if len(parts) >= 3 {
			lengthStr := parts[1]
			length, err := strconv.Atoi(lengthStr)
			if err != nil {
				return nil, err
			}
			content := parts[2]
			if strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\";") {
				str := content[1 : len(content)-2]
				if len(str) == length {
					return str, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("unsupported serialization format")
}
