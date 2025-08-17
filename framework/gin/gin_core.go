// Package gin - 核心类型定义
// 包含所有核心类型定义：HandlerFunc、Context、Engine、RouterGroup、Params等

package gin

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/zsy619/yyhertz/framework/gin/render"
)

// HandlerFunc 定义Gin风格的处理函数类型
//
// 处理函数接收一个Context参数，用于处理HTTP请求和响应。
// 这与原生Gin的HandlerFunc保持完全兼容。
//
// 示例：
//
//	func handler(c *gin.Context) {
//		c.JSON(200, gin.H{"status": "ok"})
//	}
type HandlerFunc func(*Context)

// Context 表示Gin风格的请求上下文
//
// Context封装了HTTP请求和响应的所有信息，提供了处理请求、
// 参数绑定、响应渲染等功能。它包装了Hertz的RequestContext，
// 同时添加了Gin兼容的API和中间件支持。
//
// 主要功能：
//   - 路径参数和查询参数获取
//   - 请求数据绑定（JSON、Query、URI等）
//   - 响应渲染（JSON、HTML、String等）
//   - 中间件执行控制
//   - 键值对存储
//   - 错误收集
type Context struct {
	*app.RequestContext // Hertz原生请求上下文

	// 参数存储
	Params Params         // 路径参数（如：/user/:id中的id）
	Keys   map[string]any // 用户自定义的键值对存储
	Errors []error        // 请求处理过程中收集的错误

	// 中间件控制
	handlers []HandlerFunc // 当前请求的处理函数链
	index    int8          // 当前执行的处理函数索引

	// 引擎引用
	engine *Engine // 所属的引擎实例
}

// responseWriterAdapter 适配器，将Hertz的Response转换为http.ResponseWriter
type responseWriterAdapter struct {
	ctx *Context
}

func (w *responseWriterAdapter) Header() http.Header {
	header := make(http.Header)
	w.ctx.RequestContext.Response.Header.VisitAll(func(key, value []byte) {
		header.Add(string(key), string(value))
	})
	return header
}

func (w *responseWriterAdapter) Write(data []byte) (int, error) {
	n, err := w.ctx.RequestContext.Write(data)
	return n, err
}

func (w *responseWriterAdapter) WriteHeader(statusCode int) {
	w.ctx.RequestContext.SetStatusCode(statusCode)
}

// convertHertzHeaderToHTTP 将Hertz的Header转换为http.Header
func convertHertzHeaderToHTTP(hertzHeader *protocol.RequestHeader) http.Header {
	header := make(http.Header)
	hertzHeader.VisitAll(func(key, value []byte) {
		header.Add(string(key), string(value))
	})
	return header
}

// Writer 返回响应Writer（Gin兼容性方法）
//
// 返回一个http.ResponseWriter接口实例，用于兼容Gin的API。
// 这个方法返回一个适配器，将Hertz的Response包装为标准的http.ResponseWriter。
//
// 返回值：
//   - http.ResponseWriter: 响应写入器接口
func (c *Context) Writer() http.ResponseWriter {
	return &responseWriterAdapter{ctx: c}
}

// Request 返回HTTP请求（Gin兼容性方法）
//
// 返回一个*http.Request实例，用于兼容Gin的API。
// 这个方法将Hertz的Request转换为标准的http.Request，
// 包括正确设置协议版本信息（Proto字段）。
//
// 功能特性：
//   - 完整的URL信息转换（包括Scheme、Host、Path、Query）
//   - HTTP头部信息转换
//   - 请求体数据转换
//   - 协议版本信息提取和设置
//   - 与Gin原版API完全兼容
//
// 返回值：
//   - *http.Request: HTTP请求实例，包含完整的协议版本信息
func (c *Context) Request() *http.Request {
	// 构建URL
	uri := c.RequestContext.Request.URI()
	u := &url.URL{
		Scheme:   string(uri.Scheme()),
		Host:     string(uri.Host()),
		Path:     string(uri.Path()),
		RawQuery: string(uri.QueryString()),
	}

	// 提取协议版本信息
	proto := extractProtocolVersion(c.RequestContext)

	// 创建http.Request
	req := &http.Request{
		Method: string(c.RequestContext.Request.Method()),
		URL:    u,
		Proto:  proto,                                                          // 设置协议版本
		Header: convertHertzHeaderToHTTP(&c.RequestContext.Request.Header),
		Body:   io.NopCloser(bytes.NewReader(c.RequestContext.Request.Body())),
		Host:   string(uri.Host()),
	}

	return req
}

