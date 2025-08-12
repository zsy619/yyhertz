package example

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
)

// TestPatternAutoMatching 测试框架自动进行pattern匹配
func TestPatternAutoMatching(t *testing.T) {
	app := core.NewApp()

	// 测试计数器
	var apiExecuted, userExecuted, globalExecuted int

	// API过滤器 - 只应在 /api/* 路径执行
	apiFilter := func(ctx *contextenhanced.Context) {
		apiExecuted++
		ctx.Set("api_executed", true)
	}

	// 用户过滤器 - 只应在 /user/* 路径执行
	userFilter := func(ctx *contextenhanced.Context) {
		userExecuted++
		ctx.Set("user_executed", true)
	}

	// 全局过滤器 - 应在所有路径执行
	globalFilter := func(ctx *contextenhanced.Context) {
		globalExecuted++
		ctx.Set("global_executed", true)
	}

	// 插入过滤器
	app.InsertFilter("/api/*", constant.BeforeRouter, apiFilter)
	app.InsertFilter("/user/*", constant.BeforeRouter, userFilter)
	app.InsertFilter("*", constant.BeforeRouter, globalFilter)

	// 测试用例
	testCases := []struct {
		path                string
		expectedApi         int
		expectedUser        int
		expectedGlobal      int
	}{
		{"/api/test", 1, 0, 1},      // API + 全局
		{"/user/profile", 1, 1, 2},  // 用户 + 全局
		{"/other/path", 1, 1, 3},    // 只有全局
		{"/api/users", 2, 1, 4},     // API + 全局
	}

	for i, tc := range testCases {
		// 创建模拟上下文
		mockCtx := createMockContextWithPath(tc.path)
		
		// 执行过滤器
		app.ExecuteFilters(mockCtx, constant.BeforeRouter)

		// 验证执行次数
		if apiExecuted != tc.expectedApi {
			t.Errorf("Test case %d (%s): API filter executed %d times, expected %d",
				i, tc.path, apiExecuted, tc.expectedApi)
		}
		if userExecuted != tc.expectedUser {
			t.Errorf("Test case %d (%s): User filter executed %d times, expected %d",
				i, tc.path, userExecuted, tc.expectedUser)
		}
		if globalExecuted != tc.expectedGlobal {
			t.Errorf("Test case %d (%s): Global filter executed %d times, expected %d",
				i, tc.path, globalExecuted, tc.expectedGlobal)
		}
	}
}

// TestNoManualPathCheckNeeded 测试过滤器不需要手动检查路径
func TestNoManualPathCheckNeeded(t *testing.T) {
	app := core.NewApp()

	var executedPaths []string

	// 过滤器记录执行的路径，不做任何路径判断
	pathRecorderFilter := func(ctx *contextenhanced.Context) {
		// 注意：这里故意不检查路径，依赖框架的自动匹配
		path := string(ctx.Request.Path())
		executedPaths = append(executedPaths, path)
	}

	// 只对 /secure/* 路径注册过滤器
	app.InsertFilter("/secure/*", constant.BeforeRouter, pathRecorderFilter)

	// 测试不同路径
	testPaths := []string{
		"/secure/admin",     // 应该执行
		"/secure/user",      // 应该执行
		"/public/info",      // 不应该执行
		"/secure/api/data",  // 应该执行
		"/other/path",       // 不应该执行
	}

	expectedPaths := []string{
		"/secure/admin",
		"/secure/user", 
		"/secure/api/data",
	}

	for _, path := range testPaths {
		mockCtx := createMockContextWithPath(path)
		app.ExecuteFilters(mockCtx, constant.BeforeRouter)
	}

	// 验证只有匹配的路径被执行
	if len(executedPaths) != len(expectedPaths) {
		t.Errorf("Expected %d executions, got %d", len(expectedPaths), len(executedPaths))
	}

	for i, expectedPath := range expectedPaths {
		if i >= len(executedPaths) || executedPaths[i] != expectedPath {
			t.Errorf("Expected path %s at position %d, got %s", 
				expectedPath, i, executedPaths[i])
		}
	}
}

// TestComplexPatternMatching 测试复杂模式匹配
func TestComplexPatternMatching(t *testing.T) {
	app := core.NewApp()

	var executions []string

	// 不同pattern的过滤器
	filters := map[string]func(*contextenhanced.Context){
		"exact": func(ctx *contextenhanced.Context) {
			executions = append(executions, "exact")
		},
		"prefix": func(ctx *contextenhanced.Context) {
			executions = append(executions, "prefix")
		},
		"suffix": func(ctx *contextenhanced.Context) {
			executions = append(executions, "suffix")
		},
		"middle": func(ctx *contextenhanced.Context) {
			executions = append(executions, "middle")
		},
	}

	// 注册不同pattern的过滤器
	app.InsertFilter("/exact/path", constant.BeforeRouter, filters["exact"])
	app.InsertFilter("/api/*", constant.BeforeRouter, filters["prefix"])
	app.InsertFilter("*.json", constant.BeforeRouter, filters["suffix"])
	app.InsertFilter("/api/*/users", constant.BeforeRouter, filters["middle"])

	// 测试用例
	testCases := []struct {
		path     string
		expected []string
	}{
		{"/exact/path", []string{"exact"}},
		{"/api/v1", []string{"prefix"}},
		{"/data.json", []string{"suffix"}},
		{"/api/v1.json", []string{"prefix", "suffix"}},
		{"/api/v1/users", []string{"prefix", "middle"}},
		{"/api/v2/users.json", []string{"prefix", "suffix"}}, // /api/*/users不匹配.json结尾的路径
		{"/api/v2/users", []string{"prefix", "middle"}}, // 这个会匹配middle模式
		{"/other/path", []string{}},
	}

	for _, tc := range testCases {
		// 重置执行记录
		executions = []string{}

		// 执行过滤器
		mockCtx := createMockContextWithPath(tc.path)
		app.ExecuteFilters(mockCtx, constant.BeforeRouter)

		// 验证执行结果
		if len(executions) != len(tc.expected) {
			t.Errorf("Path %s: expected %d executions, got %d (%v)", 
				tc.path, len(tc.expected), len(executions), executions)
			continue
		}

		for _, expected := range tc.expected {
			found := false
			for _, actual := range executions {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Path %s: expected execution '%s' not found in %v", 
					tc.path, expected, executions)
			}
		}
	}
}

// createMockContextWithPath 创建带指定路径的模拟上下文
func createMockContextWithPath(path string) *contextenhanced.Context {
	// 创建模拟的RequestContext
	reqCtx := &app.RequestContext{}
	
	// 设置请求路径
	reqCtx.Request.SetRequestURI(path)
	reqCtx.Request.Header.SetMethod("GET")
	
	// 创建增强上下文
	enhancedCtx := contextenhanced.NewContext(reqCtx)
	
	return enhancedCtx
}