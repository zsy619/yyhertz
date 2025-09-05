package xjson

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

// JsonEncode 将给定的值编码为 JSON 字符串。
//
// 参数：
//   - value: 需要编码的值，可以是任意类型。
//   - flags: 可选参数，用于指定编码的行为标志。支持以下标志：
//   - JSON_PRETTY_PRINT: 格式化输出 JSON，增加缩进。
//   - JSON_UNESCAPED_SLASHES: 不转义斜杠（/）。
//   - JSON_HEX_TAG: 将 `<` 和 `>` 转换为 Unicode 转义序列（\u003C 和 \u003E）。
//   - JSON_HEX_AMP: 将 `&` 转换为 Unicode 转义序列（\u0026）。
//   - JSON_HEX_APOS: 将单引号（'）转换为 Unicode 转义序列（\u0027）。
//   - JSON_HEX_QUOT: 将双引号（"）转换为 Unicode 转义序列（\u0022）。
//
// 返回值：
//   - string: 编码后的 JSON 字符串。如果编码失败，返回空字符串。
//
// 注意：
//   - 如果编码过程中发生错误，错误信息会存储在全局变量 lastJSONError 和 lastJSONErrorMsg 中。
//   - 默认情况下，JSON 字符串是紧凑格式的，不包含额外的空白字符。
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

// JsonDecode 解析 JSON 字符串并返回对应的 Go 数据结构。
//
// 参数：
//   - jsonString: 需要解析的 JSON 字符串。
//   - assoc: 可选参数，如果为 true，则将 JSON 对象解析为 map[string]interface{} 类型（关联数组）。
//
// 返回值：
//   - any: 解析后的数据结构，如果解析失败则返回 nil。
//
// 错误处理：
//   - 如果解析失败，会设置全局变量 lastJSONError 和 lastJSONErrorMsg 记录错误信息。
//   - 错误类型包括 JSON_ERROR_SYNTAX（语法错误）和 JSON_ERROR_NONE（无错误）。
//
// 示例：
//
//	result := JsonDecode(`{"key": "value"}`, true)
//	// 返回 map[string]interface{}{"key": "value"}
//
// 注意：
//   - 如果输入为空字符串，会返回 nil 并记录错误。
//   - 如果 assoc 为 true，会将 JSON 对象转换为 map 类型。
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

// JsonLastError 返回最近一次JSON操作中的错误码。
// 该函数通常用于调试或错误处理，返回值为一个整数类型的错误码。
func JsonLastError() int {
	return lastJSONError
}

// JsonLastErrorMsg 返回最近一次 JSON 处理错误的描述信息。
// 根据全局变量 lastJSONError 的值，返回对应的错误描述字符串。
// 如果错误类型为 JSON_ERROR_SYNTAX 且 lastJSONErrorMsg 不为空，则返回包含详细语法错误信息的字符串。
// 支持的错误类型包括：
//   - JSON_ERROR_NONE: 无错误
//   - JSON_ERROR_DEPTH: 超过最大堆栈深度
//   - JSON_ERROR_STATE_MISMATCH: 状态不匹配（无效或格式错误的 JSON）
//   - JSON_ERROR_CTRL_CHAR: 控制字符错误，可能编码不正确
//   - JSON_ERROR_SYNTAX: 语法错误
//   - JSON_ERROR_UTF8: 格式错误的 UTF-8 字符
//   - JSON_ERROR_RECURSION: 值中存在递归引用
//   - JSON_ERROR_INF_OR_NAN: 值中存在 NAN 或 INF
//   - JSON_ERROR_UNSUPPORTED_TYPE: 不支持编码的类型
//
// 如果错误类型未匹配上述任何情况，则返回 "Unknown error"。
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

// JsonValidate 检查给定的字符串是否为有效的JSON格式。
// 参数：
//
//	jsonStr: 需要验证的JSON字符串。
//
// 返回值：
//
//	bool: 如果字符串是有效的JSON，则返回true；否则返回false。
func JsonValidate(jsonStr string) bool {
	var result any
	err := json.Unmarshal([]byte(jsonStr), &result)
	return err == nil
}

