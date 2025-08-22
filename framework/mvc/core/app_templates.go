package core

import (
	"html/template"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// ============= 模板函数管理方法（重构版本 - 直接使用 view 引擎） =============

// AddFuncMap 添加全局模板函数
// 参数：name - 函数名字符串，fn - 函数实现
// 示例：AddFuncMap("containString", tool.ContainString)
//
// 注意：此方法已重构为直接使用统一的模板函数管理器，避免重复存储
// 注意：为了优化性能，此方法不会自动重载模板，需要手动调用 ReloadTemplates()
func (app *App) AddFuncMap(name string, fn any) {
	// 直接使用 view 引擎的统一管理器
	view.AddGlobalFunction(name, fn)
	app.LogInfof("Template function registered: %s", name)

	// 移除自动重载以减少重复的模板解析和警告
	// 用户需要在批量注册函数后手动调用 ReloadTemplates()
}

// ReloadTemplates 重新加载模板
func (app *App) ReloadTemplates() {
	// 自动重新加载模板以确保新函数能被识别
	if err := view.ReloadDefaultTemplates(); err != nil {
		app.LogWarnf("Failed to reload templates after registering functions: %v", err)
	} else {
		app.LogInfof("Templates reloaded successfully after registering functions")
	}
}

// GetGlobalFuncMap 获取全局模板函数映射（只读副本）
//
// 注意：此方法已重构为从统一管理器获取数据
func (app *App) GetGlobalFuncMap() template.FuncMap {
	// 从 view 引擎获取全局函数
	return view.GetGlobalFunctions()
}

// RemoveFuncMap 移除全局模板函数
//
// 注意：此方法已重构为直接操作统一管理器
func (app *App) RemoveFuncMap(name string) {
	// 直接使用 view 引擎的统一管理器
	view.RemoveGlobalFunction(name)
	app.LogInfof("Template function removed: %s", name)
}

// ListFuncMap 列出所有已注册的模板函数名称
//
// 注意：此方法已重构为从统一管理器获取数据
func (app *App) ListFuncMap() []string {
	// 从统一管理器获取所有函数
	funcList := view.ListTemplateFunctions()

	// 合并所有类型的函数名称
	allNames := make([]string, 0)
	for _, names := range funcList {
		allNames = append(allNames, names...)
	}

	return allNames
}

// ============= 新增的增强方法 =============

// AddFuncMapGroup 批量添加模板函数组
// 参数：groupName - 函数组名称，funcs - 函数映射
func (app *App) AddFuncMapGroup(groupName string, funcs template.FuncMap) {
	view.RegisterTemplateFunctionGroup(groupName, funcs)
	app.LogInfof("Template function group registered: %s (%d functions)", groupName, len(funcs))
}

// HasFuncMap 检查是否存在指定的模板函数
func (app *App) HasFuncMap(name string) bool {
	return view.HasTemplateFunction(name)
}

// GetFuncMapSource 获取模板函数的来源
func (app *App) GetFuncMapSource(name string) string {
	return string(view.GetTemplateFunctionSource(name))
}

// ListFuncMapBySource 按来源分类列出模板函数
func (app *App) ListFuncMapBySource() map[string][]string {
	return view.ListTemplateFunctions()
}

// GetFuncMapCount 获取模板函数总数
func (app *App) GetFuncMapCount() int {
	funcList := view.ListTemplateFunctions()
	total := 0
	for _, names := range funcList {
		total += len(names)
	}
	return total
}
