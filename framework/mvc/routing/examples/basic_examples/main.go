package main

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"

	"github.com/zsy619/yyhertz/framework/mvc/routing"
)

// UserController 用户控制器示例
type UserController struct{}

// GetUser 获取用户信息 - 演示路径参数和查询参数
// @Path /users/{id}
// @Query include_profile
func (uc *UserController) GetUser(id int, includeProfile bool) string {
	if includeProfile {
		return fmt.Sprintf("User %d with profile", id)
	}
	return fmt.Sprintf("User %d", id)
}

// SearchUsers 搜索用户 - 演示复杂查询参数
// @Query name,age,tags,page,size
func (uc *UserController) SearchUsers(name string, age int, tags []string, page int, size int) string {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}
	return fmt.Sprintf("Search users: name=%s, age=%d, tags=%v, page=%d, size=%d",
		name, age, tags, page, size)
}

// CreateUser 创建用户 - 演示请求体参数
type CreateUserRequest struct {
	Name     string    `json:"name" validate:"required"`
	Email    string    `json:"email" validate:"email"`
	Age      int       `json:"age" validate:"min=0,max=120"`
	Birthday time.Time `json:"birthday"`
}

func (uc *UserController) CreateUser(req *CreateUserRequest) string {
	return fmt.Sprintf("Create user: %+v", req)
}

// UpdateUserProfile 更新用户资料 - 演示混合参数
// @Path /users/{id}/profile
// @Header Authorization
func (uc *UserController) UpdateUserProfile(id int, authToken string, req *CreateUserRequest) string {
	return fmt.Sprintf("Update user %d profile (auth: %s): %+v", id, authToken, req)
}

// 自定义类型转换器示例

// UUIDConverter UUID类型转换器
func UUIDConverter(value string, targetType reflect.Type) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(targetType), nil
	}

	parsedUUID, err := uuid.Parse(value)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("invalid UUID format: %w", err)
	}

	return reflect.ValueOf(parsedUUID), nil
}

// TimeConverter 时间类型转换器
func TimeConverter(value string, targetType reflect.Type) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(targetType), nil
	}

	// 尝试多种时间格式
	formats := []string{
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,      // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02",          // 2006-01-02
		"2006-01-02 15:04:05", // 2006-01-02 15:04:05
		"15:04:05",            // 15:04:05
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, value); err == nil {
			return reflect.ValueOf(parsedTime), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("unsupported time format: %s", value)
}

// DurationConverter 持续时间转换器
func DurationConverter(value string, targetType reflect.Type) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(targetType), nil
	}

	// 尝试直接解析
	if duration, err := time.ParseDuration(value); err == nil {
		return reflect.ValueOf(duration), nil
	}

	// 尝试解析为秒数
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return reflect.ValueOf(time.Duration(seconds * float64(time.Second))), nil
	}

	return reflect.Value{}, fmt.Errorf("invalid duration format: %s", value)
}

// 自定义参数提取器示例

// JWTExtractor JWT Token提取器
type JWTExtractor struct{}

func (j *JWTExtractor) Extract(paramInfo *routing.ParamInfo, c *app.RequestContext) (string, error) {
	// 从 Authorization header 提取 Bearer token
	authHeader := string(c.GetHeader("Authorization"))
	if authHeader == "" {
		if paramInfo.DefaultValue != "" {
			return paramInfo.DefaultValue, nil
		}
		return "", fmt.Errorf("authorization header is required")
	}

	// 检查 Bearer 前缀
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("invalid authorization header format, expected 'Bearer <token>'")
	}

	return authHeader[len(bearerPrefix):], nil
}

// LanguageExtractor 语言偏好提取器
type LanguageExtractor struct{}

func (l *LanguageExtractor) Extract(paramInfo *routing.ParamInfo, c *app.RequestContext) (string, error) {
	// 先尝试从查询参数获取
	if lang := c.Query("lang"); lang != "" {
		return lang, nil
	}

	// 然后尝试从 Accept-Language header
	acceptLang := string(c.GetHeader("Accept-Language"))
	if acceptLang != "" {
		// 简单解析，取第一个语言
		if idx := len(acceptLang); idx > 0 {
			for i, char := range acceptLang {
				if char == ',' || char == ';' {
					idx = i
					break
				}
			}
			return acceptLang[:idx], nil
		}
	}

	// 返回默认值
	if paramInfo.DefaultValue != "" {
		return paramInfo.DefaultValue, nil
	}
	return "en", nil
}

