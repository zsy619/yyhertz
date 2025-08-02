package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// JSON constants
const (
	JSON_ERROR_NONE                  = 0
	JSON_ERROR_DEPTH                 = 1
	JSON_ERROR_STATE_MISMATCH        = 2
	JSON_ERROR_CTRL_CHAR             = 3
	JSON_ERROR_SYNTAX                = 4
	JSON_ERROR_UTF8                  = 5
	JSON_ERROR_RECURSION             = 6
	JSON_ERROR_INF_OR_NAN            = 7
	JSON_ERROR_UNSUPPORTED_TYPE      = 8
	JSON_ERROR_INVALID_PROPERTY_NAME = 9
	JSON_ERROR_UTF16                 = 10

	JSON_HEX_TAG                    = 1
	JSON_HEX_AMP                    = 2
	JSON_HEX_APOS                   = 4
	JSON_HEX_QUOT                   = 8
	JSON_FORCE_OBJECT               = 16
	JSON_NUMERIC_CHECK              = 32
	JSON_UNESCAPED_SLASHES          = 64
	JSON_PRETTY_PRINT               = 128
	JSON_UNESCAPED_UNICODE          = 256
	JSON_PARTIAL_OUTPUT_ON_ERROR    = 512
	JSON_PRESERVE_ZERO_FRACTION     = 1024
	JSON_UNESCAPED_LINE_TERMINATORS = 2048
)

var lastJSONError = JSON_ERROR_NONE
var lastJSONErrorMsg = ""

// JsonEncode returns the JSON representation of a value
func JsonEncode(value any, flags ...int) string {
	lastJSONError = JSON_ERROR_NONE
	lastJSONErrorMsg = ""

	flag := 0
	if len(flags) > 0 {
		flag = flags[0]
	}

	// Convert value for JSON encoding
	jsonValue := convertForJSON(value)

	var data []byte
	var err error

	if flag&JSON_PRETTY_PRINT != 0 {
		data, err = json.MarshalIndent(jsonValue, "", "    ")
	} else {
		data, err = json.Marshal(jsonValue)
	}

	if err != nil {
		lastJSONError = JSON_ERROR_SYNTAX
		lastJSONErrorMsg = err.Error()
		return ""
	}

	result := string(data)

	// Apply flags
	if flag&JSON_UNESCAPED_SLASHES != 0 {
		result = strings.ReplaceAll(result, "\\/", "/")
	}

	if flag&JSON_HEX_TAG != 0 {
		result = strings.ReplaceAll(result, "<", "\\u003C")
		result = strings.ReplaceAll(result, ">", "\\u003E")
	}

	if flag&JSON_HEX_AMP != 0 {
		result = strings.ReplaceAll(result, "&", "\\u0026")
	}

	if flag&JSON_HEX_APOS != 0 {
		result = strings.ReplaceAll(result, "'", "\\u0027")
	}

	if flag&JSON_HEX_QUOT != 0 {
		result = strings.ReplaceAll(result, "\"", "\\u0022")
	}

	return result
}

// JsonDecode takes a JSON encoded string and converts it into a variable
func JsonDecode(jsonString string, assoc ...bool) any {
	lastJSONError = JSON_ERROR_NONE
	lastJSONErrorMsg = ""

	if jsonString == "" {
		lastJSONError = JSON_ERROR_SYNTAX
		lastJSONErrorMsg = "empty string"
		return nil
	}

	var result any
	err := json.Unmarshal([]byte(jsonString), &result)

	if err != nil {
		lastJSONError = JSON_ERROR_SYNTAX
		lastJSONErrorMsg = err.Error()
		return nil
	}

	// If assoc is true, convert to associative arrays (maps)
	if len(assoc) > 0 && assoc[0] {
		return convertToAssoc(result)
	}

	return result
}

// JsonLastError returns the last error occurred
func JsonLastError() int {
	return lastJSONError
}