// extractProtocolVersion 从Hertz RequestContext中提取HTTP协议版本
//
// 此函数分析Hertz请求上下文，提取并格式化HTTP协议版本信息。
// 协议版本用于日志记录、性能分析和客户端兼容性检查。
//
// 支持的协议版本：
//   - HTTP/1.0: 传统HTTP 1.0协议
//   - HTTP/1.1: 标准HTTP 1.1协议（最常见）
//   - HTTP/2.0: 现代HTTP 2.0协议
//
// 参数：
//   - ctx: Hertz的RequestContext实例
//
// 返回值：
//   - string: 格式化的协议版本字符串（如："HTTP/1.1"）
//
// 注意：
//   - 如果无法确定协议版本，默认返回"HTTP/1.1"
//   - 此实现确保与标准http.Request.Proto字段格式兼容
func extractProtocolVersion(ctx *app.RequestContext) string {
	// 获取Hertz请求的协议信息
	// Hertz通常在连接信息中包含协议版本
	
	// 方法1: 尝试从连接信息获取协议版本
	if conn := ctx.GetConn(); conn != nil {
		// 检查是否为HTTP/2连接
		// Hertz的HTTP/2实现通常会在连接中标识协议版本
		connStr := conn.RemoteAddr().String()
		
		// 这里可以根据Hertz的具体实现进行协议版本检测
		// 由于Hertz的协议检测可能因版本而异，我们采用保守策略
		_ = connStr // 避免未使用变量警告
	}
	
	// 方法2: 从请求头部检查协议版本指示符
	headers := &ctx.Request.Header
	
	// 检查HTTP/2特有的头部
	if headers.Peek("http2-settings") != nil || 
	   string(headers.Peek("upgrade")) == "h2c" {
		return "HTTP/2.0"
	}
	
	// 检查Connection头部的upgrade字段
	if connection := string(headers.Peek("Connection")); connection != "" {
		if strings.Contains(strings.ToLower(connection), "upgrade") {
			if upgrade := string(headers.Peek("Upgrade")); upgrade != "" {
				switch strings.ToLower(upgrade) {
				case "h2c", "http/2.0":
					return "HTTP/2.0"
				}
			}
		}
	}
	
	// 方法3: 检查HTTP版本头部（如果存在）
	if version := string(headers.Peek("HTTP-Version")); version != "" {
		switch version {
		case "2.0", "2":
			return "HTTP/2.0"
		case "1.0":
			return "HTTP/1.0"
		case "1.1":
			return "HTTP/1.1"
		}
	}
	
	// 默认返回HTTP/1.1（最常见的协议版本）
	// 这与大多数现代Web服务器的默认行为一致
	return "HTTP/1.1"
}

// Params 表示路径参数的集合
//
// Params是Param的切片，用于存储从URL路径中解析出的参数。
// 例如：路由"/user/:id/book/:title"会产生两个参数。
type Params []Param

// Param 表示单个路径参数
//
// 包含参数名和参数值。例如：
// 路由"/user/:id"，请求"/user/123"会产生Param{Key: "id", Value: "123"}
type Param struct {
	Key   string // 参数名（如："id"）
	Value string // 参数值（如："123"）
}

// Engine 表示Gin风格的Web框架引擎
//
// Engine是整个Web应用的核心，负责路由注册、中间件管理、
// 请求分发等功能。它包装了Hertz服务器，同时提供Gin兼容的API。
//
// 主要功能：
//   - HTTP路由注册和管理
//   - 全局中间件管理
//   - 静态文件服务
//   - 404/405错误处理
//   - 服务器启动和配置
type Engine struct {
	*server.Hertz               // Hertz服务器实例
	RouterGroup                 // 根路由组，提供路由注册功能
	middleware    []HandlerFunc // 全局中间件（已废弃，使用RouterGroup.handlers）
	noRoute       []HandlerFunc // 404错误处理器
	noMethod      []HandlerFunc // 405错误处理器（方法不允许）

	RedirectTrailingSlash  bool
	RedirectFixedPath      bool
	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool
	AppEngine              bool
	UseRawPath             bool
	UnescapePathValues     bool
	RemoveExtraSlash       bool

	RemoteIPHeaders     []string
	TrustedPlatform     string
	MaxMultipartMemory  int64
	UseH2C              bool
	ContextWithFallback bool
	HTMLRender          render.HTMLRender
	FuncMap             template.FuncMap
	delims              render.Delims
}

// EngineOption 定义Engine的配置选项
// OptionFunc 是一个函数类型，用于配置 Engine 实例。
// 通过传递 OptionFunc 类型的函数，可以在创建或初始化 Engine 时动态地修改其属性或行为。
// 这种模式常用于实现函数式选项（Functional Options）的设计模式。
type OptionFunc func(*Engine)

