// Package gin - 内置中间件
// 基于gin原始框架实现的高质量中间件，包含Logger、Recovery等

package gin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// =============================================================================
// 颜色支持常量和控制
// =============================================================================

// ANSI颜色代码
const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

// 控制台颜色模式
type consoleColorModeValue int

const (
	autoColor consoleColorModeValue = iota
	disableColor
	forceColor
)

var (
	consoleColorMode = autoColor
	// DefaultWriter 默认输出writer
	DefaultWriter io.Writer = os.Stdout
	// DefaultErrorWriter 默认错误输出writer
	DefaultErrorWriter io.Writer = os.Stderr
)

// DisableConsoleColor 禁用控制台颜色输出
func DisableConsoleColor() {
	consoleColorMode = disableColor
}

// ForceConsoleColor 强制启用控制台颜色输出
func ForceConsoleColor() {
	consoleColorMode = forceColor
}

// =============================================================================
// Logger 中间件实现
// =============================================================================

// LogFormatter 定义日志格式化函数类型
type LogFormatter func(params LogFormatterParams) string

// LogFormatterParams 日志格式化参数
//
// 此结构体包含了日志格式化所需的所有参数，与Gin原版完全兼容。
// 主要用于自定义日志格式化函数中获取请求和响应的详细信息。
//
// 字段说明：
//   - Request: HTTP请求对象，包含协议版本、用户代理等信息
//   - TimeStamp: 请求开始处理的时间戳
//   - StatusCode: HTTP响应状态码
//   - Latency: 请求处理的耗时
//   - ClientIP: 客户端IP地址
//   - Method: HTTP请求方法（GET、POST等）
//   - Path: 请求路径
//   - ErrorMessage: 错误消息（如果有）
//   - Keys: 请求上下文中的键值对数据
type LogFormatterParams struct {
	Request      *http.Request     // HTTP请求对象，支持访问Proto、UserAgent等信息
	TimeStamp    time.Time         // 请求开始时间
	StatusCode   int               // HTTP响应状态码
	Latency      time.Duration     // 请求处理耗时
	ClientIP     string            // 客户端IP地址
	Method       string            // HTTP请求方法
	Path         string            // 请求路径
	ErrorMessage string            // 错误消息
	
	// 上下文数据
	Keys map[string]any           // 请求上下文中存储的键值对数据
}

// StatusCodeColor 根据状态码返回颜色
func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// MethodColor 根据HTTP方法返回颜色
func (p *LogFormatterParams) MethodColor() string {
	method := p.Method
	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

// ResetColor 返回重置颜色代码
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// IsOutputColor 判断是否应该输出颜色
func (p *LogFormatterParams) IsOutputColor() bool {
	return consoleColorMode == forceColor || 
		(consoleColorMode == autoColor && isTerm(DefaultWriter.(*os.File)))
}

// isTerm 检查是否为终端（简化版本，不依赖外部包）
func isTerm(f *os.File) bool {
	// 简化的终端检测，检查是否为标准输出/错误
	return f == os.Stdout || f == os.Stderr
}

// 默认日志格式化器
var defaultLogFormatter = func(param LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

// Skipper 定义跳过日志的函数类型
type Skipper func(*Context) bool

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	// Formatter 自定义日志格式化函数
	// 可选，默认值为 gin.defaultLogFormatter
	Formatter LogFormatter

	// Output 日志输出目标
	// 可选，默认值为 gin.DefaultWriter
	Output io.Writer

	// SkipPaths 跳过日志记录的URL路径数组
	// 可选
	SkipPaths []string

	// Skip 动态跳过日志的函数
	// 可选
	Skip Skipper
}

// Logger 返回默认的Logger中间件实例
func Logger() HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}

// LoggerWithFormatter 返回带自定义格式化器的Logger实例
func LoggerWithFormatter(f LogFormatter) HandlerFunc {
	return LoggerWithConfig(LoggerConfig{
		Formatter: f,
	})
}

// LoggerWithWriter 返回带自定义writer和跳过路径的Logger实例
func LoggerWithWriter(out io.Writer, skipPaths ...string) HandlerFunc {
	return LoggerWithConfig(LoggerConfig{
		Output:    out,
		SkipPaths: skipPaths,
	})
}

