// Package view - 路径映射和约定管理功能
//
// 本文件包含了模板引擎的路径映射和约定管理相关功能：
// - 自动模板推导 (AutoTemplateInference)
// - 路径模板匹配 (PathTemplate)
// - 路径解析和标准化
// - 命名约定转换
package view

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= Beego自动推导功能 =============

// AutoTemplateInference Beego风格的自动模板推导器
type AutoTemplateInference struct {
	controllerMappings map[string]string // 控制器名映射
	actionMappings     map[string]string // 动作名映射
	pathTemplates      []PathTemplate    // 路径模板
	mu                 sync.RWMutex
}

// PathTemplate 路径模板结构
type PathTemplate struct {
	Pattern     string // 路径模式，如 "{controller}/{action}"
	Priority    int    // 优先级，数字越大优先级越高
	Description string // 描述
}

var (
	globalInference     *AutoTemplateInference
	globalInferenceOnce sync.Once
)

// GetAutoTemplateInference 获取全局自动推导器
func GetAutoTemplateInference() *AutoTemplateInference {
	globalInferenceOnce.Do(func() {
		globalInference = NewAutoTemplateInference()
	})
	return globalInference
}

// NewAutoTemplateInference 创建自动推导器
func NewAutoTemplateInference() *AutoTemplateInference {
	inference := &AutoTemplateInference{
		controllerMappings: make(map[string]string),
		actionMappings:     make(map[string]string),
		pathTemplates: []PathTemplate{
			// 默认的Beego路径模板，按优先级排序
			{Pattern: "{controller}/{action}", Priority: 100, Description: "标准Beego模式"},
			{Pattern: "{controller}/{action}/index", Priority: 90, Description: "带index后缀"},
			{Pattern: "{controller}/index", Priority: 80, Description: "控制器index"},
			{Pattern: "{action}", Priority: 70, Description: "仅动作名"},
			{Pattern: "shared/{action}", Priority: 60, Description: "共享模板"},
			{Pattern: "common/{action}", Priority: 50, Description: "通用模板"},
		},
	}

	// 初始化默认映射
	inference.initDefaultMappings()
	return inference
}

// initDefaultMappings 初始化默认映射
func (ati *AutoTemplateInference) initDefaultMappings() {
	// 常见控制器名映射（CamelCase -> lowercase）
	commonControllers := map[string]string{
		"UserController":    "user",
		"AdminController":   "admin",
		"LoginController":   "login",
		"IndexController":   "index",
		"HomeController":    "home",
		"ProductController": "product",
		"OrderController":   "order",
		"SystemController":  "system",
		"ConfigController":  "config",
		"ProfileController": "profile",
	}

	// 常见动作名映射
	commonActions := map[string]string{
		"Index":   "index",
		"List":    "list",
		"Show":    "show",
		"Edit":    "edit",
		"Create":  "create",
		"Update":  "update",
		"Delete":  "delete",
		"Login":   "login",
		"Logout":  "logout",
		"Profile": "profile",
		"Setting": "setting",
		"Detail":  "detail",
		"Search":  "search",
	}

	ati.mu.Lock()
	defer ati.mu.Unlock()

	maps.Copy(ati.controllerMappings, commonControllers)
	maps.Copy(ati.actionMappings, commonActions)
}

// InferTemplatePath 推导模板路径（核心功能 - 增强版）
func (ati *AutoTemplateInference) InferTemplatePath(controllerName, actionName string) []string {
	ati.mu.RLock()
	defer ati.mu.RUnlock()

	// 获取约定管理器
	conventionManager := GetConventionManager()

	// 首先尝试直接映射
	directKey := fmt.Sprintf("%s.%s", controllerName, actionName)
	if mapped, exists := ati.controllerMappings[directKey]; exists {
		return []string{mapped}
	}

	// 应用命名约定进行标准化
	normalizedController := ati.normalizeControllerName(controllerName)
	normalizedAction := ati.normalizeActionName(actionName)

	// 使用约定管理器进一步标准化
	conventionInput := fmt.Sprintf("%s.%s", controllerName, actionName)
	conventionPath := conventionManager.ApplyConvention(conventionInput, conventionManager.GetDefaultConvention())

	var candidates []string

	// 1. 添加约定管理器的结果（优先级最高）
	if conventionPath != conventionInput {
		candidates = append(candidates, conventionPath)
	}

	// 2. 根据路径模板生成候选路径
	for _, template := range ati.pathTemplates {
		path := ati.applyPathTemplate(template.Pattern, normalizedController, normalizedAction)
		if path != "" {
			candidates = append(candidates, path)
		}
	}

	// 3. 添加直接组合的路径
	if normalizedController != "" && normalizedAction != "" {
		candidates = append(candidates, fmt.Sprintf("%s/%s", normalizedController, normalizedAction))
	}

	// 4. 添加其他约定的变体
	for _, convention := range []NamingConvention{SnakeCase, KebabCase, LowerCase} {
		altPath := conventionManager.ApplyConvention(conventionInput, convention)
		if altPath != conventionInput && altPath != conventionPath {
			candidates = append(candidates, altPath)
		}
	}

	// 去重并保持顺序
	return ati.deduplicatePaths(candidates)
}

// normalizeControllerName 标准化控制器名
func (ati *AutoTemplateInference) normalizeControllerName(controllerName string) string {
	if controllerName == "" {
		return ""
	}

	// 检查是否有直接映射
	if mapped, exists := ati.controllerMappings[controllerName]; exists {
		return mapped
	}

	// 移除Controller后缀
	name := strings.TrimSuffix(controllerName, "Controller")

	// CamelCase转换为lowercase
	return ati.camelToLowercase(name)
}

