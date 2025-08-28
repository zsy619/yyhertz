package main

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/mvc/routing"
)

// BenchmarkParamBinderCreation 参数绑定器创建性能测试
func BenchmarkParamBinderCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = routing.NewParamBinder()
	}
}

// BenchmarkTypeConversions 类型转换性能测试
func BenchmarkTypeConversions(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("String转换", func(b *testing.B) {
		targetType := reflect.TypeOf("")
		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			for pb_bench.Next() {
				_, _ = pb.ConvertValue("hello_world", targetType)
			}
		})
	})

	b.Run("Int转换", func(b *testing.B) {
		targetType := reflect.TypeOf(0)
		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			for pb_bench.Next() {
				_, _ = pb.ConvertValue("12345", targetType)
			}
		})
	})

	b.Run("Bool转换", func(b *testing.B) {
		targetType := reflect.TypeOf(false)
		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			for pb_bench.Next() {
				_, _ = pb.ConvertValue("true", targetType)
			}
		})
	})

	b.Run("Float64转换", func(b *testing.B) {
		targetType := reflect.TypeOf(0.0)
		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			for pb_bench.Next() {
				_, _ = pb.ConvertValue("123.456", targetType)
			}
		})
	})

	b.Run("StringSlice转换", func(b *testing.B) {
		targetType := reflect.TypeOf([]string{})
		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			for pb_bench.Next() {
				_, _ = pb.ConvertValue("a,b,c,d,e", targetType)
			}
		})
	})
}

// BenchmarkNumericTypes 数值类型转换性能测试
func BenchmarkNumericTypes(b *testing.B) {
	pb := routing.NewParamBinder()

	benchCases := []struct {
		name       string
		value      string
		targetType reflect.Type
	}{
		{"Int8", "127", reflect.TypeOf(int8(0))},
		{"Int16", "32767", reflect.TypeOf(int16(0))},
		{"Int32", "2147483647", reflect.TypeOf(int32(0))},
		{"Int64", "9223372036854775807", reflect.TypeOf(int64(0))},
		{"Uint8", "255", reflect.TypeOf(uint8(0))},
		{"Uint16", "65535", reflect.TypeOf(uint16(0))},
		{"Uint32", "4294967295", reflect.TypeOf(uint32(0))},
		{"Uint64", "18446744073709551615", reflect.TypeOf(uint64(0))},
		{"Float32", "3.14159", reflect.TypeOf(float32(0))},
		{"Float64", "2.718281828459045", reflect.TypeOf(float64(0))},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = pb.ConvertValue(bc.value, bc.targetType)
			}
		})
	}
}

// BenchmarkPointerConversion 指针类型转换性能测试
func BenchmarkPointerConversion(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("IntPointer_NonEmpty", func(b *testing.B) {
		targetType := reflect.TypeOf((*int)(nil))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("123", targetType)
		}
	})

	b.Run("IntPointer_Empty", func(b *testing.B) {
		targetType := reflect.TypeOf((*int)(nil))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("", targetType)
		}
	})

	b.Run("StringPointer_NonEmpty", func(b *testing.B) {
		targetType := reflect.TypeOf((*string)(nil))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("hello", targetType)
		}
	})
}

// BenchmarkCustomConverters 自定义转换器性能测试
func BenchmarkCustomConverters(b *testing.B) {
	pb := routing.NewParamBinder()

	// 注册自定义转换器
	customConverter := func(value string, targetType reflect.Type) (reflect.Value, error) {
		return reflect.ValueOf("CUSTOM_" + value), nil
	}
	pb.RegisterConverter(reflect.String, customConverter)

	b.Run("CustomStringConverter", func(b *testing.B) {
		targetType := reflect.TypeOf("")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("test", targetType)
		}
	})

	b.Run("DefaultStringConverter", func(b *testing.B) {
		// 创建新的绑定器使用默认转换器
		defaultPB := routing.NewParamBinder()
		targetType := reflect.TypeOf("")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = defaultPB.ConvertValue("test", targetType)
		}
	})
}

// BenchmarkCacheOperations 缓存操作性能测试
func BenchmarkCacheOperations(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("GetCacheStats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pb.GetCacheStats()
		}
	})

	b.Run("ClearCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pb.ClearCache()
		}
	})
}

