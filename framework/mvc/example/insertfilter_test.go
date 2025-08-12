package example

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
)

// TestAppInsertFilter 测试基本的过滤器插入功能
func TestAppInsertFilter(t *testing.T) {
	app := core.NewApp()

	// 测试过滤器计数器
	var counter int

	// 创建测试过滤器
	testFilter := func(ctx *contextenhanced.Context) {
		counter++
		ctx.Set("filter_executed", true)
	}

	// 插入过滤器
	app.InsertFilter("/test/*", constant.BeforeRouter, testFilter)

	// 验证过滤器是否被注册
	filters := app.ListFilters(constant.BeforeRouter)
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}

	// 验证过滤器属性
	filter := filters[0]
	if filter.Pattern != "/test/*" {
		t.Errorf("Expected pattern '/test/*', got '%s'", filter.Pattern)
	}
	if filter.Position != constant.BeforeRouter {
		t.Errorf("Expected position %d, got %d", constant.BeforeRouter, filter.Position)
	}
	if !filter.Enabled {
		t.Error("Expected filter to be enabled")
	}
}

// TestInsertFilterWithParams 测试带参数的过滤器插入
func TestInsertFilterWithParams(t *testing.T) {
	app := core.NewApp()

	testFilter := func(ctx *contextenhanced.Context) {
		ctx.Set("test", "executed")
	}

	// 插入禁用的过滤器
	app.InsertFilter("/api/*", constant.BeforeExec, testFilter, false)

	filters := app.ListFilters(constant.BeforeExec)
	if len(filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(filters))
	}

	if filters[0].Enabled {
		t.Error("Expected filter to be disabled")
	}
}

// TestRemoveFilter 测试移除过滤器
func TestRemoveFilter(t *testing.T) {
	app := core.NewApp()

	testFilter := func(ctx *contextenhanced.Context) {}

	// 插入过滤器
	app.InsertFilter("/user/*", constant.BeforeRouter, testFilter)

	// 验证过滤器存在
	filters := app.ListFilters(constant.BeforeRouter)
	if len(filters) != 1 {
		t.Fatalf("Expected 1 filter before removal, got %d", len(filters))
	}

	// 移除过滤器
	removed := app.RemoveFilter("/user/*", constant.BeforeRouter)
	if !removed {
		t.Error("Expected filter to be removed")
	}

	// 验证过滤器已被移除
	filters = app.ListFilters(constant.BeforeRouter)
	if len(filters) != 0 {
		t.Errorf("Expected 0 filters after removal, got %d", len(filters))
	}

	// 尝试移除不存在的过滤器
	removed = app.RemoveFilter("/nonexistent/*", constant.BeforeRouter)
	if removed {
		t.Error("Expected removal of nonexistent filter to return false")
	}
}

// TestPatternMatchingViaFilters 通过实际执行过滤器测试模式匹配功能
func TestPatternMatchingViaFilters(t *testing.T) {
	app := core.NewApp()

	var matchedPaths []string
	var mu sync.Mutex

	// 创建测试过滤器
	testFilter := func(ctx *contextenhanced.Context) {
		mu.Lock()
		path := string(ctx.Request.Path())
		matchedPaths = append(matchedPaths, path)
		mu.Unlock()
	}

	// 测试API路径过滤器
	app.InsertFilter("/api/*", constant.BeforeRouter, testFilter)

	// 测试路径
	testPaths := []struct {
		path          string
		shouldMatch   bool
	}{
		{"/api/users", true},
		{"/api/users/123", true},
		{"/other/path", false},
		{"/api", false}, // /api 不匹配 /api/*
	}

	// 执行测试
	for _, test := range testPaths {
		mockCtx := createMockContext(test.path)
		app.ExecuteFilters(mockCtx, constant.BeforeRouter)
	}

	// 验证结果
	mu.Lock()
	defer mu.Unlock()

	expectedMatches := 0
	for _, test := range testPaths {
		if test.shouldMatch {
			expectedMatches++
		}
	}

	if len(matchedPaths) != expectedMatches {
		t.Errorf("Expected %d matches, got %d. Matched paths: %v", 
			expectedMatches, len(matchedPaths), matchedPaths)
	}
}

// TestMultipleFilters 测试多个过滤器的执行顺序
func TestMultipleFilters(t *testing.T) {
	app := core.NewApp()

	var execution []string
	var mu sync.Mutex

	// 创建多个过滤器
	filter1 := func(ctx *contextenhanced.Context) {
		mu.Lock()
		execution = append(execution, "filter1")
		mu.Unlock()
	}

	filter2 := func(ctx *contextenhanced.Context) {
		mu.Lock()
		execution = append(execution, "filter2")
		mu.Unlock()
	}

	filter3 := func(ctx *contextenhanced.Context) {
		mu.Lock()
		execution = append(execution, "filter3")
		mu.Unlock()
	}

	// 按顺序插入过滤器
	app.InsertFilter("/test/*", constant.BeforeRouter, filter1)
	app.InsertFilter("/test/*", constant.BeforeRouter, filter2)
	app.InsertFilter("/test/*", constant.BeforeRouter, filter3)

	// 创建模拟上下文并执行过滤器
	mockCtx := createMockContext("/test/path")
	app.ExecuteFilters(mockCtx, constant.BeforeRouter)

	// 验证执行顺序
	mu.Lock()
	defer mu.Unlock()

	if len(execution) != 3 {
		t.Fatalf("Expected 3 executions, got %d", len(execution))
	}

	// 验证执行顺序是插入顺序
	expected := []string{"filter1", "filter2", "filter3"}
	for i, expected := range expected {
		if execution[i] != expected {
			t.Errorf("Expected execution[%d] = '%s', got '%s'", i, expected, execution[i])
		}
	}
}

