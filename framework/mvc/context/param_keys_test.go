package context

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestParamKeys(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试空params
	keys := ctx.ParamKeys()
	if keys != nil {
		t.Errorf("Expected nil for empty params, got %v", keys)
	}

	// 设置测试参数
	testParams := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
		{Key: "category", Value: "api"},
	}
	ctx.SetParams(testParams)

	// 测试获取keys
	keys = ctx.ParamKeys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	expectedKeys := []string{"id", "name", "category"}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s, got %s", expectedKeys[i], key)
		}
	}

	// 测试顺序保持
	if keys[0] != "id" || keys[1] != "name" || keys[2] != "category" {
		t.Errorf("Keys order not preserved: %v", keys)
	}
}

func TestParamKeysEmpty(t *testing.T) {
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试空Params
	keys := ctx.ParamKeys()
	if keys != nil {
		t.Errorf("Expected nil for empty params, got %v", keys)
	}

	// 设置空的params slice
	ctx.SetParams(Params{})
	keys = ctx.ParamKeys()
	if keys != nil {
		t.Errorf("Expected nil for empty params slice, got %v", keys)
	}
}

func TestParamMap(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试空params
	paramMap := ctx.ParamMap()
	if paramMap == nil {
		t.Error("Expected non-nil map for empty params")
	}
	if len(paramMap) != 0 {
		t.Errorf("Expected empty map for empty params, got %v", paramMap)
	}

	// 设置测试参数
	testParams := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
		{Key: "category", Value: "api"},
	}
	ctx.SetParams(testParams)

	// 测试转换为map
	paramMap = ctx.ParamMap()
	if len(paramMap) != 3 {
		t.Errorf("Expected map with 3 entries, got %d", len(paramMap))
	}

	// 验证map内容
	expectedMap := map[string]string{
		"id":       "123",
		"name":     "test",
		"category": "api",
	}

	for key, expectedValue := range expectedMap {
		if value, exists := paramMap[key]; !exists {
			t.Errorf("Expected key %s not found in param map", key)
		} else if value != expectedValue {
			t.Errorf("Expected value %s for key %s, got %s", expectedValue, key, value)
		}
	}

	// 验证不存在额外的键
	if len(paramMap) != len(expectedMap) {
		t.Errorf("Map contains unexpected keys: %v", paramMap)
	}
}

func TestParamMapEmpty(t *testing.T) {
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试空Params
	paramMap := ctx.ParamMap()
	if paramMap == nil {
		t.Error("Expected non-nil map for empty params")
	}
	if len(paramMap) != 0 {
		t.Errorf("Expected empty map, got %v", paramMap)
	}

	// 设置空的params slice
	ctx.SetParams(Params{})
	paramMap = ctx.ParamMap()
	if paramMap == nil {
		t.Error("Expected non-nil map for empty params slice")
	}
	if len(paramMap) != 0 {
		t.Errorf("Expected empty map, got %v", paramMap)
	}
}

func TestParamMapDuplicateKeys(t *testing.T) {
	// 测试重复键的情况（虽然在正常路由中不会出现，但要保证健壮性）
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 设置包含重复键的参数（后面的值会覆盖前面的值）
	testParams := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "first"},
		{Key: "id", Value: "456"}, // 重复的键
		{Key: "name", Value: "second"}, // 重复的键
	}
	ctx.SetParams(testParams)

	paramMap := ctx.ParamMap()
	
	// 验证最后的值被保留
	if paramMap["id"] != "456" {
		t.Errorf("Expected id=456 (last value), got %s", paramMap["id"])
	}
	if paramMap["name"] != "second" {
		t.Errorf("Expected name=second (last value), got %s", paramMap["name"])
	}
}

func TestParamValues(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试空params
	values := ctx.ParamValues()
	if values != nil {
		t.Errorf("Expected nil for empty params, got %v", values)
	}

	// 设置测试参数
	testParams := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
		{Key: "category", Value: "api"},
	}
	ctx.SetParams(testParams)

	// 测试获取values
	values = ctx.ParamValues()
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	expectedValues := []string{"123", "test", "api"}
	for i, value := range values {
		if value != expectedValues[i] {
			t.Errorf("Expected value %s at index %d, got %s", expectedValues[i], i, value)
		}
	}

	// 测试顺序保持
	if values[0] != "123" || values[1] != "test" || values[2] != "api" {
		t.Errorf("Values order not preserved: %v", values)
	}
}

func TestParamValuesEmpty(t *testing.T) {
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试空Params
	values := ctx.ParamValues()
	if values != nil {
		t.Errorf("Expected nil for empty params, got %v", values)
	}

	// 设置空的params slice
	ctx.SetParams(Params{})
	values = ctx.ParamValues()
	if values != nil {
		t.Errorf("Expected nil for empty params slice, got %v", values)
	}
}

func TestParamValuesWithDuplicateKeys(t *testing.T) {
	// 测试重复键的情况，验证所有值都被包含
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 设置包含重复键的参数
	testParams := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "first"},
		{Key: "id", Value: "456"}, // 重复的键，但值不同
		{Key: "category", Value: "api"},
	}
	ctx.SetParams(testParams)

	values := ctx.ParamValues()
	
	// 验证所有值都被包含（按顺序）
	expectedValues := []string{"123", "first", "456", "api"}
	if len(values) != len(expectedValues) {
		t.Errorf("Expected %d values, got %d", len(expectedValues), len(values))
	}

	for i, expected := range expectedValues {
		if i >= len(values) || values[i] != expected {
			t.Errorf("Expected value %s at index %d, got %s", expected, i, values[i])
		}
	}
}

func TestParamMethodsConsistency(t *testing.T) {
	// 测试所有param相关方法的一致性
	c := &app.RequestContext{}
	ctx := NewContext(c)

	testParams := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
		{Key: "category", Value: "api"},
	}
	ctx.SetParams(testParams)

	// 获取所有相关数据
	keys := ctx.ParamKeys()
	values := ctx.ParamValues()
	paramMap := ctx.ParamMap()
	params := ctx.Params()

	// 验证长度一致性
	if len(keys) != len(values) || len(keys) != len(paramMap) || len(keys) != len(params) {
		t.Errorf("Length mismatch: keys=%d, values=%d, map=%d, params=%d", 
			len(keys), len(values), len(paramMap), len(params))
	}

	// 验证内容一致性
	for i, key := range keys {
		// 验证key-value对应关系
		if paramMap[key] != values[i] {
			t.Errorf("Inconsistency at index %d: key=%s, mapValue=%s, sliceValue=%s", 
				i, key, paramMap[key], values[i])
		}
		
		// 验证与原始params的一致性
		if params[i].Key != key || params[i].Value != values[i] {
			t.Errorf("Inconsistency with original params at index %d", i)
		}
	}
}