// BenchmarkConcurrentOperations 并发操作性能测试
func BenchmarkConcurrentOperations(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("并发类型转换", func(b *testing.B) {
		targetType := reflect.TypeOf(0)
		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			for pb_bench.Next() {
				_, _ = pb.ConvertValue("123", targetType)
			}
		})
	})

	b.Run("并发缓存访问", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			for pb_bench.Next() {
				_ = pb.GetCacheStats()
			}
		})
	})

	b.Run("混合并发操作", func(b *testing.B) {
		targetTypes := []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(0),
			reflect.TypeOf(false),
			reflect.TypeOf(0.0),
		}
		values := []string{"hello", "123", "true", "3.14"}

		b.ResetTimer()
		b.RunParallel(func(pb_bench *testing.PB) {
			i := 0
			for pb_bench.Next() {
				idx := i % len(targetTypes)
				_, _ = pb.ConvertValue(values[idx], targetTypes[idx])
				i++
			}
		})
	})
}

// BenchmarkMemoryAllocation 内存分配性能测试
func BenchmarkMemoryAllocation(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("字符串转换内存分配", func(b *testing.B) {
		targetType := reflect.TypeOf("")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("test_string", targetType)
		}
	})

	b.Run("整数转换内存分配", func(b *testing.B) {
		targetType := reflect.TypeOf(0)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("123456", targetType)
		}
	})

	b.Run("切片转换内存分配", func(b *testing.B) {
		targetType := reflect.TypeOf([]string{})
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("a,b,c,d,e,f,g,h,i,j", targetType)
		}
	})
}

// BenchmarkErrorHandling 错误处理性能测试
func BenchmarkErrorHandling(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("有效转换_无错误", func(b *testing.B) {
		targetType := reflect.TypeOf(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("123", targetType)
		}
	})

	b.Run("无效转换_有错误", func(b *testing.B) {
		targetType := reflect.TypeOf(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("invalid_number", targetType)
		}
	})
}

// BenchmarkRealWorldScenarios 真实世界场景性能测试
func BenchmarkRealWorldScenarios(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("用户注册场景", func(b *testing.B) {
		// 模拟用户注册时的多个参数转换
		scenarios := []struct {
			value      string
			targetType reflect.Type
		}{
			{"john_doe", reflect.TypeOf("")},           // 用户名
			{"25", reflect.TypeOf(0)},                  // 年龄
			{"true", reflect.TypeOf(false)},            // 是否激活
			{"admin,user", reflect.TypeOf([]string{})}, // 角色列表
			{"john@example.com", reflect.TypeOf("")},   // 邮箱
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, scenario := range scenarios {
				_, _ = pb.ConvertValue(scenario.value, scenario.targetType)
			}
		}
	})

	b.Run("分页查询场景", func(b *testing.B) {
		// 模拟分页查询的参数转换
		pagingScenarios := []struct {
			value      string
			targetType reflect.Type
		}{
			{"1", reflect.TypeOf(0)},                    // page
			{"20", reflect.TypeOf(0)},                   // size
			{"created_at", reflect.TypeOf("")},          // sort
			{"desc", reflect.TypeOf("")},                // order
			{"user,active", reflect.TypeOf([]string{})}, // filters
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, scenario := range pagingScenarios {
				_, _ = pb.ConvertValue(scenario.value, scenario.targetType)
			}
		}
	})

	b.Run("搜索过滤场景", func(b *testing.B) {
		// 模拟复杂搜索过滤的参数转换
		filterScenarios := []struct {
			value      string
			targetType reflect.Type
		}{
			{"john", reflect.TypeOf("")},                   // 关键字
			{"18", reflect.TypeOf(0)},                      // 最小年龄
			{"65", reflect.TypeOf(0)},                      // 最大年龄
			{"true", reflect.TypeOf(false)},                // 只显示活跃用户
			{"tag1,tag2,tag3", reflect.TypeOf([]string{})}, // 标签过滤
			{"2023-01-01", reflect.TypeOf("")},             // 开始日期
			{"2023-12-31", reflect.TypeOf("")},             // 结束日期
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, scenario := range filterScenarios {
				_, _ = pb.ConvertValue(scenario.value, scenario.targetType)
			}
		}
	})
}

