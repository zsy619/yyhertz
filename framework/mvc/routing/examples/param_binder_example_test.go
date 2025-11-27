package main

import (
	"reflect"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/mvc/routing"
)

// TestParamBinderCreation 测试参数绑定器的创建
func TestParamBinderCreation(t *testing.T) {
	pb := routing.NewParamBinder()

	if pb == nil {
		t.Fatal("参数绑定器创建失败")
	}

	stats := pb.GetCacheStats()
	if stats["method_cache_size"] != 0 {
		t.Errorf("初始方法缓存大小应为0，实际为 %d", stats["method_cache_size"])
	}

	if stats["converters_count"] == 0 {
		t.Error("应该有默认的类型转换器")
	}

	if stats["extractors_count"] == 0 {
		t.Error("应该有默认的参数提取器")
	}

	t.Logf("✅ 参数绑定器创建成功，缓存统计: %+v", stats)
}

// TestBasicTypeConversion 测试基本类型转换
func TestBasicTypeConversion(t *testing.T) {
	pb := routing.NewParamBinder()

	testCases := []struct {
		name        string
		value       string
		targetType  reflect.Type
		expected    any
		shouldError bool
	}{
		{"字符串转换", "hello world", reflect.TypeOf(""), "hello world", false},
		{"空字符串", "", reflect.TypeOf(""), "", false},

		{"整数转换", "123", reflect.TypeOf(0), 123, false},
		{"空整数", "", reflect.TypeOf(0), 0, false},
		{"负整数", "-456", reflect.TypeOf(0), -456, false},
		{"无效整数", "abc", reflect.TypeOf(0), nil, true},

		{"布尔真值", "true", reflect.TypeOf(false), true, false},
		{"布尔假值", "false", reflect.TypeOf(false), false, false},
		{"布尔1", "1", reflect.TypeOf(false), true, false},
		{"布尔0", "0", reflect.TypeOf(false), false, false},
		{"空布尔", "", reflect.TypeOf(false), false, false},
		{"无效布尔", "invalid", reflect.TypeOf(false), nil, true},

		{"浮点数", "123.45", reflect.TypeOf(0.0), 123.45, false},
		{"科学记数法", "1.23e2", reflect.TypeOf(0.0), 123.0, false},
		{"空浮点数", "", reflect.TypeOf(0.0), 0.0, false},
		{"无效浮点数", "not_a_number", reflect.TypeOf(0.0), nil, true},

		{"字符串切片", "a,b,c", reflect.TypeOf([]string{}), []string{"a", "b", "c"}, false},
		{"带空格的切片", "a, b , c", reflect.TypeOf([]string{}), []string{"a", "b", "c"}, false},
		{"空切片", "", reflect.TypeOf([]string{}), []string{}, false},
		{"单元素切片", "single", reflect.TypeOf([]string{}), []string{"single"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := pb.ConvertValue(tc.value, tc.targetType)

			if tc.shouldError {
				if err == nil {
					t.Errorf("期望出现错误但没有，值: %s", tc.value)
				}
				return
			}

			if err != nil {
				t.Errorf("不期望的错误: %v", err)
				return
			}

			if !reflect.DeepEqual(tc.expected, result.Interface()) {
				t.Errorf("转换结果不匹配。期望: %v (%T)，实际: %v (%T)",
					tc.expected, tc.expected, result.Interface(), result.Interface())
			}
		})
	}
}

// TestNumericTypeConversion 测试各种数值类型转换
func TestNumericTypeConversion(t *testing.T) {
	pb := routing.NewParamBinder()

	testCases := []struct {
		name       string
		value      string
		targetType reflect.Type
		expected   any
	}{
		// 整数类型
		{"int8", "127", reflect.TypeOf(int8(0)), int8(127)},
		{"int16", "32767", reflect.TypeOf(int16(0)), int16(32767)},
		{"int32", "2147483647", reflect.TypeOf(int32(0)), int32(2147483647)},
		{"int64", "9223372036854775807", reflect.TypeOf(int64(0)), int64(9223372036854775807)},

		// 无符号整数类型
		{"uint8", "255", reflect.TypeOf(uint8(0)), uint8(255)},
		{"uint16", "65535", reflect.TypeOf(uint16(0)), uint16(65535)},
		{"uint32", "4294967295", reflect.TypeOf(uint32(0)), uint32(4294967295)},
		{"uint64", "18446744073709551615", reflect.TypeOf(uint64(0)), uint64(18446744073709551615)},

		// 浮点数类型
		{"float32", "3.14", reflect.TypeOf(float32(0)), float32(3.14)},
		{"float64", "2.718281828", reflect.TypeOf(float64(0)), float64(2.718281828)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := pb.ConvertValue(tc.value, tc.targetType)
			if err != nil {
				t.Errorf("转换失败: %v", err)
				return
			}

			if !reflect.DeepEqual(tc.expected, result.Interface()) {
				t.Errorf("转换结果不匹配。期望: %v (%T)，实际: %v (%T)",
					tc.expected, tc.expected, result.Interface(), result.Interface())
			}
		})
	}
}