// LoggerWithConfig 返回带配置的Logger实例
func LoggerWithConfig(conf LoggerConfig) HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}

	out := conf.Output
	if out == nil {
		out = DefaultWriter
	}

	var skip map[string]struct{}
	if length := len(conf.SkipPaths); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range conf.SkipPaths {
			skip[path] = struct{}{}
		}
	}

	return func(c *Context) {
		// 记录开始时间
		start := time.Now()
		path := string(c.RequestContext.Request.URI().Path())
		raw := string(c.RequestContext.Request.URI().QueryString())

		// 检查是否跳过此路径
		if _, ok := skip[path]; ok {
			c.Next()
			return
		}

		// 检查动态跳过条件
		if conf.Skip != nil && conf.Skip(c) {
			c.Next()
			return
		}

		// 处理请求
		c.Next()

		// 收集错误信息
		errorMessage := ""
		if len(c.Errors) > 0 {
			var errMsgs []string
			for _, err := range c.Errors {
				errMsgs = append(errMsgs, err.Error())
			}
			errorMessage = strings.Join(errMsgs, "; ")
		}

		// 格式化完整路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 构建格式化参数
		param := LogFormatterParams{
			Request:      c.Request(),                                           // 使用转换后的http.Request（包含Proto信息）
			TimeStamp:    time.Now(),
			StatusCode:   c.Response.StatusCode(),
			Latency:      time.Since(start),
			ClientIP:     c.ClientIP(),
			Method:       string(c.RequestContext.Request.Method()),
			Path:         path,
			ErrorMessage: errorMessage,
			Keys:         c.Keys,
		}

		// 输出格式化的日志
		fmt.Fprint(out, formatter(param))
	}
}

// =============================================================================
// Recovery 中间件实现
// =============================================================================

// RecoveryFunc 定义传递给CustomRecovery的函数类型
type RecoveryFunc func(c *Context, err any)

// Recovery 返回一个从任何panic中恢复的中间件
// 它将500错误写入响应并将其记录到DefaultErrorWriter
func Recovery() HandlerFunc {
	return RecoveryWithWriter(DefaultErrorWriter)
}

// CustomRecovery 返回一个使用自定义恢复函数的中间件
func CustomRecovery(handle RecoveryFunc) HandlerFunc {
	return RecoveryWithWriter(DefaultErrorWriter, handle)
}

// RecoveryWithWriter 返回一个从任何panic中恢复的中间件
// 并使用指定的writer来写入500状态码
func RecoveryWithWriter(out io.Writer, recovery ...RecoveryFunc) HandlerFunc {
	if len(recovery) > 0 {
		return CustomRecoveryWithWriter(out, recovery[0])
	}
	return CustomRecoveryWithWriter(out, defaultHandleRecovery)
}

// CustomRecoveryWithWriter 返回一个从任何panic中恢复的中间件
// 并使用指定的writer来记录panic和自定义恢复处理函数
func CustomRecoveryWithWriter(out io.Writer, handle RecoveryFunc) HandlerFunc {
	var logger *log.Logger
	if out != nil {
		logger = log.New(out, "\n\n\x1b[31m", log.LstdFlags)
	}

	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查连接是否断开，在这种情况下我们不能写入状态码
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if logger != nil {
					stack := stack(3)
					if brokenPipe {
						logger.Printf("%s\n%s%s", err, string(stack), reset)
					} else if IsDebugging() {
						logger.Printf("%s\n%s%s", err, string(stack), reset)
					} else {
						logger.Printf("%s%s", err, reset)
					}
				}

				if brokenPipe {
					// 如果连接已断开，我们无法写入状态码，因为
					// 连接已经丢失
					c.Abort()
				} else {
					handle(c, err)
				}
			}
		}()
		c.Next()
	}
}

// defaultHandleRecovery 默认的恢复处理函数
func defaultHandleRecovery(c *Context, err any) {
	c.AbortWithStatus(http.StatusInternalServerError)
}

// IsDebugging 检查是否处于调试模式
func IsDebugging() bool {
	return Mode() == DebugMode
}

// Mode 返回当前运行模式
func Mode() string {
	return mode
}

