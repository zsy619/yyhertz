// Package gin - 高性能路由引擎实现
// 基于Radix Tree的零分配路由匹配引擎
package gin

import (
	"net/http"
	"sync"
)

// RouterEngine 高性能路由引擎
type RouterEngine struct {
	trees  methodTrees
	
	// 性能优化配置
	RedirectTrailingSlash  bool
	RedirectFixedPath      bool
	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool
	UseRawPath            bool
	UnescapePathValues    bool
	RemoveExtraSlash      bool
	
	// 错误处理
	NoRoute     HandlerChain
	NoMethod    HandlerChain
	
	// 对象池优化
	pool sync.Pool
	
	// 性能统计
	stats *RouterStats
	
	// 路由约束支持
	constraints *ConstraintRegistry
}

// RouterStats 路由性能统计
type RouterStats struct {
	TotalRequests   uint64
	RouteHits       uint64
	RouteMisses     uint64
	AvgLookupTime   int64  // 纳秒
	ParamAllocated  uint64
	ParamPoolHits   uint64
}

// NewRouterEngine 创建新的路由引擎
func NewRouterEngine() *RouterEngine {
	engine := &RouterEngine{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		ForwardedByClientIP:    true,
		UseRawPath:            false,
		UnescapePathValues:    true,
		RemoveExtraSlash:      false,
		stats:                 &RouterStats{},
	}
	
	// 初始化参数池
	engine.pool.New = func() interface{} {
		return make(Params, 0, 16) // 预分配16个参数的容量
	}
	
	return engine
}

// addRoute 添加路由
func (engine *RouterEngine) addRoute(method, path string, handlers HandlerChain) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}
	if method == "" {
		panic("HTTP method can not be empty")
	}
	if len(handlers) == 0 {
		panic("there must be at least one handler")
	}

	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)

	// 特殊处理：确保根路径总是存在
	if path == "/" {
		engine.addHeadRoute(method, path, handlers)
	}
}

// addHeadRoute 为GET路由自动添加HEAD路由
func (engine *RouterEngine) addHeadRoute(method, path string, handlers HandlerChain) {
	if method == "GET" {
		// 检查是否已有HEAD路由
		if engine.trees.get("HEAD") == nil {
			// 创建HEAD路由树
			headRoot := new(node)
			headRoot.fullPath = "/"
			engine.trees = append(engine.trees, methodTree{method: "HEAD", root: headRoot})
		}
		
		// 为HEAD方法添加相同的处理器
		headRoot := engine.trees.get("HEAD")
		if headRoot != nil {
			headRoot.addRoute(path, handlers)
		}
	}
}

