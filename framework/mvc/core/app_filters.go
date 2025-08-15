package core

import (
	"strings"
	"github.com/zsy619/yyhertz/framework/constant"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 过滤器管理方法 =============

// InsertFilter 插入过滤器到指定位置
// 参数：pattern - 路径模式 (支持通配符 *)
//
//	position - 过滤器位置 (BeforeStatic, BeforeRouter, BeforeExec, AfterExec, FinishRouter)
//	filter - 过滤器函数
//	params - 可选参数 (第一个bool值表示是否启用，默认true)
func (app *App) InsertFilter(pattern string, position int, filter FilterFunc, params ...bool) {
	// 验证位置参数
	if !constant.IsValidFilterPosition(position) {
		app.LogErrorf("Invalid filter position: %d", position)
		return
	}

	// 处理可选参数
	enabled := true
	if len(params) > 0 {
		enabled = params[0]
	}

	app.filtersMutex.Lock()
	defer app.filtersMutex.Unlock()

	// 创建过滤器模式
	filterPattern := &FilterPattern{
		Pattern:  pattern,
		Position: position,
		Filter:   filter,
		Enabled:  enabled,
		Priority: int(app.nextFilterID), // 使用ID作为优先级，保证插入顺序
	}

	// 添加到对应位置的过滤器列表
	app.filters[position] = append(app.filters[position], filterPattern)
	app.nextFilterID++

	app.LogInfof("Filter inserted: pattern=%s, position=%d", pattern, position)
}

// RemoveFilter 移除指定模式和位置的过滤器
func (app *App) RemoveFilter(pattern string, position int) bool {
	app.filtersMutex.Lock()
	defer app.filtersMutex.Unlock()

	filters := app.filters[position]
	for i, filter := range filters {
		if filter.Pattern == pattern {
			// 从切片中移除
			app.filters[position] = append(filters[:i], filters[i+1:]...)
			app.LogInfof("Filter removed: pattern=%s, position=%d", pattern, position)
			return true
		}
	}

	return false
}

// ListFilters 列出指定位置的所有过滤器
func (app *App) ListFilters(position int) []*FilterPattern {
	app.filtersMutex.RLock()
	defer app.filtersMutex.RUnlock()

	filters := app.filters[position]
	// 返回副本，避免并发修改
	result := make([]*FilterPattern, len(filters))
	copy(result, filters)
	return result
}

// GetAllFilters 获取所有位置的过滤器
func (app *App) GetAllFilters() map[int][]*FilterPattern {
	app.filtersMutex.RLock()
	defer app.filtersMutex.RUnlock()

	// 创建深度副本
	result := make(map[int][]*FilterPattern)
	for position, filters := range app.filters {
		result[position] = make([]*FilterPattern, len(filters))
		copy(result[position], filters)
	}

	return result
}

// matchPattern 检查路径是否匹配模式 (支持 * 通配符)
func (app *App) matchPattern(pattern, path string) bool {
	// 简单的通配符匹配实现
	if pattern == "*" || pattern == "/*" {
		return true
	}

	// 精确匹配
	if pattern == path {
		return true
	}

	// 通配符匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	// 中间通配符支持
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(path, parts[0]) && strings.HasSuffix(path, parts[1])
		}
	}

	return false
}

// ExecuteFilters 执行指定位置的过滤器
func (app *App) ExecuteFilters(ctx *contextenhanced.Context, position int) {
	app.filtersMutex.RLock()
	filters := app.filters[position]
	app.filtersMutex.RUnlock()

	if len(filters) == 0 {
		return
	}

	// 获取请求路径
	path := string(ctx.Request().Path())

	// 执行匹配的过滤器
	for _, filter := range filters {
		if !filter.Enabled {
			continue
		}

		// 检查路径是否匹配
		if app.matchPattern(filter.Pattern, path) {
			// 执行过滤器
			filter.Filter(ctx)

			// 如果请求被中止，停止执行后续过滤器
			if ctx.IsAborted() {
				break
			}
		}
	}
}