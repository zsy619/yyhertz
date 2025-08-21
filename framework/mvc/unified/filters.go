package unified

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// FilterResult 过滤器执行结果
type FilterResult int

const (
	// FilterContinue 继续执行后续过滤器和处理器
	FilterContinue FilterResult = iota
	// FilterStop 停止执行，不再调用后续过滤器和处理器
	FilterStop
	// FilterSkip 跳过当前请求处理，但继续执行后续过滤器
	FilterSkip
)

// FilterFunc 过滤器函数类型
//
// 过滤器函数接收上下文参数，返回过滤结果和可能的错误。
// 
// 参数：
//   - ctx: 请求上下文
//
// 返回：
//   - FilterResult: 过滤结果（继续/停止/跳过）
//   - error: 执行错误（如果有）
//
// 示例：
//
//	var authFilter FilterFunc = func(ctx *context.Context) (FilterResult, error) {
//	    if !isAuthenticated(ctx) {
//	        return FilterStop, errors.New("authentication required")
//	    }
//	    return FilterContinue, nil
//	}
type FilterFunc func(ctx *context.Context) (FilterResult, error)

// Filter 过滤器定义
//
// 包含过滤器的完整信息，包括模式匹配、优先级等。
type Filter struct {
	Name        string     // 过滤器名称
	Pattern     string     // 路径匹配模式（支持正则表达式）
	FilterFunc  FilterFunc // 过滤器函数
	Priority    int        // 优先级（数值越小优先级越高）
	Enabled     bool       // 是否启用
	Description string     // 描述信息
	
	// 编译后的正则表达式（内部使用）
	compiledRegex *regexp.Regexp
}

// FilterChain 过滤器链
//
// 管理一组过滤器的顺序执行。
type FilterChain struct {
	filters []*Filter // 过滤器列表（按优先级排序）
}

// NewFilterChain 创建新的过滤器链
func NewFilterChain() *FilterChain {
	return &FilterChain{
		filters: make([]*Filter, 0),
	}
}

// AddFilter 添加过滤器到链中
//
// 过滤器会根据优先级自动排序，优先级数值越小越先执行。
//
// 参数：
//   - filter: 要添加的过滤器
//
// 返回：
//   - error: 添加失败时返回错误
//
// 示例：
//
//	chain := NewFilterChain()
//	err := chain.AddFilter(&Filter{
//	    Name: "auth",
//	    Pattern: "/api/*",
//	    FilterFunc: authFilter,
//	    Priority: 100,
//	    Enabled: true,
//	})
func (fc *FilterChain) AddFilter(filter *Filter) error {
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}
	
	// 检查是否已存在同名过滤器
	for _, f := range fc.filters {
		if f.Name == filter.Name {
			return fmt.Errorf("filter with name '%s' already exists", filter.Name)
		}
	}
	
	// 编译正则表达式
	if filter.Pattern != "" {
		regex, err := regexp.Compile(filter.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", filter.Pattern, err)
		}
		filter.compiledRegex = regex
	}
	
	// 添加到链中
	fc.filters = append(fc.filters, filter)
	
	// 按优先级排序
	fc.sortFilters()
	
	return nil
}

// RemoveFilter 从链中移除过滤器
//
// 参数：
//   - name: 要移除的过滤器名称
//
// 返回：
//   - bool: 是否成功移除
func (fc *FilterChain) RemoveFilter(name string) bool {
	for i, filter := range fc.filters {
		if filter.Name == name {
			// 从切片中移除
			fc.filters = append(fc.filters[:i], fc.filters[i+1:]...)
			return true
		}
	}
	return false
}

// GetFilter 获取指定名称的过滤器
func (fc *FilterChain) GetFilter(name string) *Filter {
	for _, filter := range fc.filters {
		if filter.Name == name {
			return filter
		}
	}
	return nil
}