// normalizeActionName 标准化动作名
func (ati *AutoTemplateInference) normalizeActionName(actionName string) string {
	if actionName == "" {
		return "index" // 默认动作
	}

	// 检查是否有直接映射
	if mapped, exists := ati.actionMappings[actionName]; exists {
		return mapped
	}

	// CamelCase转换为lowercase
	return ati.camelToLowercase(actionName)
}

// camelToLowercase 将CamelCase转换为lowercase
func (ati *AutoTemplateInference) camelToLowercase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			_, _ = result.WriteRune('_') // 忽略写入错误
		}
		if r >= 'A' && r <= 'Z' {
			_, _ = result.WriteRune(r - 'A' + 'a') // 忽略写入错误
		} else {
			_, _ = result.WriteRune(r) // 忽略写入错误
		}
	}
	return strings.ToLower(result.String())
}

// applyPathTemplate 应用路径模板
func (ati *AutoTemplateInference) applyPathTemplate(template, controller, action string) string {
	path := template
	path = strings.ReplaceAll(path, "{controller}", controller)
	path = strings.ReplaceAll(path, "{action}", action)

	// 清理路径
	path = strings.Trim(path, "/")
	path = strings.ReplaceAll(path, "//", "/")

	// 验证路径有效性
	if path == "" || strings.Contains(path, "{") {
		return "" // 模板变量未完全替换
	}

	return path
}

// deduplicatePaths 去重路径但保持顺序
func (ati *AutoTemplateInference) deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, path := range paths {
		if path != "" && !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}

	return result
}

// AddControllerMapping 添加控制器映射
func (ati *AutoTemplateInference) AddControllerMapping(controllerName, templatePath string) {
	ati.mu.Lock()
	defer ati.mu.Unlock()
	ati.controllerMappings[controllerName] = templatePath
	config.Debugf("Added controller mapping: %s -> %s", controllerName, templatePath)
}

// AddActionMapping 添加动作映射
func (ati *AutoTemplateInference) AddActionMapping(actionName, templatePath string) {
	ati.mu.Lock()
	defer ati.mu.Unlock()
	ati.actionMappings[actionName] = templatePath
	config.Debugf("Added action mapping: %s -> %s", actionName, templatePath)
}

// AddPathTemplate 添加路径模板
func (ati *AutoTemplateInference) AddPathTemplate(pattern string, priority int, description string) {
	ati.mu.Lock()
	defer ati.mu.Unlock()

	template := PathTemplate{
		Pattern:     pattern,
		Priority:    priority,
		Description: description,
	}

	// 插入到正确位置以保持优先级顺序
	inserted := false
	for i, existing := range ati.pathTemplates {
		if template.Priority > existing.Priority {
			ati.pathTemplates = append(ati.pathTemplates[:i], append([]PathTemplate{template}, ati.pathTemplates[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		ati.pathTemplates = append(ati.pathTemplates, template)
	}

	config.Debugf("Added path template: %s (priority: %d)", pattern, priority)
}

// ============= 路径解析辅助函数 =============

// resolveViewPaths 解析视图路径列表，将相对路径转换为绝对路径
func resolveViewPaths(viewPaths []string) ([]string, error) {
	resolvedPaths := make([]string, 0, len(viewPaths))

	for _, path := range viewPaths {
		absolutePath, err := resolveAbsolutePath(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve view path '%s': %w", path, err)
		}
		resolvedPaths = append(resolvedPaths, absolutePath)
	}

	return resolvedPaths, nil
}

// resolveAbsolutePath 解析单个路径，将相对路径转换为绝对路径
func resolveAbsolutePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// 将相对路径转换为绝对路径
	absolutePath := filepath.Join(cwd, path)
	return filepath.Clean(absolutePath), nil
}

// getCurrentWorkingDir 获取当前工作目录
func getCurrentWorkingDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return cwd
}

// loadCurrentTemplateConfig 加载当前的模板配置
// 尝试从配置文件加载最新的模板配置
func loadCurrentTemplateConfig() (*config.TemplateConfig, error) {
	// 尝试从配置文件加载
	if cfg, err := config.GetTemplateConfig(); err == nil && cfg != nil {
		config.Infof("✅ Loaded template config from file")
		return cfg, nil
	}

	// 如果有全局配置，使用全局配置
	if config.GlobalTemplate != nil {
		config.Infof("✅ Using global template config")
		return config.GlobalTemplate, nil
	}

	// 最后使用默认配置
	cfg := DefaultTemplateConfig()
	config.Infof("✅ Using default template config")
	return cfg, nil
}

// ============= 路径查找和自动推导功能 =============

// RenderWithAutoInference Beego风格的自动推导渲染
func RenderWithAutoInference(controllerName, actionName string, data any) (string, error) {
	// 获取自动推导器
	inference := GetAutoTemplateInference()

	// 推导候选模板路径
	candidates := inference.InferTemplatePath(controllerName, actionName)
	if len(candidates) == 0 {
		return "", fmt.Errorf("no template candidates found for controller: %s, action: %s", controllerName, actionName)
	}

	// 获取模板管理器
	manager := GetEnhancedTemplateManager()

	// 尝试按优先级渲染候选模板
	var lastErr error
	for _, templatePath := range candidates {
		config.Debugf("Trying template path: %s", templatePath)

		result, err := manager.Render(templatePath, data)
		if err == nil {
			config.Debugf("Successfully rendered template: %s", templatePath)
			return result, nil
		}

		lastErr = err
		config.Debugf("Failed to render template %s: %v", templatePath, err)
	}

	// 如果所有候选模板都失败，返回最后一个错误
	return "", fmt.Errorf("failed to render any candidate template for %s.%s, last error: %w, candidates tried: %v",
		controllerName, actionName, lastErr, candidates)
}