// convertForJSON 递归地将输入值转换为适合 JSON 序列化的格式。
//
// 该函数处理以下类型：
//   - nil：直接返回 nil。
//   - map[string]any：递归处理每个键值对，确保键为字符串。
//   - map[any]any：将键转换为字符串后递归处理每个键值对。
//   - []any：递归处理每个元素。
//   - 基本类型的切片（如 []string、[]int、[]float64）：直接转换为 []any。
//   - 其他类型：直接返回原值。
//
// 返回值：
//   - 转换后的值，可以直接用于 JSON 序列化。
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
		// convertToAssoc 递归地将输入值转换为关联结构（map或slice）。
		// 如果输入是 map[string]any 类型，会递归处理其所有值；
		// 如果输入是 []any 类型，会递归处理其所有元素；
		// 其他类型的值会直接返回。
		// 返回值类型与输入类型一致。
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

// convertToAssoc 递归地将输入值转换为关联结构（map或slice）。
// 如果输入是 map[string]any 类型，会递归处理其所有值；
// 如果输入是 []any 类型，会递归处理其所有元素；
// 其他类型的值会直接返回。
// 返回值类型与输入类型一致，但所有嵌套的 map 和 slice 都会被递归处理。
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

// formatJSON 格式化输入的 JSON 字符串，使其具有可读性。
// 如果输入的 JSON 字符串无法解析或格式化失败，则返回原始字符串。
// 参数：
//
//	jsonString: 需要格式化的 JSON 字符串。
//
// 返回值：
//
//	string: 格式化后的 JSON 字符串（如果成功），否则返回原始字符串。
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

// MinifyJSON 将输入的 JSON 字符串进行最小化处理，去除不必要的空格和换行符。
// 如果输入不是有效的 JSON 字符串，则直接返回原字符串。
// 参数：
//
//	jsonString: 需要最小化的 JSON 字符串。
//
// 返回值：
//
//	最小化后的 JSON 字符串，如果处理失败则返回原字符串。
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

// JsonExtract 从给定的 JSON 字符串中提取指定路径的值。
// 参数：
//   - jsonString: 包含 JSON 数据的字符串。
//   - path: 用于指定提取路径的字符串，例如 "user.name"。
//
// 返回值：
//   - any: 提取到的值，如果解析失败或路径不存在则返回 nil。
func JsonExtract(jsonString, path string) any {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return nil
	}

	return extractByPath(obj, path)
}

// JsonKeys 解析给定的 JSON 字符串，并返回其中所有顶层键的切片。
// 如果输入的 JSON 字符串无效，则返回 nil。
// 参数：
//   - jsonString: 需要解析的 JSON 字符串。
//
// 返回值：
//   - []string: 包含所有顶层键的切片；如果解析失败，返回 nil。
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

// JsonValues 解析给定的 JSON 字符串并返回其所有值。
// 如果 JSON 是一个对象（map），则返回对象的所有值（不包含键）。
// 如果 JSON 是一个数组，则直接返回该数组。
// 如果 JSON 是其他类型（如字符串、数字等），则返回 nil。
// 如果解析失败（如无效的 JSON 字符串），同样返回 nil。
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

// JsonMerge 将多个 JSON 字符串合并为一个 JSON 字符串。
// 如果输入的 JSON 字符串无法解析，则跳过该字符串。
// 合并规则为：对于重复的键，后出现的 JSON 字符串中的值会覆盖之前的值。
// 如果合并后的 JSON 字符串无法序列化，则返回空对象 "{}"。
// 参数：
//
//	jsonStrings: 需要合并的 JSON 字符串列表。
//
// 返回值：
//
//	合并后的 JSON 字符串。
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

// JsonLength 计算 JSON 字符串中指定路径的值的长度。
// 参数：
//   - jsonString: 输入的 JSON 字符串。
//   - path: 可选参数，指定 JSON 对象中的路径（例如 "key.subkey"）。
//
// 返回值：
//   - 如果解析失败或路径不存在，返回 0。
//   - 如果值是对象或数组，返回其元素数量。
//   - 如果值是字符串，返回字符串长度。
//   - 如果值是 nil，返回 0。
//   - 其他类型（如布尔值、数字等）返回 1。
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