// BenchmarkComparisonWithReflection 与直接反射对比测试
func BenchmarkComparisonWithReflection(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("ParamBinder_Int转换", func(b *testing.B) {
		targetType := reflect.TypeOf(0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue("123", targetType)
		}
	})

	b.Run("直接反射_Int转换", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 模拟直接使用反射的方式
			targetType := reflect.TypeOf(0)
			_ = targetType
			value := reflect.ValueOf(123)
			_ = value
		}
	})
}

// BenchmarkObjectPooling 对象池化效果测试
func BenchmarkObjectPooling(b *testing.B) {
	// 这个测试主要是演示概念，实际的对象池在PrepareMethodArgs中

	var pool sync.Pool
	pool.New = func() any {
		return make([]reflect.Value, 0, 8)
	}

	b.Run("使用对象池", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := pool.Get().([]reflect.Value)
			slice = slice[:0] // 重置但保留容量
			slice = append(slice, reflect.ValueOf(123))
			pool.Put(slice)
		}
	})

	b.Run("直接分配", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := make([]reflect.Value, 0, 8)
			slice = append(slice, reflect.ValueOf(123))
			_ = slice
		}
	})
}

// BenchmarkParameterValidation 参数验证性能测试
func BenchmarkParameterValidation(b *testing.B) {
	pb := routing.NewParamBinder()

	// 创建模拟的RequestContext
	ctx := &app.RequestContext{}
	ctx.Request.SetRequestURI("/test?name=john&age=25&active=true")
	ctx.Request.Header.SetMethod("GET")

	// 创建参数配置
	params := []*routing.ParamInfo{
		{Name: "name", Source: routing.ParamSourceQuery, Required: true},
		{Name: "age", Source: routing.ParamSourceQuery, Required: true},
		{Name: "active", Source: routing.ParamSourceQuery, Required: false},
		{Name: "optional", Source: routing.ParamSourceQuery, Required: false},
	}

	b.Run("参数验证", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pb.ValidateParams(params, ctx)
		}
	})
}

// BenchmarkComplexConversions 复杂转换性能测试
func BenchmarkComplexConversions(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("长字符串处理", func(b *testing.B) {
		longString := "this_is_a_very_long_string_that_needs_to_be_processed_and_converted_to_demonstrate_performance_characteristics_of_string_handling"
		targetType := reflect.TypeOf("")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue(longString, targetType)
		}
	})

	b.Run("大数值处理", func(b *testing.B) {
		bigNumber := "999999999999999999"
		targetType := reflect.TypeOf(int64(0))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue(bigNumber, targetType)
		}
	})

	b.Run("长切片处理", func(b *testing.B) {
		longSlice := "item1,item2,item3,item4,item5,item6,item7,item8,item9,item10,item11,item12,item13,item14,item15,item16,item17,item18,item19,item20"
		targetType := reflect.TypeOf([]string{})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pb.ConvertValue(longSlice, targetType)
		}
	})
}

// 性能对比辅助函数
func BenchmarkPerformanceComparison(b *testing.B) {
	b.Run("性能总结", func(b *testing.B) {
		// 这个测试用于显示关键指标，不实际运行代码
		b.Skip("这是一个性能总结测试，查看其他基准测试的结果")
	})
}

// 内存使用情况测试
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("整体内存使用", func(b *testing.B) {
		pb := routing.NewParamBinder()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// 执行一系列典型操作
			_, _ = pb.ConvertValue("test", reflect.TypeOf(""))
			_, _ = pb.ConvertValue("123", reflect.TypeOf(0))
			_, _ = pb.ConvertValue("true", reflect.TypeOf(false))
			_ = pb.GetCacheStats()
		}
	})
}

// 长期运行稳定性测试
func BenchmarkLongTermStability(b *testing.B) {
	pb := routing.NewParamBinder()

	b.Run("长期运行稳定性", func(b *testing.B) {
		targetTypes := []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(0),
			reflect.TypeOf(false),
			reflect.TypeOf([]string{}),
		}
		values := []string{"hello", "123", "true", "a,b,c"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % len(targetTypes)
			_, _ = pb.ConvertValue(values[idx], targetTypes[idx])

			// 偶尔清理缓存，模拟长期运行
			if i%1000 == 0 {
				pb.ClearCache()
			}
		}
	})
}