// handleHTTPRequest 处理HTTP请求的核心方法
func (engine *RouterEngine) handleHTTPRequest(c *Context) {
	httpMethod := string(c.RequestContext.Request.Header.Method())
	rPath := string(c.RequestContext.Request.URI().Path())

	if engine.UseRawPath && len(c.RequestContext.Request.URI().QueryString()) > 0 {
		rPath = string(c.RequestContext.Request.URI().PathOriginal())
	}

	if engine.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}

	// 开始路由查找
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		
		root := t[i].root
		// 使用对象池获取参数切片
		value, params, tsr := root.getValue(rPath, engine.getParams)
		
		if value != nil {
			c.handlers = value
			c.Params = *params
			engine.putParams(params)
			c.Next()
			return
		}
		
		if httpMethod != "CONNECT" && rPath != "/" {
			if tsr && engine.RedirectTrailingSlash {
				engine.redirectTrailingSlash(c, rPath)
				return
			}
			
			if engine.RedirectFixedPath {
				if redirectPath, found := root.findCaseInsensitivePath(
					rPath,
					engine.RedirectTrailingSlash,
				); found {
					engine.redirectFixedPath(c, redirectPath)
					return
				}
			}
		}
		
		engine.putParams(params)
		break
	}

	// 处理方法不允许的情况
	if engine.HandleMethodNotAllowed {
		if allow := engine.getMethodNotAllowed(rPath, httpMethod); allow != "" {
			c.Header("Allow", allow)
			if engine.NoMethod != nil {
				c.handlers = engine.NoMethod
				c.Next()
			} else {
				c.AbortWithStatus(http.StatusMethodNotAllowed)
			}
			return
		}
	}

	// 404处理
	if engine.NoRoute != nil {
		c.handlers = engine.NoRoute
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

// getParams 从对象池获取参数切片
func (engine *RouterEngine) getParams() *Params {
	ps := engine.pool.Get().(Params)
	ps = ps[:0] // 重置长度但保持容量
	return &ps
}

// putParams 将参数切片归还给对象池
func (engine *RouterEngine) putParams(params *Params) {
	if params != nil {
		engine.pool.Put(*params)
	}
}

// redirectTrailingSlash 处理尾随斜杠重定向
func (engine *RouterEngine) redirectTrailingSlash(c *Context, rPath string) {
	var code int
	method := string(c.RequestContext.Request.Header.Method())
	
	if method == "GET" || method == "HEAD" {
		code = http.StatusMovedPermanently
	} else {
		code = http.StatusTemporaryRedirect
	}
	
	var location string
	if len(rPath) > 1 && rPath[len(rPath)-1] == '/' {
		location = rPath[:len(rPath)-1]
	} else {
		location = rPath + "/"
	}
	
	// 保持查询参数
	if queryBuf := c.RequestContext.Request.URI().QueryString(); len(queryBuf) > 0 {
		location += "?" + string(queryBuf)
	}
	
	c.Header("Location", location)
	c.AbortWithStatus(code)
}

// redirectFixedPath 处理固定路径重定向
func (engine *RouterEngine) redirectFixedPath(c *Context, fixedPath string) {
	method := string(c.RequestContext.Request.Header.Method())
	code := http.StatusMovedPermanently
	if method != "GET" {
		code = http.StatusTemporaryRedirect
	}
	
	// 保持查询参数
	if queryBuf := c.RequestContext.Request.URI().QueryString(); len(queryBuf) > 0 {
		fixedPath += "?" + string(queryBuf)
	}
	
	c.Header("Location", fixedPath)
	c.AbortWithStatus(code)
}

// getMethodNotAllowed 获取方法不允许的Allow头
func (engine *RouterEngine) getMethodNotAllowed(path, httpMethod string) string {
	var allow string
	
	for _, tree := range engine.trees {
		if tree.method == httpMethod {
			continue
		}
		
		if handlers, _, _ := tree.root.getValue(path, nil); handlers != nil {
			if allow != "" {
				allow += ", "
			}
			allow += tree.method
		}
	}
	
	return allow
}

// cleanPath 清理路径，移除多余的斜杠
func cleanPath(p string) string {
	const stackBufSize = 128
	
	// 处理空路径
	if p == "" {
		return "/"
	}
	
	// 使用栈缓冲区优化小路径
	var buf []byte
	if len(p) <= stackBufSize {
		buf = make([]byte, 0, stackBufSize)
	} else {
		buf = make([]byte, 0, len(p))
	}
	
	// 确保以/开头
	if p[0] != '/' {
		buf = append(buf, '/')
	}
	
	// 清理路径
	n := len(p)
	r := 1
	
	for r < n {
		switch {
		case p[r] == '/':
			// 跳过重复的斜杠
			for r < n && p[r] == '/' {
				r++
			}
			// 添加单个斜杠
			if r < n {
				buf = append(buf, '/')
			}
			
		case p[r] == '.' && (r+1 == n || p[r+1] == '/'):
			// 跳过 ./
			r++
			
		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			// 处理 ../
			if len(buf) > 1 {
				// 回退到上一个目录
				for len(buf) > 1 && buf[len(buf)-1] != '/' {
					buf = buf[:len(buf)-1]
				}
				if len(buf) > 1 {
					buf = buf[:len(buf)-1]
				}
			}
			r += 2
			
		default:
			// 复制字符直到下一个斜杠
			for r < n && p[r] != '/' {
				buf = append(buf, p[r])
				r++
			}
		}
	}
	
	// 确保至少有根路径
	if len(buf) == 0 {
		return "/"
	}
	
	return string(buf)
}

// 注意：RouterGroup方法已在routergroup_methods.go中定义

// 性能监控方法

// GetStats 获取路由性能统计
func (engine *RouterEngine) GetStats() *RouterStats {
	return engine.stats
}

// ResetStats 重置性能统计
func (engine *RouterEngine) ResetStats() {
	engine.stats = &RouterStats{}
}

// 注意：abortIndex常量和IRoutes接口已在其他文件中定义