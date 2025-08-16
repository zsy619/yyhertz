package errors

import (
	"errors"
	"testing"
	"time"
)

// TestConfigManager 测试配置管理器
func TestConfigManager(t *testing.T) {
	manager := NewDefaultConfigManager()
	
	// 测试获取状态码配置
	config := manager.GetStatusConfig(404)
	if config == nil {
		t.Error("404状态码配置不应为空")
	}
	
	if config.StatusCode != 404 {
		t.Errorf("期望状态码404，实际%d", config.StatusCode)
	}
	
	t.Logf("✅ 配置管理器工作正常，404配置: %s", config.Message)
}

// TestErrorContext 测试错误上下文
func TestErrorContext(t *testing.T) {
	ctx := &ErrorContext{
		StatusCode:    500,
		StatusText:    "Internal Server Error",
		ErrorMessage:  "服务器内部错误",
		RequestPath:   "/test",
		RequestMethod: "GET",
		Timestamp:     time.Now(),
		Details:       make(map[string]any),
		Suggestions:   []string{"重试请求", "联系技术支持"},
	}
	
	if ctx.StatusCode != 500 {
		t.Errorf("期望状态码500，实际%d", ctx.StatusCode)
	}
	
	if len(ctx.Suggestions) != 2 {
		t.Errorf("期望2个建议，实际%d", len(ctx.Suggestions))
	}
	
	t.Log("✅ 错误上下文创建成功")
}

// TestTemplateManagerBasic 测试模板管理器基础功能
func TestTemplateManagerBasic(t *testing.T) {
	manager := NewDefaultTemplateManager()
	
	// 测试加载简单模板
	err := manager.LoadTemplate("test", "Error {{.StatusCode}}: {{.Message}}")
	if err != nil {
		t.Errorf("加载模板失败: %v", err)
	}
	
	// 测试渲染模板
	data := map[string]interface{}{
		"StatusCode": 404,
		"Message":    "页面未找到",
	}
	
	result, err := manager.RenderTemplate("test", data)
	if err != nil {
		t.Errorf("渲染模板失败: %v", err)
	}
	
	expected := "Error 404: 页面未找到"
	if result != expected {
		t.Errorf("期望结果'%s'，实际'%s'", expected, result)
	}
	
	t.Log("✅ 模板管理器基础功能正常")
}

// TestStatisticsManagerBasic 测试统计管理器基础功能
func TestStatisticsManagerBasic(t *testing.T) {
	manager := NewDefaultStatisticsManager()
	
	// 记录错误
	manager.RecordError(404, "/api/user", "GET", errors.New("用户未找到"))
	manager.RecordError(500, "/api/order", "POST", errors.New("服务器错误"))
	
	// 等待一下让异步处理完成
	time.Sleep(100 * time.Millisecond)
	
	// 获取统计信息
	stats := manager.GetStatistics()
	if stats.TotalErrors < 2 {
		t.Errorf("期望至少2个错误，实际%d", stats.TotalErrors)
	}
	
	// 检查错误分类统计
	if stats.ErrorsByStatus[404] < 1 {
		t.Error("404错误计数不正确")
	}
	
	if stats.ErrorsByStatus[500] < 1 {
		t.Error("500错误计数不正确")
	}
	
	t.Logf("✅ 统计管理器基础功能正常，总错误数: %d", stats.TotalErrors)
}

// TestFactoryBasic 测试工厂基础功能
func TestFactoryBasic(t *testing.T) {
	factory := NewDefaultErrorControllerFactory()
	if factory == nil {
		t.Error("工厂创建失败")
	}
	
	// 测试组件工厂
	componentFactory := NewComponentFactory()
	if componentFactory == nil {
		t.Error("组件工厂创建失败")
	}
	
	// 测试创建配置管理器
	configManager := componentFactory.CreateConfigManager("default")
	if configManager == nil {
		t.Error("配置管理器创建失败")
	}
	
	// 测试创建统计管理器
	statsManager := componentFactory.CreateStatisticsManager("default")
	if statsManager == nil {
		t.Error("统计管理器创建失败")
	}
	
	// 测试创建模板管理器
	templateManager := componentFactory.CreateTemplateManager("default")
	if templateManager == nil {
		t.Error("模板管理器创建失败")
	}
	
	t.Log("✅ 工厂基础功能正常")
}