// TestFilterAbort 测试过滤器中止请求处理
func TestFilterAbort(t *testing.T) {
	app := core.NewApp()

	var executed []string
	var mu sync.Mutex

	// 第一个过滤器中止请求
	abortFilter := func(ctx *contextenhanced.Context) {
		mu.Lock()
		executed = append(executed, "abort")
		mu.Unlock()
		ctx.Abort()
	}

	// 第二个过滤器不应该被执行
	normalFilter := func(ctx *contextenhanced.Context) {
		mu.Lock()
		executed = append(executed, "normal")
		mu.Unlock()
	}

	app.InsertFilter("/test/*", constant.BeforeRouter, abortFilter)
	app.InsertFilter("/test/*", constant.BeforeRouter, normalFilter)

	// 执行过滤器
	mockCtx := createMockContext("/test/abort")
	app.ExecuteFilters(mockCtx, constant.BeforeRouter)

	// 验证只有第一个过滤器被执行
	mu.Lock()
	defer mu.Unlock()

	if len(executed) != 1 {
		t.Errorf("Expected 1 execution, got %d", len(executed))
	}
	if executed[0] != "abort" {
		t.Errorf("Expected 'abort', got '%s'", executed[0])
	}
}

// TestInvalidPosition 测试无效位置处理
func TestInvalidPosition(t *testing.T) {
	app := core.NewApp()

	testFilter := func(ctx *contextenhanced.Context) {}

	// 测试无效位置
	app.InsertFilter("/test/*", -1, testFilter)  // 无效的负数位置
	app.InsertFilter("/test/*", 999, testFilter) // 无效的大数位置

	// 验证没有过滤器被插入
	for position := constant.BeforeStatic; position <= constant.FinishRouter; position++ {
		filters := app.ListFilters(position)
		if len(filters) > 0 {
			t.Errorf("Expected no filters for position %d, got %d", position, len(filters))
		}
	}
}

// TestGetAllFilters 测试获取所有过滤器
func TestGetAllFilters(t *testing.T) {
	app := core.NewApp()

	testFilter := func(ctx *contextenhanced.Context) {}

	// 在不同位置插入过滤器
	app.InsertFilter("/api/*", constant.BeforeRouter, testFilter)
	app.InsertFilter("/user/*", constant.BeforeExec, testFilter)
	app.InsertFilter("/admin/*", constant.AfterExec, testFilter)

	allFilters := app.GetAllFilters()

	// 验证过滤器数量
	if len(allFilters[constant.BeforeRouter]) != 1 {
		t.Errorf("Expected 1 constant.BeforeRouter filter, got %d", len(allFilters[constant.BeforeRouter]))
	}
	if len(allFilters[constant.BeforeExec]) != 1 {
		t.Errorf("Expected 1 constant.BeforeExec filter, got %d", len(allFilters[constant.BeforeExec]))
	}
	if len(allFilters[constant.AfterExec]) != 1 {
		t.Errorf("Expected 1 constant.AfterExec filter, got %d", len(allFilters[constant.AfterExec]))
	}
}

// TestFilterThreadSafety 测试过滤器的线程安全性
func TestFilterThreadSafety(t *testing.T) {
	app := core.NewApp()

	var wg sync.WaitGroup

	// 创建测试过滤器
	testFilter := func(ctx *contextenhanced.Context) {
		// 模拟一些工作
		time.Sleep(time.Millisecond)
	}

	// 并发插入过滤器
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pattern := "/test/" + strings.Repeat("x", id%5)
			app.InsertFilter(pattern, constant.BeforeRouter, testFilter)
		}(i)
	}

	// 并发移除过滤器
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pattern := "/test/" + strings.Repeat("x", id%3)
			app.RemoveFilter(pattern, constant.BeforeRouter)
		}(i)
	}

	// 并发读取过滤器
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = app.ListFilters(constant.BeforeRouter)
			_ = app.GetAllFilters()
		}()
	}

	wg.Wait()

	// 验证没有panic或竞态条件
	t.Log("Thread safety test completed successfully")
}

// createMockContext 创建模拟上下文用于测试
func createMockContext(path string) *contextenhanced.Context {
	// 使用hertz的测试工具创建真实的上下文
	ctx := app.NewContext(0)
	
	// 设置请求路径
	ctx.Request.SetRequestURI(path)
	ctx.Request.SetMethod("GET")
	
	// 包装成增强上下文
	enhancedCtx := &contextenhanced.Context{
		Request:        ctx,
		RequestContext: ctx,
		Keys:          make(map[string]interface{}),
	}
	
	return enhancedCtx
}