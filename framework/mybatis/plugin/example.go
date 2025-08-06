// Package plugin 插件使用示例
//
// 展示如何使用MyBatis插件系统
package plugin

import (
	"fmt"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis/config"
)

// PluginExample 插件使用示例
type PluginExample struct {
	manager *PluginManager
}

// NewPluginExample 创建插件示例
func NewPluginExample() *PluginExample {
	configuration := config.NewConfiguration()
	manager := NewPluginManager(configuration)

	return &PluginExample{
		manager: manager,
	}
}

// DemoBasicUsage 演示基本用法
func (example *PluginExample) DemoBasicUsage() {
	fmt.Println("=== MyBatis 插件系统基本用法演示 ===")

	// 1. 查看所有注册的插件
	fmt.Println("\n1. 已注册的插件:")
	plugins := example.manager.GetAllPlugins()
	for name, plugin := range plugins {
		fmt.Printf("  - %s (执行顺序: %d)\n", name, plugin.GetOrder())
	}

	// 2. 查看启用的插件
	fmt.Println("\n2. 启用的插件:")
	enabledPlugins := example.manager.GetEnabledPlugins()
	for _, plugin := range enabledPlugins {
		fmt.Printf("  - %s\n", plugin.GetName())
	}

	// 3. 配置插件
	fmt.Println("\n3. 配置插件:")
	example.configurePlugins()

	// 4. 演示插件功能
	fmt.Println("\n4. 插件功能演示:")
	example.demonstratePluginFeatures()
}

// configurePlugins 配置插件
func (example *PluginExample) configurePlugins() {
	// 配置分页插件
	paginationConfig := map[string]any{
		"dialectType":     "mysql",
		"defaultPageSize": 10,
		"maxPageSize":     100,
	}
	example.manager.ConfigurePlugin("pagination", paginationConfig)
	fmt.Println("  ✓ 分页插件配置完成")

	// 配置性能监控插件
	performanceConfig := map[string]any{
		"slowQueryThreshold": 500, // 500毫秒
		"enableMetrics":      true,
	}
	example.manager.ConfigurePlugin("performance", performanceConfig)
	fmt.Println("  ✓ 性能监控插件配置完成")

	// 配置SQL日志插件
	sqlLogConfig := map[string]any{
		"logLevel":     "INFO",
		"logSql":       true,
		"logParameter": true,
		"logResult":    false,
	}
	example.manager.ConfigurePlugin("sqllog", sqlLogConfig)
	fmt.Println("  ✓ SQL日志插件配置完成")

	// 配置缓存增强插件
	cacheConfig := map[string]any{
		"enableStatistics": true,
		"enablePreload":    false,
	}
	example.manager.ConfigurePlugin("cache_enhancer", cacheConfig)
	fmt.Println("  ✓ 缓存增强插件配置完成")

	// 配置结果转换插件
	transformConfig := map[string]any{
		"enableTransform": true,
	}
	example.manager.ConfigurePlugin("result_transformer", transformConfig)
	fmt.Println("  ✓ 结果转换插件配置完成")
}

// demonstratePluginFeatures 演示插件功能
func (example *PluginExample) demonstratePluginFeatures() {
	// 演示分页功能
	example.demonstratePagination()

	// 演示性能监控
	example.demonstratePerformanceMonitoring()

	// 演示SQL日志
	example.demonstrateSqlLogging()

	// 演示缓存功能
	example.demonstrateCaching()

	// 演示结果转换
	example.demonstrateResultTransformation()
}

// demonstratePagination 演示分页功能
func (example *PluginExample) demonstratePagination() {
	fmt.Println("\n  📄 分页插件演示:")

	// 创建分页请求
	pageRequest := &PageRequest{
		PageNum:  1,
		PageSize: 10,
	}

	fmt.Printf("    分页请求: 第%d页，每页%d条\n", pageRequest.PageNum, pageRequest.PageSize)

	// 模拟分页结果
	pageResult := &PageResult{
		List:       []any{"用户1", "用户2", "用户3"},
		Total:      100,
		PageNum:    1,
		PageSize:   10,
		TotalPages: 10,
		HasNext:    true,
		HasPrev:    false,
	}

	fmt.Printf("    分页结果: 总计%d条，当前第%d页，共%d页\n",
		pageResult.Total, pageResult.PageNum, pageResult.TotalPages)
}