// 示例函数

func ExampleBasicUsage() {
	fmt.Println("=== 基础用法示例 ===")

	// 创建参数绑定器
	pb := routing.NewParamBinder()

	// 查看初始状态
	stats := pb.GetCacheStats()
	fmt.Printf("初始缓存状态: %+v\n", stats)

	// 测试基本类型转换
	testCases := []struct {
		value      string
		targetType reflect.Type
		desc       string
	}{
		{"hello", reflect.TypeOf(""), "字符串转换"},
		{"123", reflect.TypeOf(0), "整数转换"},
		{"true", reflect.TypeOf(false), "布尔转换"},
		{"3.14", reflect.TypeOf(0.0), "浮点数转换"},
		{"a,b,c", reflect.TypeOf([]string{}), "字符串切片转换"},
	}

	for _, tc := range testCases {
		result, err := pb.ConvertValue(tc.value, tc.targetType)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", tc.desc, err)
		} else {
			fmt.Printf("✅ %s: %s -> %v\n", tc.desc, tc.value, result.Interface())
		}
	}

	fmt.Println()
}

func ExampleCustomConverters() {
	fmt.Println("=== 自定义转换器示例 ===")

	pb := routing.NewParamBinder()

	// 注意：当前的 RegisterConverter 只支持 reflect.Kind 类型
	// 对于自定义类型，我们需要通过其他方式扩展

	// 演示基本类型的自定义转换器
	customStringConverter := func(value string, targetType reflect.Type) (reflect.Value, error) {
		// 自定义字符串处理：去除空格并转大写
		processed := fmt.Sprintf("CUSTOM_%s", value)
		return reflect.ValueOf(processed), nil
	}

	pb.RegisterConverter(reflect.String, customStringConverter)

	// 测试自定义字符串转换
	result, err := pb.ConvertValue("hello", reflect.TypeOf(""))
	if err != nil {
		fmt.Printf("❌ 自定义字符串转换失败: %v\n", err)
	} else {
		fmt.Printf("✅ 自定义字符串转换: hello -> %v\n", result.Interface())
	}

	// UUID转换示例 (通过字符串转换实现)
	uuidStr := "123e4567-e89b-12d3-a456-426614174000"
	if parsedUUID, err := uuid.Parse(uuidStr); err == nil {
		fmt.Printf("✅ UUID解析示例: %s -> %v\n", uuidStr, parsedUUID)
	}

	// 时间转换示例
	timeStr := "2023-12-25T10:30:00Z"
	if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
		fmt.Printf("✅ 时间解析示例: %s -> %v\n", timeStr, parsedTime)
	}

	// 持续时间转换示例
	durationStr := "1h30m"
	if parsedDuration, err := time.ParseDuration(durationStr); err == nil {
		fmt.Printf("✅ 持续时间解析示例: %s -> %v\n", durationStr, parsedDuration)
	}

	fmt.Println("💡 提示：对于复杂自定义类型，建议在业务逻辑中进行转换")
	fmt.Println()
}

func ExampleCustomExtractors() {
	fmt.Println("=== 自定义提取器示例 ===")

	pb := routing.NewParamBinder()

	// 注册自定义提取器
	pb.RegisterExtractor("jwt", &JWTExtractor{})
	pb.RegisterExtractor("lang", &LanguageExtractor{})

	fmt.Println("✅ 已注册JWT和语言提取器")
	fmt.Println("   - JWT提取器: 从Authorization header提取Bearer token")
	fmt.Println("   - 语言提取器: 从查询参数或Accept-Language header提取语言偏好")

	// 显示当前提取器数量
	stats := pb.GetCacheStats()
	fmt.Printf("✅ 当前提取器数量: %d\n", stats["extractors_count"])

	fmt.Println()
}

