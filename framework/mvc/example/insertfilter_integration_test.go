package example

import (
	"sync"
	"testing"

	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// TestGlobalInsertFilter 测试全局静态方法的过滤器插入
func TestGlobalInsertFilter(t *testing.T) {
	// 初始化测试应用
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		if mvc.HertzApp != nil {
			mvc.HertzApp = nil
		}
	}()

	var executionCount int
	var mu sync.Mutex

	// 创建测试过滤器
	testFilter := func(ctx *context.Context) {
		mu.Lock()
		executionCount++
		mu.Unlock()
		ctx.Set("global_filter", true)
	}

	// 使用全局静态方法插入过滤器
	mvc.InsertFilter("/api/*", constant.BeforeRouter, testFilter)

	// 验证过滤器被插入
	filters := mvc.ListFilters(constant.BeforeRouter)
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}

	// 验证过滤器属性
	filter := filters[0]
	if filter.Pattern != "/api/*" {
		t.Errorf("Expected pattern '/api/*', got '%s'", filter.Pattern)
	}
	if filter.Position != constant.BeforeRouter {
		t.Errorf("Expected position %d, got %d", constant.BeforeRouter, filter.Position)
	}
}

// TestGlobalRemoveFilter 测试全局移除过滤器
func TestGlobalRemoveFilter(t *testing.T) {
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		mvc.HertzApp = nil
	}()

	testFilter := func(ctx *context.Context) {}

	// 插入过滤器
	mvc.InsertFilter("/user/*", constant.BeforeExec, testFilter)

	// 验证过滤器存在
	filters := mvc.ListFilters(constant.BeforeExec)
	if len(filters) != 1 {
		t.Fatalf("Expected 1 filter before removal, got %d", len(filters))
	}

	// 移除过滤器
	removed := mvc.RemoveFilter("/user/*", constant.BeforeExec)
	if !removed {
		t.Error("Expected filter to be removed")
	}

	// 验证过滤器已被移除
	filters = mvc.ListFilters(constant.BeforeExec)
	if len(filters) != 0 {
		t.Errorf("Expected 0 filters after removal, got %d", len(filters))
	}
}

// TestGlobalFilterPositions 测试所有5个位置的过滤器
func TestGlobalFilterPositions(t *testing.T) {
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		mvc.HertzApp = nil
	}()

	var executions []string
	var mu sync.Mutex

	// 为每个位置创建过滤器
	positions := []int{constant.BeforeStatic, constant.BeforeRouter, constant.BeforeExec, constant.AfterExec, constant.FinishRouter}
	positionNames := []string{"constant.BeforeStatic", "constant.BeforeRouter", "constant.BeforeExec", "constant.AfterExec", "constant.FinishRouter"}

	for i, position := range positions {
		name := positionNames[i]
		filter := func(posName string) mvc.FilterFunc {
			return func(ctx *context.Context) {
				mu.Lock()
				executions = append(executions, posName)
				mu.Unlock()
			}
		}(name)

		mvc.InsertFilter("/*", position, filter)
	}

	// 验证所有过滤器都被插入
	for i, position := range positions {
		filters := mvc.ListFilters(position)
		if len(filters) != 1 {
			t.Errorf("Position %s: expected 1 filter, got %d", positionNames[i], len(filters))
		}
	}

	// 验证mvc.GetAllFilters
	allFilters := mvc.GetAllFilters()
	for i, position := range positions {
		if len(allFilters[position]) != 1 {
			t.Errorf("mvc.GetAllFilters for position %s: expected 1 filter, got %d",
				positionNames[i], len(allFilters[position]))
		}
	}
}

// TestGlobalFilterWithParams 测试带参数的全局过滤器插入
func TestGlobalFilterWithParams(t *testing.T) {
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		mvc.HertzApp = nil
	}()

	testFilter := func(ctx *context.Context) {}

	// 插入禁用的过滤器
	mvc.InsertFilter("/admin/*", constant.BeforeRouter, testFilter, false)

	filters := mvc.ListFilters(constant.BeforeRouter)
	if len(filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(filters))
	}

	if filters[0].Enabled {
		t.Error("Expected filter to be disabled")
	}

	// 插入启用的过滤器（默认行为）
	mvc.InsertFilter("/public/*", constant.BeforeRouter, testFilter)

	filters = mvc.ListFilters(constant.BeforeRouter)
	if len(filters) != 2 {
		t.Fatalf("Expected 2 filters, got %d", len(filters))
	}

	// 找到公共过滤器并验证它是启用的
	var publicFilter *mvc.FilterPattern
	for _, filter := range filters {
		if filter.Pattern == "/public/*" {
			publicFilter = filter
			break
		}
	}

	if publicFilter == nil {
		t.Fatal("Could not find public filter")
	}

	if !publicFilter.Enabled {
		t.Error("Expected public filter to be enabled by default")
	}
}