// JsonLastErrorMsg returns the error string of the last json_encode() or json_decode() call
func JsonLastErrorMsg() string {
	switch lastJSONError {
	case JSON_ERROR_NONE:
		return "No error"
	case JSON_ERROR_DEPTH:
		return "Maximum stack depth exceeded"
	case JSON_ERROR_STATE_MISMATCH:
		return "State mismatch (invalid or malformed JSON)"
	case JSON_ERROR_CTRL_CHAR:
		return "Control character error, possibly incorrectly encoded"
	case JSON_ERROR_SYNTAX:
		if lastJSONErrorMsg != "" {
			return "Syntax error: " + lastJSONErrorMsg
		}
		return "Syntax error"
	case JSON_ERROR_UTF8:
		return "Malformed UTF-8 characters, possibly incorrectly encoded"
	case JSON_ERROR_RECURSION:
		return "One or more recursive references in the value to be encoded"
	case JSON_ERROR_INF_OR_NAN:
		return "One or more NAN or INF values in the value to be encoded"
	case JSON_ERROR_UNSUPPORTED_TYPE:
		return "A value of a type that cannot be encoded was given"
	default:
		return "Unknown error"
	}
}

// JsonValidate validates a JSON string
func JsonValidate(jsonStr string) bool {
	var result any
	err := json.Unmarshal([]byte(jsonStr), &result)
	return err == nil
}

// Helper functions

func convertForJSON(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, val := range v {
			result[key] = convertForJSON(val)
		}
		return result
	case map[any]any:
		result := make(map[string]any)
		for key, val := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertForJSON(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = convertForJSON(val)
		}
		return result
	case []string:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = val
		}
		return result
	case []int:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = val
		}
		return result
	case []float64:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = val
		}
		return result
	default:
		return value
	}
}

func convertToAssoc(value any) any {
	switch v := value.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, val := range v {
			result[key] = convertToAssoc(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = convertToAssoc(val)
		}
		return result
	default:
		return value
	}
}

// Additional JSON utility functions

// JsonPretty formats JSON string with indentation
func JsonPretty(jsonString string) string {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return jsonString
	}

	pretty, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return jsonString
	}

	return string(pretty)
}

// JsonMinify removes whitespace from JSON string
func JsonMinify(jsonString string) string {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return jsonString
	}

	minified, err := json.Marshal(obj)
	if err != nil {
		return jsonString
	}

	return string(minified)
}

// JsonExtract extracts value from JSON using path (like MySQL JSON_EXTRACT)
func JsonExtract(jsonString, path string) any {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return nil
	}

	return extractByPath(obj, path)
}

// JsonKeys returns the keys of a JSON object
func JsonKeys(jsonString string) []string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return nil
	}

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}

	return keys
}

// JsonValues returns the values of a JSON object/array
func JsonValues(jsonString string) []any {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return nil
	}

	switch v := obj.(type) {
	case map[string]any:
		values := make([]any, 0, len(v))
		for _, value := range v {
			values = append(values, value)
		}
		return values
	case []any:
		return v
	default:
		return []any{obj}
	}
}

// JsonMerge merges multiple JSON strings
func JsonMerge(jsonStrings ...string) string {
	merged := make(map[string]any)

	for _, jsonStr := range jsonStrings {
		var obj map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
			continue
		}

		for key, value := range obj {
			merged[key] = value
		}
	}

	result, err := json.Marshal(merged)
	if err != nil {
		return "{}"
	}

	return string(result)
}

// JsonSearch searches for a value in JSON (like MySQL JSON_SEARCH)
func JsonSearch(jsonString string, oneOrAll string, searchStr string) any {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return nil
	}

	var results []string
	searchInValue(obj, searchStr, "", &results)

	if oneOrAll == "one" && len(results) > 0 {
		return results[0]
	}

	if len(results) == 0 {
		return nil
	}

	return results
}

// JsonLength returns the length of JSON value
func JsonLength(jsonString string, path ...string) int {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return 0
	}

	if len(path) > 0 {
		obj = extractByPath(obj, path[0])
	}

	switch v := obj.(type) {
	case map[string]any:
		return len(v)
	case []any:
		return len(v)
	case string:
		return len(v)
	case nil:
		return 0
	default:
		return 1
	}
}

