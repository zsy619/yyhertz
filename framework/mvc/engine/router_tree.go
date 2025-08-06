package engine

import (
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// Params 路由参数类型
type Params []Param

// Param 单个参数
type Param struct {
	Key   string
	Value string
}

// ByName 根据名称获取参数值
func (ps Params) ByName(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// RouterTree 简化的高性能路由树
type RouterTree struct {
	root  *RouterNode
	cache *RouterCache
	mu    sync.RWMutex
}

// RouterNode 路由节点
type RouterNode struct {
	path       string                          // 路径段
	isParam    bool                           // 是否为参数节点(:id)
	isCatchAll bool                           // 是否为捕获所有节点(*path)
	paramName  string                         // 参数名称
	handlers   map[string]core.HandlerFunc    // HTTP方法对应的处理器
	children   map[string]*RouterNode         // 子节点映射
	paramChild *RouterNode                    // 参数子节点
	catchChild *RouterNode                    // 捕获所有子节点
}

// RouterCache 路由缓存
type RouterCache struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
	max   int
}

// CacheEntry 缓存条目
type CacheEntry struct {
	handler core.HandlerFunc
	params  Params
}

// NewRouterTree 创建路由树
func NewRouterTree() *RouterTree {
	return &RouterTree{
		root: &RouterNode{
			handlers: make(map[string]core.HandlerFunc),
			children: make(map[string]*RouterNode),
		},
		cache: NewRouterCache(1000),
	}
}

// NewRouterCache 创建路由缓存
func NewRouterCache(maxSize int) *RouterCache {
	return &RouterCache{
		cache: make(map[string]*CacheEntry),
		max:   maxSize,
	}
}

// AddRoute 添加路由
func (tree *RouterTree) AddRoute(method, path string, handler core.HandlerFunc) {
	if path == "" || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	tree.mu.Lock()
	defer tree.mu.Unlock()

	segments := splitPath(path)
	current := tree.root

	// 遍历路径段
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		if segment[0] == ':' {
			// 参数节点
			paramName := segment[1:]
			if current.paramChild == nil {
				current.paramChild = &RouterNode{
					path:      segment,
					isParam:   true,
					paramName: paramName,
					handlers:  make(map[string]core.HandlerFunc),
					children:  make(map[string]*RouterNode),
				}
			}
			current = current.paramChild
		} else if segment[0] == '*' {
			// 捕获所有节点
			paramName := segment[1:]
			if current.catchChild == nil {
				current.catchChild = &RouterNode{
					path:       segment,
					isCatchAll: true,
					paramName:  paramName,
					handlers:   make(map[string]core.HandlerFunc),
					children:   make(map[string]*RouterNode),
				}
			}
			current = current.catchChild
		} else {
			// 静态节点
			child, exists := current.children[segment]
			if !exists {
				child = &RouterNode{
					path:     segment,
					handlers: make(map[string]core.HandlerFunc),
					children: make(map[string]*RouterNode),
				}
				current.children[segment] = child
			}
			current = child
		}
	}

	// 设置处理器
	if current.handlers[method] != nil {
		panic("路由冲突: " + method + " " + path)
	}
	current.handlers[method] = handler
}

// GetRoute 获取路由处理器
func (tree *RouterTree) GetRoute(method, path string) (core.HandlerFunc, Params) {
	// 检查缓存
	cacheKey := method + ":" + path
	if entry := tree.cache.Get(cacheKey); entry != nil {
		return entry.handler, entry.params
	}

	tree.mu.RLock()
	handler, params := tree.search(method, path)
	tree.mu.RUnlock()

	// 缓存结果
	if handler != nil {
		tree.cache.Set(cacheKey, &CacheEntry{
			handler: handler,
			params:  params,
		})
	}

	return handler, params
}

// search 搜索路由
func (tree *RouterTree) search(method, path string) (core.HandlerFunc, Params) {
	segments := splitPath(path)
	current := tree.root
	var params Params

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// 优先匹配静态路径
		if child, exists := current.children[segment]; exists {
			current = child
			continue
		}

		// 匹配参数路径
		if current.paramChild != nil {
			params = append(params, Param{
				Key:   current.paramChild.paramName,
				Value: segment,
			})
			current = current.paramChild
			continue
		}

		// 匹配捕获所有路径
		if current.catchChild != nil {
			// 收集剩余路径
			remaining := strings.Join(segments[indexOf(segments, segment):], "/")
			params = append(params, Param{
				Key:   current.catchChild.paramName,
				Value: remaining,
			})
			current = current.catchChild
			break
		}

		// 没有找到匹配
		return nil, nil
	}

	// 返回处理器
	handler := current.handlers[method]
	if handler == nil {
		handler = current.handlers["ANY"]
	}

	return handler, params
}

// Compile 编译路由树（占位符方法）
func (tree *RouterTree) Compile() {
	// 这个简化版本不需要特殊的编译步骤
}

// Get/Set 缓存方法
func (cache *RouterCache) Get(key string) *CacheEntry {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.cache[key]
}

func (cache *RouterCache) Set(key string, entry *CacheEntry) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if len(cache.cache) >= cache.max {
		// 简单清理：删除第一个元素
		for k := range cache.cache {
			delete(cache.cache, k)
			break
		}
	}

	cache.cache[key] = entry
}

// 辅助函数
func splitPath(path string) []string {
	if path == "/" {
		return []string{}
	}
	return strings.Split(strings.Trim(path, "/"), "/")
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}