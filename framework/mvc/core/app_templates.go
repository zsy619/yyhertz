package core

import (
	"html/template"

	"github.com/zsy619/yyhertz/framework/view"
)

// ============= 模板函数管理方法 =============

// AddFuncMap 添加全局模板函数
// 参数：name - 函数名字符串，fn - 函数实现
// 示例：AddFuncMap("containString", tool.ContainString)
func (app *App) AddFuncMap(name string, fn any) {
	app.funcMapMutex.Lock()
	defer app.funcMapMutex.Unlock()

	// 添加到应用级别的全局模板函数映射
	app.globalFuncMap[name] = fn

	// 同时添加到view引擎的全局存储中
	view.AddGlobalFunction(name, fn)

	app.LogInfof("Template function registered: %s", name)
}

// GetGlobalFuncMap 获取全局模板函数映射（只读副本）
func (app *App) GetGlobalFuncMap() template.FuncMap {
	app.funcMapMutex.RLock()
	defer app.funcMapMutex.RUnlock()

	// 创建副本以避免并发修改
	funcMapCopy := make(template.FuncMap, len(app.globalFuncMap))
	for name, fn := range app.globalFuncMap {
		funcMapCopy[name] = fn
	}

	return funcMapCopy
}

// RemoveFuncMap 移除全局模板函数
func (app *App) RemoveFuncMap(name string) {
	app.funcMapMutex.Lock()
	defer app.funcMapMutex.Unlock()

	delete(app.globalFuncMap, name)

	// 同时从view引擎的全局存储中移除
	view.RemoveGlobalFunction(name)

	app.LogInfof("Template function removed: %s", name)
}

// ListFuncMap 列出所有已注册的模板函数名称
func (app *App) ListFuncMap() []string {
	app.funcMapMutex.RLock()
	defer app.funcMapMutex.RUnlock()

	names := make([]string, 0, len(app.globalFuncMap))
	for name := range app.globalFuncMap {
		names = append(names, name)
	}

	return names
}