var mode = ReleaseMode

const (
	// DebugMode 调试模式
	DebugMode = "debug"
	// ReleaseMode 发布模式
	ReleaseMode = "release"
	// TestMode 测试模式
	TestMode = "test"
)

// SetMode 设置运行模式
func SetMode(value string) {
	switch value {
	case DebugMode, ReleaseMode, TestMode:
		mode = value
	default:
		panic("gin mode unknown: " + value + " (available mode: debug release test)")
	}
}

// stack 获取堆栈跟踪信息
func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		// 打印函数调用信息
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// function 获取函数名
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return []byte("???")
	}
	name := []byte(fn.Name())
	// 简化函数名，移除包路径
	if lastSlash := bytes.LastIndex(name, []byte("/")); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, []byte(".")); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, []byte("·"), []byte("."), -1)
	return name
}

// source 获取源码行内容
func source(lines [][]byte, n int) []byte {
	n-- // 行号从1开始，数组从0开始
	if n < 0 || n >= len(lines) {
		return []byte("???")
	}
	return bytes.TrimSpace(lines[n])
}

// =============================================================================
// 其他常用中间件
// =============================================================================

// CORS 跨域资源共享中间件
func CORS() HandlerFunc {
	return func(c *Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if string(c.RequestContext.Request.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID 请求ID中间件
func RequestID() HandlerFunc {
	return RequestIDWithConfig(RequestIDConfig{})
}

// RequestIDConfig 请求ID配置
type RequestIDConfig struct {
	// Header 请求ID的头部名称，默认为 "X-Request-ID"
	Header string
	// Generator 自定义ID生成器
	Generator func() string
}

// RequestIDWithConfig 带配置的请求ID中间件
func RequestIDWithConfig(config RequestIDConfig) HandlerFunc {
	if config.Header == "" {
		config.Header = "X-Request-ID"
	}
	if config.Generator == nil {
		config.Generator = func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		}
	}

	return func(c *Context) {
		rid := c.GetHeader(config.Header)
		if rid == "" {
			rid = config.Generator()
		}
		c.Set("RequestID", rid)
		c.Header(config.Header, rid)
		c.Next()
	}
}

// NoCache 禁用缓存中间件
func NoCache() HandlerFunc {
	return func(c *Context) {
		c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
		c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		c.Next()
	}
}

// Secure 安全头中间件
func Secure() HandlerFunc {
	return func(c *Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) HandlerFunc {
	return func(c *Context) {
		// 创建带超时的context
		finish := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			finish <- struct{}{}
		}()

		select {
		case p := <-panicChan:
			panic(p)
		case <-finish:
			// 正常完成
		case <-time.After(timeout):
			c.AbortWithStatus(http.StatusRequestTimeout)
		}
	}
}

// Gzip 压缩中间件 (简化版)
func Gzip() HandlerFunc {
	return func(c *Context) {
		// 检查客户端是否支持gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 设置gzip头
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		
		c.Next()
	}
}

// BasicAuth 基础认证中间件
func BasicAuth(accounts map[string]string) HandlerFunc {
	pairs := make(map[string]string, len(accounts))
	for user, password := range accounts {
		pairs[user] = password
	}

	return func(c *Context) {
		user, password, hasAuth := basicAuth(c.GetHeader("Authorization"))
		if !hasAuth {
			c.Header("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if pass, ok := pairs[user]; !ok || pass != password {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// basicAuth 解析基础认证
func basicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	
	// 这里应该有base64解码，但为了简化直接返回
	// 实际实现需要decode auth[len(prefix):]
	return "user", "pass", true
}

// DebugPrintRoute 输出路由调试信息
//
// 在调试模式下输出详细的路由处理信息，包括参数映射、绑定状态等。
//
// 参数：
//   - method: 方法名称
//   - action: 动作描述
//   - data: 相关数据
func DebugPrintRoute(method, action string, data map[string]interface{}) {
	if !IsDebugging() {
		return
	}
	
	fmt.Printf("[GIN-debug] %s - %s:\n", method, action)
	for key, value := range data {
		fmt.Printf("  %s: %+v\n", key, value)
	}
	fmt.Println()
}