// TestPointerTypeConversion 测试指针类型转换
func TestPointerTypeConversion(t *testing.T) {
	pb := routing.NewParamBinder()

	// 测试非空值的指针转换
	result, err := pb.ConvertValue("123", reflect.TypeOf((*int)(nil)))
	if err != nil {
		t.Errorf("指针类型转换失败: %v", err)
		return
	}

	if result.Kind() != reflect.Ptr {
		t.Errorf("结果应该是指针类型，实际为: %v", result.Kind())
		return
	}

	if result.Elem().Interface() != 123 {
		t.Errorf("指针指向的值不正确。期望: 123，实际: %v", result.Elem().Interface())
	}

	// 测试空值的指针转换
	result, err = pb.ConvertValue("", reflect.TypeOf((*int)(nil)))
	if err != nil {
		t.Errorf("空值指针转换失败: %v", err)
		return
	}

	if !result.IsNil() {
		t.Error("空值应该转换为nil指针")
	}

	t.Log("✅ 指针类型转换测试通过")
}

// TestCustomConverters 测试自定义转换器
func TestCustomConverters(t *testing.T) {
	pb := routing.NewParamBinder()

	// 注册自定义字符串转换器
	customConverter := func(value string, targetType reflect.Type) (reflect.Value, error) {
		return reflect.ValueOf("CUSTOM_" + value), nil
	}

	pb.RegisterConverter(reflect.String, customConverter)

	result, err := pb.ConvertValue("test", reflect.TypeOf(""))
	if err != nil {
		t.Errorf("自定义转换器执行失败: %v", err)
		return
	}

	expected := "CUSTOM_test"
	if result.Interface() != expected {
		t.Errorf("自定义转换结果不匹配。期望: %s，实际: %v", expected, result.Interface())
	}

	t.Log("✅ 自定义转换器测试通过")
}

// TestParameterExtractors 测试参数提取器
func TestParameterExtractors(t *testing.T) {
	// 创建模拟的RequestContext
	ctx := &app.RequestContext{}
	ctx.Request.SetRequestURI("/test?name=john&age=25")
	ctx.Request.Header.Set("Authorization", "Bearer token123")
	ctx.Request.Header.Set("Content-Type", "application/json")

	// 注意：由于无法直接访问私有的extractors，我们测试基本的参数提取逻辑
	t.Run("查询参数提取", func(t *testing.T) {
		value := ctx.Query("name")
		expected := "john"
		if value != expected {
			t.Errorf("查询参数提取失败。期望: %s，实际: %s", expected, value)
		}
	})

	t.Run("请求头提取", func(t *testing.T) {
		value := string(ctx.GetHeader("Authorization"))
		expected := "Bearer token123"
		if value != expected {
			t.Errorf("请求头提取失败。期望: %s，实际: %s", expected, value)
		}
	})

	t.Run("不存在的参数", func(t *testing.T) {
		value := ctx.Query("nonexistent")
		if value != "" {
			t.Errorf("不存在的参数应返回空字符串，实际: %s", value)
		}
	})

	t.Log("✅ 参数提取器基础测试通过")
}

// TestCacheManagement 测试缓存管理
func TestCacheManagement(t *testing.T) {
	pb := routing.NewParamBinder()

	// 初始状态
	stats := pb.GetCacheStats()
	if stats["method_cache_size"] != 0 {
		t.Errorf("初始缓存大小应为0，实际为: %d", stats["method_cache_size"])
	}

	// 模拟添加缓存项（通过内部访问）
	// 注意：这里我们不能直接访问私有字段，所以我们测试清空功能
	pb.ClearCache()

	stats = pb.GetCacheStats()
	if stats["method_cache_size"] != 0 {
		t.Errorf("清空后缓存大小应为0，实际为: %d", stats["method_cache_size"])
	}

	t.Logf("✅ 缓存管理测试通过，最终统计: %+v", stats)
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	pb := routing.NewParamBinder()

	errorCases := []struct {
		name       string
		value      string
		targetType reflect.Type
		expectErr  bool
	}{
		{"整数转换错误", "not_a_number", reflect.TypeOf(0), true},
		{"浮点数转换错误", "invalid_float", reflect.TypeOf(0.0), true},
		{"布尔值转换错误", "invalid_bool", reflect.TypeOf(false), true},
		{"有效转换", "123", reflect.TypeOf(0), false},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pb.ConvertValue(tc.value, tc.targetType)

			hasErr := err != nil
			if hasErr != tc.expectErr {
				if tc.expectErr {
					t.Errorf("期望出现错误但没有，值: %s", tc.value)
				} else {
					t.Errorf("不期望的错误: %v", err)
				}
			}
		})
	}

	t.Log("✅ 错误处理测试通过")
}