// Execute 执行过滤器链
//
// 按优先级顺序执行所有匹配的过滤器。
// 如果某个过滤器返回FilterStop，则停止执行后续过滤器。
//
// 参数：
//   - ctx: 请求上下文
//   - path: 请求路径（用于模式匹配）
//
// 返回：
//   - FilterResult: 最终执行结果
//   - error: 执行过程中的错误
//
// 示例：
//
//	result, err := chain.Execute(ctx, "/api/users")
//	if err != nil {
//	    // 处理过滤器执行错误
//	    return err
//	}
//	if result == FilterStop {
//	    // 请求被过滤器拦截
//	    return
//	}
func (fc *FilterChain) Execute(ctx *context.Context, path string) (FilterResult, error) {
	for _, filter := range fc.filters {
		// 检查过滤器是否启用
		if !filter.Enabled {
			continue
		}
		
		// 检查路径是否匹配
		if !fc.matchPattern(filter, path) {
			continue
		}
		
		// 执行过滤器
		result, err := filter.FilterFunc(ctx)
		if err != nil {
			return FilterStop, fmt.Errorf("filter '%s' error: %w", filter.Name, err)
		}
		
		// 根据结果决定是否继续
		switch result {
		case FilterStop:
			return FilterStop, nil
		case FilterSkip:
			return FilterSkip, nil
		case FilterContinue:
			continue
		}
	}
	
	return FilterContinue, nil
}

// matchPattern 检查路径是否匹配过滤器模式
func (fc *FilterChain) matchPattern(filter *Filter, path string) bool {
	if filter.Pattern == "" || filter.Pattern == "*" {
		return true
	}
	
	if filter.compiledRegex != nil {
		return filter.compiledRegex.MatchString(path)
	}
	
	// 简单的通配符匹配
	if filter.Pattern == "/*" {
		return true
	}
	
	return filter.Pattern == path
}

// sortFilters 按优先级对过滤器进行排序
func (fc *FilterChain) sortFilters() {
	sort.Slice(fc.filters, func(i, j int) bool {
		return fc.filters[i].Priority < fc.filters[j].Priority
	})
}

// GetFilters 获取所有过滤器（只读）
func (fc *FilterChain) GetFilters() []*Filter {
	// 返回副本，避免外部修改
	result := make([]*Filter, len(fc.filters))
	copy(result, fc.filters)
	return result
}

// GetEnabledFilters 获取所有启用的过滤器
func (fc *FilterChain) GetEnabledFilters() []*Filter {
	var result []*Filter
	for _, filter := range fc.filters {
		if filter.Enabled {
			result = append(result, filter)
		}
	}
	return result
}

// Clear 清空过滤器链
func (fc *FilterChain) Clear() {
	fc.filters = fc.filters[:0]
}

// Size 获取过滤器数量
func (fc *FilterChain) Size() int {
	return len(fc.filters)
}

// EnableFilter 启用指定过滤器
func (fc *FilterChain) EnableFilter(name string) bool {
	if filter := fc.GetFilter(name); filter != nil {
		filter.Enabled = true
		return true
	}
	return false
}

// DisableFilter 禁用指定过滤器
func (fc *FilterChain) DisableFilter(name string) bool {
	if filter := fc.GetFilter(name); filter != nil {
		filter.Enabled = false
		return true
	}
	return false
}

// ============= Manager 中的过滤器管理方法 =============

// AddFilter 向统一管理器添加过滤器
//
// 参数：
//   - filter: 要添加的过滤器
//
// 返回：
//   - error: 添加失败时返回错误
//
// 示例：
//
//	manager := GetManager()
//	err := manager.AddFilter(&Filter{
//	    Name: "sso",
//	    Pattern: "/*",
//	    FilterFunc: FilterSSO,
//	    Priority: 10,
//	    Enabled: true,
//	    Description: "Single Sign-On filter",
//	})
func (m *Manager) AddFilter(filter *Filter) error {
	m.filterMutex.Lock()
	defer m.filterMutex.Unlock()
	
	// 初始化过滤器列表（如果尚未初始化）
	if m.filters == nil {
		m.filters = make([]Filter, 0)
	}
	
	// 检查是否已存在同名过滤器
	for _, f := range m.filters {
		if f.Name == filter.Name {
			return fmt.Errorf("filter with name '%s' already exists", filter.Name)
		}
	}
	
	// 编译正则表达式
	if filter.Pattern != "" {
		regex, err := regexp.Compile(filter.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", filter.Pattern, err)
		}
		filter.compiledRegex = regex
	}
	
	// 添加过滤器
	m.filters = append(m.filters, *filter)
	
	// 按优先级排序
	m.sortFilters()
	
	return nil
}