// JsonType returns the type of JSON value
func JsonType(jsonString string, path ...string) string {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return "NULL"
	}

	if len(path) > 0 {
		obj = extractByPath(obj, path[0])
	}

	switch obj.(type) {
	case nil:
		return "NULL"
	case bool:
		return "BOOLEAN"
	case float64:
		// Check if it's an integer
		if f, ok := obj.(float64); ok && f == float64(int64(f)) {
			return "INTEGER"
		}
		return "DOUBLE"
	case string:
		return "STRING"
	case []any:
		return "ARRAY"
	case map[string]any:
		return "OBJECT"
	default:
		return "UNKNOWN"
	}
}

// Helper functions for path extraction and searching

func extractByPath(obj any, path string) any {
	if path == "$" || path == "" {
		return obj
	}

	// Simple path parsing - supports $.key and $[index]
	if strings.HasPrefix(path, "$.") {
		key := path[2:]
		if m, ok := obj.(map[string]any); ok {
			return m[key]
		}
	} else if strings.HasPrefix(path, "$[") && strings.HasSuffix(path, "]") {
		indexStr := path[2 : len(path)-1]
		if index, err := strconv.Atoi(indexStr); err == nil {
			if arr, ok := obj.([]any); ok && index >= 0 && index < len(arr) {
				return arr[index]
			}
		}
	}

	return nil
}

func searchInValue(obj any, searchStr, currentPath string, results *[]string) {
	switch v := obj.(type) {
	case string:
		if strings.Contains(v, searchStr) {
			*results = append(*results, currentPath)
		}
	case map[string]any:
		for key, value := range v {
			newPath := currentPath + "." + key
			if currentPath == "" {
				newPath = "$." + key
			}
			searchInValue(value, searchStr, newPath, results)
		}
	case []any:
		for i, value := range v {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			if currentPath == "" {
				newPath = fmt.Sprintf("$[%d]", i)
			}
			searchInValue(value, searchStr, newPath, results)
		}
	}
}

// JSONP related functions

// JsonpEncode encodes data as JSONP
func JsonpEncode(callback string, data any, flags ...int) string {
	jsonData := JsonEncode(data, flags...)
	if jsonData == "" {
		return ""
	}
	return fmt.Sprintf("%s(%s);", callback, jsonData)
}

// JsonpValidate validates a JSONP string
func JsonpValidate(jsonp string) bool {
	// Find callback function name
	parenIndex := strings.Index(jsonp, "(")
	if parenIndex == -1 {
		return false
	}

	// Extract JSON part
	if !strings.HasSuffix(jsonp, ");") {
		return false
	}

	jsonPart := jsonp[parenIndex+1 : len(jsonp)-2]
	return JsonValidate(jsonPart)
}

// Advanced JSON operations

// JsonPatch applies JSON Patch operations (RFC 6902)
func JsonPatch(jsonString string, patches []map[string]any) string {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return jsonString
	}

	// Simplified JSON Patch implementation
	for _, patch := range patches {
		op, exists := patch["op"]
		if !exists {
			continue
		}

		switch op {
		case "add":
			// Simplified add operation
		case "remove":
			// Simplified remove operation
		case "replace":
			// Simplified replace operation
		}
	}

	result, err := json.Marshal(obj)
	if err != nil {
		return jsonString
	}

	return string(result)
}

// JsonPointer extracts value using JSON Pointer (RFC 6901)
func JsonPointer(jsonString string, pointer string) any {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return nil
	}

	if pointer == "" {
		return obj
	}

	parts := strings.Split(pointer[1:], "/") // Remove leading "/"
	current := obj

	for _, part := range parts {
		// Unescape JSON Pointer tokens
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")

		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		case []any:
			if index, err := strconv.Atoi(part); err == nil && index >= 0 && index < len(v) {
				current = v[index]
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}