// TestGlobalFilterNilApp 测试当HertzApp为nil时的行为
func TestGlobalFilterNilApp(t *testing.T) {
	// 确保HertzApp为nil
	oldApp := mvc.HertzApp
	mvc.HertzApp = nil
	defer func() {
		mvc.HertzApp = oldApp
	}()

	testFilter := func(ctx *context.Context) {}

	// 这些操作不应该panic
	mvc.InsertFilter("/test/*", constant.BeforeRouter, testFilter)
	removed := mvc.RemoveFilter("/test/*", constant.BeforeRouter)
	filters := mvc.ListFilters(constant.BeforeRouter)
	allFilters := mvc.GetAllFilters()

	// 验证返回值
	if removed {
		t.Error("Expected mvc.RemoveFilter to return false when HertzApp is nil")
	}
	if len(filters) != 0 {
		t.Errorf("Expected empty filter list when HertzApp is nil, got %d", len(filters))
	}
	if len(allFilters) != 0 {
		t.Errorf("Expected empty all filters map when HertzApp is nil, got %d", len(allFilters))
	}
}

// TestFilterPatternMatching 测试复杂的模式匹配场景
func TestFilterPatternMatching(t *testing.T) {
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		mvc.HertzApp = nil
	}()

	var matchedPatterns []string
	var mu sync.Mutex

	// 创建多个具有不同模式的过滤器
	patterns := []string{
		"/api/*",
		"/api/v1/*",
		"*/json",
		"/exact",
		"*",
	}

	for _, pattern := range patterns {
		p := pattern // 捕获循环变量
		filter := func(ctx *context.Context) {
			mu.Lock()
			matchedPatterns = append(matchedPatterns, p)
			mu.Unlock()
		}
		mvc.InsertFilter(p, constant.BeforeRouter, filter)
	}

	// 验证所有过滤器都被插入
	filters := mvc.ListFilters(constant.BeforeRouter)
	if len(filters) != len(patterns) {
		t.Errorf("Expected %d filters, got %d", len(patterns), len(filters))
	}

	// 验证模式存储正确
	for i, filter := range filters {
		if filter.Pattern != patterns[i] {
			t.Errorf("Filter %d: expected pattern '%s', got '%s'",
				i, patterns[i], filter.Pattern)
		}
	}
}

// TestConcurrentGlobalFilterOperations 测试并发的全局过滤器操作
func TestConcurrentGlobalFilterOperations(t *testing.T) {
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		mvc.HertzApp = nil
	}()

	var wg sync.WaitGroup
	testFilter := func(ctx *context.Context) {}

	// 并发插入过滤器
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pattern := "/test/" + string(rune('a'+id%26))
			mvc.InsertFilter(pattern, constant.BeforeRouter, testFilter)
		}(i)
	}

	// 并发移除过滤器
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pattern := "/test/" + string(rune('a'+id%26))
			mvc.RemoveFilter(pattern, constant.BeforeRouter)
		}(i)
	}

	// 并发读取过滤器
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mvc.ListFilters(constant.BeforeRouter)
			_ = mvc.GetAllFilters()
		}()
	}

	wg.Wait()

	// 验证操作完成且没有panic
	t.Log("Concurrent operations completed successfully")
}

// BenchmarkGlobalInsertFilter 性能测试：过滤器插入
func BenchmarkGlobalInsertFilter(b *testing.B) {
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		mvc.HertzApp = nil
	}()

	testFilter := func(ctx *context.Context) {}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := "/benchmark/*"
		mvc.InsertFilter(pattern, constant.BeforeRouter, testFilter)
		mvc.RemoveFilter(pattern, constant.BeforeRouter) // 清理
	}
}

// BenchmarkGlobalListFilters 性能测试：列出过滤器
func BenchmarkGlobalListFilters(b *testing.B) {
	mvc.HertzApp = mvc.NewApp()
	defer func() {
		mvc.HertzApp = nil
	}()

	testFilter := func(ctx *context.Context) {}

	// 预先插入一些过滤器
	for i := 0; i < 100; i++ {
		pattern := "/bench/" + string(rune('a'+i%26))
		mvc.InsertFilter(pattern, constant.BeforeRouter, testFilter)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mvc.ListFilters(constant.BeforeRouter)
	}
}