// TestParameterValidation 测试参数验证
func TestParameterValidation(t *testing.T) {
	pb := routing.NewParamBinder()

	// 创建模拟的RequestContext
	ctx := &app.RequestContext{}
	ctx.Request.SetRequestURI("/test?name=john")
	ctx.Request.Header.SetMethod("GET")

	testCases := []struct {
		name        string
		params      []*routing.ParamInfo
		shouldError bool
	}{
		{
			"有效的必需参数",
			[]*routing.ParamInfo{
				{Name: "name", Source: routing.ParamSourceQuery, Required: true},
			},
			false,
		},
		{
			"缺失的必需参数",
			[]*routing.ParamInfo{
				{Name: "missing", Source: routing.ParamSourceQuery, Required: true},
			},
			true,
		},
		{
			"可选参数",
			[]*routing.ParamInfo{
				{Name: "optional", Source: routing.ParamSourceQuery, Required: false},
			},
			false,
		},
		{
			"带默认值的参数",
			[]*routing.ParamInfo{
				{Name: "missing", Source: routing.ParamSourceQuery, Required: true, DefaultValue: "default"},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := pb.ValidateParams(tc.params, ctx)

			hasErr := err != nil
			if hasErr != tc.shouldError {
				if tc.shouldError {
					t.Error("期望验证失败但通过了")
				} else {
					t.Errorf("不期望的验证错误: %v", err)
				}
			}
		})
	}

	t.Log("✅ 参数验证测试通过")
}

// TestRealWorldScenarios 测试真实世界场景
func TestRealWorldScenarios(t *testing.T) {
	pb := routing.NewParamBinder()

	t.Run("用户注册场景", func(t *testing.T) {
		// 模拟用户注册请求的参数类型
		testCases := []struct {
			param    string
			value    string
			expected any
		}{
			{"用户名", "john_doe", "john_doe"},
			{"年龄", "25", 25},
			{"邮件验证", "true", true},
			{"标签", "admin,user,vip", []string{"admin", "user", "vip"}},
		}

		for _, tc := range testCases {
			var targetType reflect.Type
			switch tc.expected.(type) {
			case string:
				targetType = reflect.TypeOf("")
			case int:
				targetType = reflect.TypeOf(0)
			case bool:
				targetType = reflect.TypeOf(false)
			case []string:
				targetType = reflect.TypeOf([]string{})
			}

			result, err := pb.ConvertValue(tc.value, targetType)
			if err != nil {
				t.Errorf("%s 转换失败: %v", tc.param, err)
				continue
			}

			if !reflect.DeepEqual(tc.expected, result.Interface()) {
				t.Errorf("%s 结果不匹配。期望: %v，实际: %v",
					tc.param, tc.expected, result.Interface())
			}
		}
	})

	t.Run("分页查询场景", func(t *testing.T) {
		pagingParams := map[string]string{
			"page":  "2",
			"size":  "20",
			"sort":  "created_at",
			"order": "desc",
		}

		for param, value := range pagingParams {
			var targetType reflect.Type
			if param == "page" || param == "size" {
				targetType = reflect.TypeOf(0)
			} else {
				targetType = reflect.TypeOf("")
			}

			result, err := pb.ConvertValue(value, targetType)
			if err != nil {
				t.Errorf("分页参数 %s 转换失败: %v", param, err)
			} else {
				t.Logf("✅ %s: %s -> %v", param, value, result.Interface())
			}
		}
	})

	t.Log("✅ 真实世界场景测试通过")
}

// TestConcurrentSafety 测试并发安全性
func TestConcurrentSafety(t *testing.T) {
	pb := routing.NewParamBinder()

	const numGoroutines = 100
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// 启动多个goroutine同时进行类型转换
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// 测试不同类型的转换
				_, err := pb.ConvertValue("123", reflect.TypeOf(0))
				if err != nil {
					t.Errorf("Goroutine %d 转换失败: %v", id, err)
					return
				}

				// 测试缓存操作
				_ = pb.GetCacheStats()

				// 测试清空缓存（偶尔）
				if j%10 == 0 {
					pb.ClearCache()
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	t.Log("✅ 并发安全性测试通过")
}

// BenchmarkBasicConversion 基础转换性能基准测试
func BenchmarkBasicConversion(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("String", func(b *testing.B) {
		targetType := reflect.TypeOf("")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("test", targetType)
		}
	})

	b.Run("Int", func(b *testing.B) {
		targetType := reflect.TypeOf(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("123", targetType)
		}
	})

	b.Run("Bool", func(b *testing.B) {
		targetType := reflect.TypeOf(false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("true", targetType)
		}
	})

	b.Run("Float", func(b *testing.B) {
		targetType := reflect.TypeOf(0.0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("123.45", targetType)
		}
	})

	b.Run("StringSlice", func(b *testing.B) {
		targetType := reflect.TypeOf([]string{})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("a,b,c", targetType)
		}
	})
}