// HandlersChain 是一组 HandlerFunc 的切片，表示请求处理函数的链（中间件与最终处理器）。
// 按切片顺序依次调用，用于在请求生命周期中管理和执行多个处理函数。
// 此命名与 Gin 保持兼容，并可用于在运行时查询或操作处理链（例如获取最后一个处理器）。
type HandlersChain []HandlerFunc

// Last 返回 HandlersChain 中的最后一个 HandlerFunc。
// 如果 HandlersChain 为空，则返回 nil。
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// RouterGroup 表示路由组
//
// RouterGroup用于组织相关的路由，支持嵌套和中间件。
// 所有路由注册方法都定义在此结构体上。
//
// 特性：
//   - 支持路径前缀
//   - 支持组级中间件
//   - 支持嵌套路由组
//   - 提供所有HTTP方法的路由注册
type RouterGroup struct {
	Handlers []HandlerFunc // 当前组的中间件列表
	basePath string        // 路由组的基础路径前缀
	engine   *Engine       // 所属的引擎实例
	root     bool          // 是否为根路由组
}

// IRoutes 定义路由注册接口
//
// 此接口定义了所有路由注册相关的方法，Engine和RouterGroup都实现了此接口。
// 这样设计可以让路由组和引擎具有相同的路由注册能力。
type IRoutes interface {
	// 中间件管理
	Use(...HandlerFunc) IRoutes // 添加中间件

	// HTTP方法路由
	Handle(string, string, ...HandlerFunc) IRoutes // 注册指定HTTP方法的路由
	Any(string, ...HandlerFunc) IRoutes            // 注册所有HTTP方法的路由
	GET(string, ...HandlerFunc) IRoutes            // 注册GET路由
	POST(string, ...HandlerFunc) IRoutes           // 注册POST路由
	DELETE(string, ...HandlerFunc) IRoutes         // 注册DELETE路由
	PATCH(string, ...HandlerFunc) IRoutes          // 注册PATCH路由
	PUT(string, ...HandlerFunc) IRoutes            // 注册PUT路由
	OPTIONS(string, ...HandlerFunc) IRoutes        // 注册OPTIONS路由
	HEAD(string, ...HandlerFunc) IRoutes           // 注册HEAD路由

	// 静态文件服务
	StaticFile(string, string) IRoutes        // 注册单个静态文件
	Static(string, string) IRoutes            // 注册静态文件目录
	StaticFS(string, http.FileSystem) IRoutes // 注册文件系统
}

// =============================================================================
// Params 方法实现
// =============================================================================

// Get 根据参数名获取参数值
//
// 在参数列表中查找指定名称的参数，返回其值和是否找到的标志。
//
// 参数：
//   - name: 参数名
//
// 返回值：
//   - string: 参数值，如果不存在返回空字符串
//   - bool: 是否找到该参数
func (ps Params) Get(name string) (string, bool) {
	for _, param := range ps {
		if param.Key == name {
			return param.Value, true
		}
	}
	return "", false
}

// ByName 根据参数名获取参数值（简化版本）
//
// 这是Get方法的简化版本，只返回参数值，不返回是否存在的标志。
// 如果参数不存在，返回空字符串。
//
// 参数：
//   - name: 参数名
//
// 返回值：
//   - string: 参数值，如果不存在返回空字符串
func (ps Params) ByName(name string) string {
	value, _ := ps.Get(name)
	return value
}

// =============================================================================
// 便捷类型定义
// =============================================================================

// H 是一个快捷的map[string]any类型，用于构建JSON响应
//
// 这是Gin框架的经典设计，让构建JSON响应更加简洁。
//
// 示例：
//
//	c.JSON(200, gin.H{
//		"message": "success",
//		"data": user,
//		"status": "ok",
//	})
type H map[string]any

// MarshalXML 实现了 xml.Marshaler 接口，用于将 H 类型的值编码为 XML 格式。
//
// 参数:
//
//	e: XML 编码器，用于生成 XML 输出。
//	start: 起始 XML 元素，用于定义 XML 的根元素。
//
// 返回值:
//
//	error: 如果编码过程中发生错误，则返回错误信息；否则返回 nil。
//
// 功能说明:
//  1. 将起始元素的名称设置为 "map"。
//  2. 遍历 H 类型的键值对，将每个键值对编码为 XML 元素。
//  3. 每个键作为 XML 元素的名称，值作为元素的内容。
//  4. 最后关闭起始元素，完成 XML 编码。
//
// 注意:
//
//	此方法适用于导出为 API 的场景，调用者需确保传入的编码器和起始元素有效。
func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}