// TestBuilderPatternBasic 测试构建器模式基础功能
func TestBuilderPatternBasic(t *testing.T) {
	builder := NewErrorControllerBuilder()
	if builder == nil {
		t.Error("构建器创建失败")
	}
	
	// 测试设置配置
	config := &ErrorControllerConfig{
		ShowDetailedError: true,
		Language:         "zh-CN",
		CustomTitle:      "测试框架",
		EnableDebugInfo:  true,
	}
	
	builder = builder.WithConfig(config)
	if builder.config != config {
		t.Error("构建器配置设置失败")
	}
	
	t.Log("✅ 构建器模式基础功能正常")
}

// TestErrorStatistics 测试错误统计数据结构
func TestErrorStatistics(t *testing.T) {
	stats := NewErrorStatistics()
	if stats == nil {
		t.Error("统计结构创建失败")
	}
	
	if stats.TotalErrors != 0 {
		t.Errorf("期望初始错误数为0，实际%d", stats.TotalErrors)
	}
	
	if stats.ErrorsByStatus == nil {
		t.Error("状态码统计map未初始化")
	}
	
	if stats.ErrorsByPath == nil {
		t.Error("路径统计map未初始化")
	}
	
	t.Log("✅ 错误统计数据结构正常")
}

// TestStatusConfig 测试状态码配置
func TestStatusConfig(t *testing.T) {
	// 测试获取常见状态码配置
	testCodes := []int{400, 401, 403, 404, 429, 500, 502, 503}
	
	for _, code := range testCodes {
		config := GetStatusConfig(code)
		if config == nil {
			t.Errorf("状态码%d的配置不应为空", code)
			continue
		}
		
		if config.StatusCode != code {
			t.Errorf("状态码%d配置错误，期望%d，实际%d", code, code, config.StatusCode)
		}
		
		if config.Message == "" {
			t.Errorf("状态码%d的消息不应为空", code)
		}
		
		t.Logf("✅ 状态码%d配置正常: %s", code, config.Message)
	}
}

// TestHighPerformanceStatistics 测试高性能统计管理器
func TestHighPerformanceStatistics(t *testing.T) {
	manager := NewHighPerformanceStatisticsManager()
	if manager == nil {
		t.Error("高性能统计管理器创建失败")
	}
	
	// 记录一些错误
	for i := 0; i < 100; i++ {
		manager.RecordError(404, "/api/test", "GET", errors.New("测试错误"))
	}
	
	stats := manager.GetStatistics()
	if stats.TotalErrors < 100 {
		t.Errorf("期望至少100个错误，实际%d", stats.TotalErrors)
	}
	
	t.Logf("✅ 高性能统计管理器正常，总错误数: %d", stats.TotalErrors)
}

// BenchmarkConfigManager 配置管理器性能测试
func BenchmarkConfigManager(b *testing.B) {
	manager := NewDefaultConfigManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetStatusConfig(404)
	}
}

// BenchmarkTemplateRender 模板渲染性能测试
func BenchmarkTemplateRender(b *testing.B) {
	manager := NewDefaultTemplateManager()
	data := map[string]interface{}{
		"StatusCode": 404,
		"Message":    "页面未找到",
	}
	
	// 预热
	_, _ = manager.RenderTemplate("error.minimal", data)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.RenderTemplate("error.minimal", data)
	}
}

// BenchmarkStatisticsRecord 统计记录性能测试
func BenchmarkStatisticsRecord(b *testing.B) {
	manager := NewDefaultStatisticsManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RecordError(404, "/api/test", "GET", errors.New("测试错误"))
	}
}