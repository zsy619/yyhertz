// Package view 提供增强的模板引擎功能
//
// 这个文件包含 TemplateFunctionManager 相关的功能
package view

import (
	"html/template"
	"maps"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// TemplateFunctionManager 统一的模板函数管理器
type TemplateFunctionManager struct {
	builtinFuncs template.FuncMap // 内置函数（原 TemplateFuncs）
	globalFuncs  template.FuncMap // 全局自定义函数
	mu           sync.RWMutex
}

// FunctionSource 函数来源类型
type FunctionSource string

const (
	FunctionSourceBuiltin FunctionSource = "builtin"
	FunctionSourceGlobal  FunctionSource = "global"
	FunctionSourceEngine  FunctionSource = "engine"
)

var (
	// 全局模板函数管理器
	globalFunctionManager     *TemplateFunctionManager
	globalFunctionManagerOnce sync.Once
)

// NewTemplateFunctionManager 创建新的模板函数管理器
func NewTemplateFunctionManager() *TemplateFunctionManager {
	return &TemplateFunctionManager{
		builtinFuncs: make(template.FuncMap),
		globalFuncs:  make(template.FuncMap),
	}
}

// GetGlobalFunctionManager 获取全局函数管理器
func GetGlobalFunctionManager() *TemplateFunctionManager {
	globalFunctionManagerOnce.Do(func() {
		globalFunctionManager = NewTemplateFunctionManager()
		// 初始化内置函数
		globalFunctionManager.initBuiltinFunctions()
	})
	return globalFunctionManager
}

// initBuiltinFunctions 初始化内置函数
func (tfm *TemplateFunctionManager) initBuiltinFunctions() {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	// 从 template.go 获取内置函数
	builtinFuncs := GetBuiltinTemplateFunctions()
	maps.Copy(tfm.builtinFuncs, builtinFuncs)

	// 添加 Beego 风格的模板函数
	beegoFuncs := GetBeegoTemplateFuncs()
	maps.Copy(tfm.builtinFuncs, beegoFuncs)
}

// AddGlobalFunction 添加全局函数
func (tfm *TemplateFunctionManager) AddGlobalFunction(name string, fn any) {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	if tfm.globalFuncs == nil {
		tfm.globalFuncs = make(template.FuncMap)
	}

	tfm.globalFuncs[name] = fn
	config.Debugf("Global template function added: %s", name)
}

// RemoveGlobalFunction 移除全局函数
func (tfm *TemplateFunctionManager) RemoveGlobalFunction(name string) {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	delete(tfm.globalFuncs, name)
	config.Debugf("Global template function removed: %s", name)
}

// GetMergedFunctions 获取合并后的函数映射
// 优先级: 引擎自定义 > 全局自定义 > 内置函数
func (tfm *TemplateFunctionManager) GetMergedFunctions(engineFuncs template.FuncMap) template.FuncMap {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	merged := make(template.FuncMap)

	// 1. 首先添加内置函数
	maps.Copy(merged, tfm.builtinFuncs)

	// 2. 然后添加全局自定义函数（可能覆盖内置函数）
	for name := range tfm.globalFuncs {
		if _, exists := merged[name]; exists {
			// TODO: 记录警告日志
			// config.Warnf("Global function '%s' overrides builtin function", name)
			// fmt.Printf("Global function '%s' overrides builtin function\n", name)
		}
	}
	maps.Copy(merged, tfm.globalFuncs)

	// 3. 最后添加引擎自定义函数（可能覆盖前面的函数）
	for name, fn := range engineFuncs {
		if _, exists := merged[name]; exists {
			source := tfm.GetFunctionSource(name)
			// config.Warnf("Engine function '%s' overrides %s function", name, source)
			_ = source
		}
		merged[name] = fn
	}

	return merged
}

// GetFunctionSource 获取函数来源
func (tfm *TemplateFunctionManager) GetFunctionSource(name string) FunctionSource {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	if _, exists := tfm.builtinFuncs[name]; exists {
		return FunctionSourceBuiltin
	}
	if _, exists := tfm.globalFuncs[name]; exists {
		return FunctionSourceGlobal
	}
	return FunctionSourceEngine
}

// HasFunction 检查函数是否存在
func (tfm *TemplateFunctionManager) HasFunction(name string) bool {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	_, builtin := tfm.builtinFuncs[name]
	_, global := tfm.globalFuncs[name]
	return builtin || global
}

// ListFunctions 列出所有函数按来源分类
func (tfm *TemplateFunctionManager) ListFunctions() map[string][]string {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	result := map[string][]string{
		"builtin": make([]string, 0, len(tfm.builtinFuncs)),
		"global":  make([]string, 0, len(tfm.globalFuncs)),
	}

	for name := range tfm.builtinFuncs {
		result["builtin"] = append(result["builtin"], name)
	}

	for name := range tfm.globalFuncs {
		result["global"] = append(result["global"], name)
	}

	return result
}

// RegisterFunctionGroup 批量注册函数组
func (tfm *TemplateFunctionManager) RegisterFunctionGroup(groupName string, funcs template.FuncMap) {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	if tfm.globalFuncs == nil {
		tfm.globalFuncs = make(template.FuncMap)
	}

	for name, fn := range funcs {
		tfm.globalFuncs[name] = fn
		config.Debugf("Function '%s' registered in group '%s'", name, groupName)
	}
}

// ============= 向后兼容的全局函数接口 =============

// AddGlobalFunction 添加全局模板函数（向后兼容接口）
func AddGlobalFunction(name string, fn any) {
	manager := GetGlobalFunctionManager()
	manager.AddGlobalFunction(name, fn)
}

// GetGlobalFunctions 获取全局模板函数映射（向后兼容接口）
func GetGlobalFunctions() template.FuncMap {
	manager := GetGlobalFunctionManager()
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	// 创建副本以避免并发修改
	funcMapCopy := make(template.FuncMap, len(manager.globalFuncs))
	maps.Copy(funcMapCopy, manager.globalFuncs)

	return funcMapCopy
}

// RemoveGlobalFunction 移除全局模板函数（向后兼容接口）
func RemoveGlobalFunction(name string) {
	manager := GetGlobalFunctionManager()
	manager.RemoveGlobalFunction(name)
}

// ============= 新的统一管理接口 =============

// GetTemplateFunctionManager 获取模板函数管理器
func GetTemplateFunctionManager() *TemplateFunctionManager {
	return GetGlobalFunctionManager()
}

// RegisterTemplateFunctionGroup 注册模板函数组
func RegisterTemplateFunctionGroup(groupName string, funcs template.FuncMap) {
	manager := GetGlobalFunctionManager()
	manager.RegisterFunctionGroup(groupName, funcs)
}

// ListTemplateFunctions 列出所有模板函数
func ListTemplateFunctions() map[string][]string {
	manager := GetGlobalFunctionManager()
	return manager.ListFunctions()
}

// HasTemplateFunction 检查模板函数是否存在
func HasTemplateFunction(name string) bool {
	manager := GetGlobalFunctionManager()
	return manager.HasFunction(name)
}

// GetTemplateFunctionSource 获取模板函数来源
func GetTemplateFunctionSource(name string) FunctionSource {
	manager := GetGlobalFunctionManager()
	return manager.GetFunctionSource(name)
}

// AddFuncMap 添加模板函数 (Beego风格)
func (e *TemplateEngine) AddFuncMap(name string, fn interface{}) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	e.funcMap[name] = fn

	// 在开发模式下立即重建模板以包含新函数
	if e.RunMode == "dev" {
		go func() {
			// BuildAllTemplates 方法已移至 engine_template.go
			// if err := e.BuildAllTemplates(); err != nil {
			// 	config.Errorf("Failed to rebuild templates after adding function %s: %v", name, err)
			// }
			config.Debugf("Function %s added, template rebuild skipped (method moved to engine_template.go)", name)
		}()
	}
}