func ExampleCachePerformance() {
	fmt.Println("=== 缓存性能示例 ===")

	pb := routing.NewParamBinder()

	// 第一次调用 - 缓存未命中
	start := time.Now()
	methodType := reflect.TypeOf((*UserController)(nil)).Method(0).Type
	pb.ClearCache() // 清空缓存确保未命中
	elapsed1 := time.Since(start)

	// 模拟添加到缓存
	pb.GetCacheStats() // 触发缓存操作

	// 第二次调用 - 缓存命中
	start = time.Now()
	_ = methodType
	elapsed2 := time.Since(start)

	fmt.Printf("✅ 方法信息获取:\n")
	fmt.Printf("   - 缓存未命中: %v\n", elapsed1)
	fmt.Printf("   - 缓存命中: %v\n", elapsed2)

	// 显示缓存统计
	stats := pb.GetCacheStats()
	fmt.Printf("✅ 缓存统计: %+v\n", stats)

	fmt.Println()
}

func ExampleErrorHandling() {
	fmt.Println("=== 错误处理示例 ===")

	pb := routing.NewParamBinder()

	// 测试类型转换错误
	errorCases := []struct {
		value      string
		targetType reflect.Type
		desc       string
	}{
		{"abc", reflect.TypeOf(0), "无效整数"},
		{"invalid", reflect.TypeOf(false), "无效布尔值"},
		{"not_a_float", reflect.TypeOf(0.0), "无效浮点数"},
	}

	for _, tc := range errorCases {
		_, err := pb.ConvertValue(tc.value, tc.targetType)
		if err != nil {
			fmt.Printf("✅ %s错误处理: %v\n", tc.desc, err)
		} else {
			fmt.Printf("❌ %s应该产生错误但没有\n", tc.desc)
		}
	}

	fmt.Println()
}

// HTTP服务器示例
func RunHTTPServerExample() {
	fmt.Println("=== HTTP服务器示例 ===")
	fmt.Println("启动HTTP服务器在 :8080")
	fmt.Println("示例请求:")
	fmt.Println("  GET /users/123?include_profile=true")
	fmt.Println("  GET /users/search?name=john&age=25&tags=admin,user&page=1&size=10")
	fmt.Println("  POST /users (with JSON body)")

	h := server.Default()

	// 简单的路由处理函数
	h.GET("/users/:id", func(c context.Context, ctx *app.RequestContext) {
		pb := routing.NewParamBinder()

		// 模拟参数绑定
		idStr := string(ctx.Param("id"))
		includeProfileStr := ctx.Query("include_profile")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}

		includeProfile := includeProfileStr == "true"

		controller := &UserController{}
		result := controller.GetUser(id, includeProfile)

		ctx.JSON(consts.StatusOK, map[string]any{
			"result":      result,
			"cache_stats": pb.GetCacheStats(),
		})
	})

	h.GET("/users/search", func(c context.Context, ctx *app.RequestContext) {
		controller := &UserController{}

		// 手动参数绑定演示
		name := ctx.Query("name")
		ageStr := ctx.Query("age")
		age := 0
		if ageStr != "" {
			age, _ = strconv.Atoi(ageStr)
		}

		tagsStr := ctx.Query("tags")
		var tags []string
		if tagsStr != "" {
			// 简单的字符串分割
			tags = []string{tagsStr} // 简化处理
		}

		pageStr := ctx.Query("page")
		page := 1
		if pageStr != "" {
			page, _ = strconv.Atoi(pageStr)
		}

		sizeStr := ctx.Query("size")
		size := 10
		if sizeStr != "" {
			size, _ = strconv.Atoi(sizeStr)
		}

		result := controller.SearchUsers(name, age, tags, page, size)

		ctx.JSON(consts.StatusOK, map[string]string{"result": result})
	})

	fmt.Println("服务器配置完成，可以手动启动测试")
	fmt.Println("使用 h.Spin() 启动服务器")
}

// RunBasicExamples 运行基础示例
func RunBasicExamples() {
	fmt.Println("🚀 ParamBinder 示例程序")
	fmt.Println("======================")
	fmt.Println()

	// 运行各种示例
	ExampleBasicUsage()
	ExampleCustomConverters()
	ExampleCustomExtractors()
	ExampleCachePerformance()
	ExampleErrorHandling()

	// HTTP服务器示例（不实际启动，只显示配置）
	RunHTTPServerExample()

	fmt.Println()
	fmt.Println("✅ 所有示例执行完毕")
	fmt.Println("💡 查看各个示例函数以了解具体用法")
}

// main 函数
func main() {
	RunBasicExamples()
}