// RemoveFilter 从统一管理器移除过滤器
func (m *Manager) RemoveFilter(name string) bool {
	m.filterMutex.Lock()
	defer m.filterMutex.Unlock()
	
	for i, filter := range m.filters {
		if filter.Name == name {
			// 从切片中移除
			m.filters = append(m.filters[:i], m.filters[i+1:]...)
			return true
		}
	}
	return false
}

// GetFilter 获取指定名称的过滤器
func (m *Manager) GetFilter(name string) *Filter {
	m.filterMutex.RLock()
	defer m.filterMutex.RUnlock()
	
	for _, filter := range m.filters {
		if filter.Name == name {
			// 返回副本，避免外部修改
			filterCopy := filter
			return &filterCopy
		}
	}
	return nil
}

// ExecuteFilters 执行所有匹配的过滤器
func (m *Manager) ExecuteFilters(ctx *context.Context, path string) (FilterResult, error) {
	m.filterMutex.RLock()
	defer m.filterMutex.RUnlock()
	
	for _, filter := range m.filters {
		// 检查过滤器是否启用
		if !filter.Enabled {
			continue
		}
		
		// 检查路径是否匹配
		if !m.matchFilterPattern(&filter, path) {
			continue
		}
		
		// 执行过滤器
		result, err := filter.FilterFunc(ctx)
		if err != nil {
			return FilterStop, fmt.Errorf("filter '%s' error: %w", filter.Name, err)
		}
		
		// 根据结果决定是否继续
		switch result {
		case FilterStop:
			return FilterStop, nil
		case FilterSkip:
			return FilterSkip, nil
		case FilterContinue:
			continue
		}
	}
	
	return FilterContinue, nil
}

// matchFilterPattern 检查路径是否匹配过滤器模式
func (m *Manager) matchFilterPattern(filter *Filter, path string) bool {
	if filter.Pattern == "" || filter.Pattern == "*" {
		return true
	}
	
	if filter.compiledRegex != nil {
		return filter.compiledRegex.MatchString(path)
	}
	
	// 简单的通配符匹配
	if filter.Pattern == "/*" {
		return true
	}
	
	return filter.Pattern == path
}

// sortFilters 按优先级对过滤器进行排序
func (m *Manager) sortFilters() {
	sort.Slice(m.filters, func(i, j int) bool {
		return m.filters[i].Priority < m.filters[j].Priority
	})
}

// GetAllFilters 获取所有过滤器
func (m *Manager) GetAllFilters() []Filter {
	m.filterMutex.RLock()
	defer m.filterMutex.RUnlock()
	
	// 返回副本，避免外部修改
	result := make([]Filter, len(m.filters))
	copy(result, m.filters)
	return result
}

// EnableFilter 启用指定过滤器
func (m *Manager) EnableFilter(name string) bool {
	m.filterMutex.Lock()
	defer m.filterMutex.Unlock()
	
	for i, filter := range m.filters {
		if filter.Name == name {
			m.filters[i].Enabled = true
			return true
		}
	}
	return false
}

// DisableFilter 禁用指定过滤器
func (m *Manager) DisableFilter(name string) bool {
	m.filterMutex.Lock()
	defer m.filterMutex.Unlock()
	
	for i, filter := range m.filters {
		if filter.Name == name {
			m.filters[i].Enabled = false
			return true
		}
	}
	return false
}