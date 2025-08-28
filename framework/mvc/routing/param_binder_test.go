package routing

import (
	"reflect"
	"testing"
)

// 测试用的控制器
type TestController struct{}

func (tc *TestController) TestMethod(id int, name string) string {
	return name
}

func TestNewParamBinder(t *testing.T) {
	pb := NewParamBinder()
	
	// 验证基本结构
	if pb == nil {
		t.Fatal("ParamBinder should not be nil")
	}
	if pb.methodCache == nil {
		t.Fatal("methodCache should not be nil")
	}
	if pb.converters == nil {
		t.Fatal("converters should not be nil")
	}
	if pb.extractors == nil {
		t.Fatal("extractors should not be nil")
	}
	
	// 验证默认转换器已注册
	if len(pb.converters) == 0 {
		t.Fatal("converters should be registered")
	}
	if len(pb.extractors) == 0 {
		t.Fatal("extractors should be registered")
	}
	
	// 验证缓存统计
	stats := pb.GetCacheStats()
	if stats["method_cache_size"] != 0 {
		t.Errorf("Expected method_cache_size to be 0, got %d", stats["method_cache_size"])
	}
	if stats["converters_count"] == 0 {
		t.Fatal("converters_count should be greater than 0")
	}
	if stats["extractors_count"] == 0 {
		t.Fatal("extractors_count should be greater than 0")
	}
}

func TestConvertValue(t *testing.T) {
	pb := NewParamBinder()
	
	testCases := []struct {
		name        string
		value       string
		targetType  reflect.Type
		expected    interface{}
		shouldError bool
	}{
		{
			name:       "string conversion",
			value:      "hello",
			targetType: reflect.TypeOf(""),
			expected:   "hello",
		},
		{
			name:       "int conversion",
			value:      "123",
			targetType: reflect.TypeOf(0),
			expected:   123,
		},
		{
			name:       "empty int conversion",
			value:      "",
			targetType: reflect.TypeOf(0),
			expected:   0,
		},
		{
			name:       "bool conversion true",
			value:      "true",
			targetType: reflect.TypeOf(false),
			expected:   true,
		},
		{
			name:       "bool conversion false",
			value:      "false",
			targetType: reflect.TypeOf(false),
			expected:   false,
		},
		{
			name:       "float64 conversion",
			value:      "123.45",
			targetType: reflect.TypeOf(0.0),
			expected:   123.45,
		},
		{
			name:       "string slice conversion",
			value:      "a,b,c",
			targetType: reflect.TypeOf([]string{}),
			expected:   []string{"a", "b", "c"},
		},
		{
			name:        "invalid int conversion",
			value:       "abc",
			targetType:  reflect.TypeOf(0),
			shouldError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := pb.ConvertValue(tc.value, tc.targetType)
			
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if !reflect.DeepEqual(tc.expected, result.Interface()) {
				t.Errorf("Expected %v, got %v", tc.expected, result.Interface())
			}
		})
	}
}

func TestCustomConverterRegistration(t *testing.T) {
	pb := NewParamBinder()
	
	// 注册自定义转换器
	customConverter := func(value string, targetType reflect.Type) (reflect.Value, error) {
		return reflect.ValueOf("custom:" + value), nil
	}
	
	pb.RegisterConverter(reflect.String, customConverter)
	
	// 测试自定义转换器
	result, err := pb.ConvertValue("test", reflect.TypeOf(""))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	expected := "custom:test"
	if result.Interface() != expected {
		t.Errorf("Expected %s, got %v", expected, result.Interface())
	}
}

func TestCacheManagement(t *testing.T) {
	pb := NewParamBinder()
	
	// 初始缓存应该为空
	stats1 := pb.GetCacheStats()
	if stats1["method_cache_size"] != 0 {
		t.Errorf("Initial cache size should be 0, got %d", stats1["method_cache_size"])
	}
	
	// 模拟添加缓存项
	pb.methodCache["test.method"] = reflect.Method{}
	
	stats2 := pb.GetCacheStats()
	if stats2["method_cache_size"] != 1 {
		t.Errorf("Cache size should be 1 after adding item, got %d", stats2["method_cache_size"])
	}
	
	// 测试清空缓存
	pb.ClearCache()
	stats3 := pb.GetCacheStats()
	if stats3["method_cache_size"] != 0 {
		t.Errorf("Cache size should be 0 after clearing, got %d", stats3["method_cache_size"])
	}
}

func BenchmarkConvertValueString(b *testing.B) {
	pb := NewParamBinder()
	targetType := reflect.TypeOf("")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pb.ConvertValue("test", targetType)
	}
}

func BenchmarkConvertValueInt(b *testing.B) {
	pb := NewParamBinder()
	targetType := reflect.TypeOf(0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pb.ConvertValue("123", targetType)
	}
}

func BenchmarkConvertValueBool(b *testing.B) {
	pb := NewParamBinder()
	targetType := reflect.TypeOf(false)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pb.ConvertValue("true", targetType)
	}
}