// demonstratePerformanceMonitoring 演示性能监控
func (example *PluginExample) demonstratePerformanceMonitoring() {
	fmt.Println("\n  📊 性能监控插件演示:")

	plugin, err := example.manager.GetPlugin("performance")
	if err != nil {
		fmt.Printf("    获取性能插件失败: %v\n", err)
		return
	}

	performancePlugin := plugin.(*PerformancePlugin)

	// 模拟一些性能数据
	performancePlugin.metrics.TotalQueries = 1000
	performancePlugin.metrics.SlowQueries = 5
	performancePlugin.metrics.FailedQueries = 2
	performancePlugin.metrics.MaxTime = 2 * time.Second
	performancePlugin.metrics.MinTime = 10 * time.Millisecond
	performancePlugin.metrics.AvgTime = 150 * time.Millisecond

	report := performancePlugin.GetPerformanceReport()
	fmt.Printf("    总查询数: %v\n", report["总查询数"])
	fmt.Printf("    慢查询数: %v\n", report["慢查询数"])
	fmt.Printf("    平均执行时间: %v\n", report["平均执行时间"])
}

// demonstrateSqlLogging 演示SQL日志
func (example *PluginExample) demonstrateSqlLogging() {
	fmt.Println("\n  📝 SQL日志插件演示:")

	// 模拟SQL日志条目
	logEntry := &SqlLogEntry{
		Timestamp:     time.Now(),
		Method:        "selectUserById",
		SQL:           "SELECT * FROM users WHERE id = ?",
		Parameters:    []any{123},
		ExecutionTime: 50 * time.Millisecond,
		Success:       true,
		Error:         nil,
		RowsAffected:  1,
	}

	fmt.Printf("    执行方法: %s\n", logEntry.Method)
	fmt.Printf("    SQL语句: %s\n", logEntry.SQL)
	fmt.Printf("    参数: %v\n", logEntry.Parameters)
	fmt.Printf("    执行时间: %v\n", logEntry.ExecutionTime)
	fmt.Printf("    执行状态: %s\n", map[bool]string{true: "成功", false: "失败"}[logEntry.Success])
}

// demonstrateCaching 演示缓存功能
func (example *PluginExample) demonstrateCaching() {
	fmt.Println("\n  🗄️ 缓存增强插件演示:")

	plugin, err := example.manager.GetPlugin("cache_enhancer")
	if err != nil {
		fmt.Printf("    获取缓存插件失败: %v\n", err)
		return
	}

	cachePlugin := plugin.(*CacheEnhancerPlugin)

	// 模拟缓存统计
	cachePlugin.statistics.TotalHits = 800
	cachePlugin.statistics.TotalMisses = 200
	cachePlugin.statistics.HitRate = 0.8
	cachePlugin.statistics.MissRate = 0.2

	report := cachePlugin.GetCacheReport()
	fmt.Printf("    总命中数: %v\n", report["总命中数"])
	fmt.Printf("    总未命中数: %v\n", report["总未命中数"])
	fmt.Printf("    命中率: %v\n", report["命中率"])
}

// demonstrateResultTransformation 演示结果转换
func (example *PluginExample) demonstrateResultTransformation() {
	fmt.Println("\n  🔄 结果转换插件演示:")

	// 模拟结构体数据
	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		CreateAt time.Time `json:"create_at"`
	}

	user := User{
		ID:       123,
		Name:     "张三",
		Email:    "zhangsan@example.com",
		CreateAt: time.Now(),
	}

	fmt.Printf("    原始数据: %+v\n", user)

	// 使用Map转换器
	mapTransformer := &MapTransformer{}
	userMap, err := mapTransformer.structToMap(user)
	if err == nil {
		fmt.Printf("    转换为Map: %v\n", userMap)
	}

	// 使用JSON转换器
	jsonTransformer := &JsonTransformer{}
	rule := TransformRule{ToType: "string"}
	jsonStr, err := jsonTransformer.Transform(user, rule)
	if err == nil {
		fmt.Printf("    转换为JSON: %s\n", jsonStr)
	}
}

// DemoAdvancedUsage 演示高级用法
func (example *PluginExample) DemoAdvancedUsage() {
	fmt.Println("\n=== MyBatis 插件系统高级用法演示 ===")

	// 1. 自定义插件
	fmt.Println("\n1. 自定义插件:")
	example.demonstrateCustomPlugin()

	// 2. 插件链执行
	fmt.Println("\n2. 插件链执行:")
	example.demonstratePluginChain()

	// 3. 动态插件管理
	fmt.Println("\n3. 动态插件管理:")
	example.demonstrateDynamicPluginManagement()

	// 4. 插件配置加载
	fmt.Println("\n4. 插件配置加载:")
	example.demonstrateConfigurationLoading()
}