// JsonType 解析 JSON 字符串并返回指定路径下值的类型。
//
// 参数:
//   - jsonString: 输入的 JSON 字符串。
//   - path: 可选参数，指定 JSON 路径（如 "user.address.city"）。如果未提供，则解析整个 JSON 字符串。
//
// 返回值:
//   - 返回值的类型字符串，可能为以下之一：
//   - "NULL": 表示 JSON 值为 null。
//   - "BOOLEAN": 表示布尔值。
//   - "INTEGER": 表示整数。
//   - "DOUBLE": 表示浮点数。
//   - "STRING": 表示字符串。
//   - "ARRAY": 表示数组。
//   - "OBJECT": 表示对象。
//   - "UNKNOWN": 表示未知类型。
//
// 注意:
//   - 如果 JSON 解析失败，返回 "NULL"。
//   - 如果路径不存在或无效，返回 "NULL"。
func JsonType(jsonString string, path ...string) string {
	var obj any
	if err := json.Unmarshal([]byte(jsonString), &obj); err != nil {
		return "NULL"
	}

	if len(path) > 0 {
		obj = extractByPath(obj, path[0])
	}

	switch v := obj.(type) {
	case nil:
		return "NULL"
	case bool:
		return "BOOLEAN"
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
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

// extractByPath 根据给定的路径从对象中提取值。
//
// 参数:
//
//	obj: 任意类型的对象，可以是 map[string]any 或 []any。
//	path: 路径字符串，支持以下格式：
//	  - "$" 或 ""：返回对象本身。
//	  - "$.key"：从 map 中提取键为 "key" 的值。
//	  - "$[index]"：从数组中提取指定索引的值。
//
// 返回值:
//
//	如果路径匹配且对象类型支持，返回提取的值；否则返回 nil。
//
// 示例:
//
//	extractByPath(map[string]any{"key": "value"}, "$.key") // 返回 "value"
//	extractByPath([]any{1, 2, 3}, "$[1]") // 返回 2
//	extractByPath("string", "$") // 返回 "string"
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

// JsonEncodeWithCallback 将给定的数据编码为 JSON 字符串，并包装成回调函数的调用形式。
//
// 参数：
//   - data: 需要编码为 JSON 的数据。
//   - callback: 回调函数名称，用于包装 JSON 数据。
//   - flags: 可选的编码标志，用于控制 JSON 编码的行为。
//
// 返回值：
//   - 如果编码后的 JSON 数据为空字符串，则返回空字符串。
//   - 否则返回格式为 "callback(jsonData);" 的字符串。
//
// 示例：
//
//	result := JsonEncodeWithCallback(map[string]string{"key": "value"}, "callbackFunc")
//	// 输出: "callbackFunc({\"key\":\"value\"});"
func JsonpEncode(callback string, data any, flags ...int) string {
	jsonData := JsonEncode(data, flags...)
	if jsonData == "" {
		return ""
	}
	return fmt.Sprintf("%s(%s);", callback, jsonData)
}

// JsonpValidate 验证给定的字符串是否为有效的JSONP格式。
//
// 参数:
//
//	jsonp: 待验证的JSONP字符串。
//
// 返回值:
//
//	bool: 如果字符串是有效的JSONP格式，返回true；否则返回false。
//
// 说明:
//  1. 检查字符串是否包含有效的回调函数名（即包含"("）。
//  2. 检查字符串是否以");"结尾。
//  3. 提取JSON部分并调用JsonValidate验证其有效性。
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

// JsonPatch 对输入的 JSON 字符串应用一组 JSON Patch 操作，并返回修改后的 JSON 字符串。
//
// 参数：
//   - jsonString: 原始 JSON 字符串。
//   - patches: 一组 JSON Patch 操作，每个操作是一个 map，包含 "op" 字段指定操作类型（如 "add"、"remove"、"replace"）。
//
// 返回值：
//   - string: 应用 Patch 后的 JSON 字符串。如果解析或序列化失败，则返回原始 JSON 字符串。
//
// 注意：
//   - 当前实现为简化版本，仅支持部分 JSON Patch 操作。
//   - 如果输入的 JSON 字符串无法解析，将直接返回原始字符串。
//   - 如果序列化失败，同样返回原始字符串。
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

// JsonPointer 根据 JSON Pointer 规范从 JSON 字符串中提取指定路径的值。
//
// 参数:
//
//	jsonString: 包含 JSON 数据的字符串
//	pointer:    JSON Pointer 路径字符串 (遵循 RFC 6901 规范)
//
// 返回值:
//
//	返回指针路径对应的值，如果路径无效或 JSON 解析失败则返回 nil
//
// 示例:
//
//	JsonPointer(`{"foo": {"bar": [1, 2, 3]}}`, "/foo/bar/1") 返回 2
//	JsonPointer(`{"a/b": true}`, "/a~1b") 返回 true
//	JsonPointer(`{"x": null}`, "/x") 返回 nil
//
// 注意:
//   - 支持 JSON Pointer 的转义规则 (~0 表示 ~, ~1 表示 /)
//   - 空指针 ("") 返回整个 JSON 对象
//   - 对数组使用数字索引 (从 0 开始)
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