// demonstrateCustomPlugin 演示自定义插件
func (example *PluginExample) demonstrateCustomPlugin() {
	// 创建自定义插件
	customPlugin := &CustomAuditPlugin{
		BasePlugin: NewBasePlugin("audit", 10),
		auditLog:   make([]AuditRecord, 0),
	}

	// 注册自定义插件
	err := example.manager.RegisterPlugin(customPlugin)
	if err != nil {
		fmt.Printf("    注册自定义插件失败: %v\n", err)
		return
	}

	fmt.Println("    ✓ 自定义审计插件注册成功")
}

// demonstratePluginChain 演示插件链执行
func (example *PluginExample) demonstratePluginChain() {
	fmt.Println("    插件执行顺序:")
	enabledPlugins := example.manager.GetEnabledPlugins()
	for i, plugin := range enabledPlugins {
		fmt.Printf("      %d. %s (顺序: %d)\n", i+1, plugin.GetName(), plugin.GetOrder())
	}
}

// demonstrateDynamicPluginManagement 演示动态插件管理
func (example *PluginExample) demonstrateDynamicPluginManagement() {
	// 禁用插件
	err := example.manager.DisablePlugin("sqllog")
	if err == nil {
		fmt.Println("    ✓ SQL日志插件已禁用")
	}

	// 重新启用插件
	err = example.manager.EnablePlugin("sqllog")
	if err == nil {
		fmt.Println("    ✓ SQL日志插件已重新启用")
	}

	// 查看插件状态
	status := example.manager.GetPluginStatus()
	fmt.Printf("    当前插件状态: %d个插件已注册，%d个插件已启用\n",
		len(status), len(example.manager.GetEnabledPlugins()))
}

// demonstrateConfigurationLoading 演示配置加载
func (example *PluginExample) demonstrateConfigurationLoading() {
	// 创建插件配置
	pluginConfig := &PluginConfiguration{
		Enabled: true,
		Plugins: []PluginConfig{
			{
				Name:    "pagination",
				Enabled: true,
				Order:   1,
				Properties: map[string]any{
					"defaultPageSize": 20,
					"maxPageSize":     500,
				},
			},
			{
				Name:    "performance",
				Enabled: true,
				Order:   2,
				Properties: map[string]any{
					"slowQueryThreshold": 1000,
					"enableMetrics":      true,
				},
			},
		},
	}

	// 加载配置
	err := example.manager.LoadConfiguration(pluginConfig)
	if err == nil {
		fmt.Println("    ✓ 插件配置加载成功")
	} else {
		fmt.Printf("    ✗ 插件配置加载失败: %v\n", err)
	}
}

// CustomAuditPlugin 自定义审计插件
type CustomAuditPlugin struct {
	*BasePlugin
	auditLog []AuditRecord
}

// AuditRecord 审计记录
type AuditRecord struct {
	Timestamp time.Time
	Method    string
	User      string
	Action    string
	Details   map[string]any
}

// Intercept 拦截方法调用
func (plugin *CustomAuditPlugin) Intercept(invocation *Invocation) (any, error) {
	// 记录审计日志
	record := AuditRecord{
		Timestamp: time.Now(),
		Method:    invocation.Method.Name,
		User:      "system", // 实际应该从上下文获取
		Action:    "execute",
		Details: map[string]any{
			"args": invocation.Args,
		},
	}

	plugin.auditLog = append(plugin.auditLog, record)

	// 执行原方法
	result, err := invocation.Proceed()

	// 记录执行结果
	record.Details["success"] = err == nil
	if err != nil {
		record.Details["error"] = err.Error()
	}

	return result, err
}

// Plugin 包装目标对象
func (plugin *CustomAuditPlugin) Plugin(target any) any {
	return target
}

// GetAuditLog 获取审计日志
func (plugin *CustomAuditPlugin) GetAuditLog() []AuditRecord {
	return plugin.auditLog
}

// RunExample 运行示例
func RunExample() {
	example := NewPluginExample()

	// 基本用法演示
	example.DemoBasicUsage()

	// 高级用法演示
	example.DemoAdvancedUsage()

	fmt.Println("\n=== 插件系统演示完成 ===")
}