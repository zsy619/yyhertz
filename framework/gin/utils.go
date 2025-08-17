// Package gin - 工具函数
// 通用工具函数，包括路径处理、地址解析等辅助功能

package gin

import (
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

const BindKey = "_yyhertz/gin/bindkey"

// WrapH 将http.Handler包装为HandlerFunc
func WrapH(h http.Handler) HandlerFunc {

	return func(c *Context) {
		// 注意：由于Hertz与标准net/http不兼容，这里需要适配
		h.ServeHTTP(c.Writer(), c.Request())
	}
}

// =============================================================================
// 路径处理工具函数
// =============================================================================

// joinPaths 连接路径段
//
// 智能地连接绝对路径和相对路径，处理斜杠和路径规范化。
//
// 参数：
//   - absolutePath: 绝对路径
//   - relativePath: 相对路径
//
// 返回：
//   - string: 连接后的路径
//
// 示例：
//
//	joinPaths("/api", "v1")     // 返回: "/api/v1"
//	joinPaths("/api/", "/v1")   // 返回: "/api/v1"
//	joinPaths("/api", "")       // 返回: "/api"
func joinPaths(absolutePath, relativePath string) string {
	// 如果相对路径为空，直接返回绝对路径
	if relativePath == "" {
		return absolutePath
	}

	// 使用path.Join连接路径
	finalPath := path.Join(absolutePath, relativePath)
	// 如果原相对路径以斜杠结尾，保持这个特征
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

// lastChar 获取字符串的最后一个字符
//
// 返回字符串的最后一个字节，空字符串返回0。
//
// 参数：
//   - str: 输入字符串
//
// 返回：
//   - uint8: 最后一个字符，空字符串返回0
func lastChar(str string) uint8 {
	if str == "" {
		return 0
	}
	return str[len(str)-1]
}

// =============================================================================
// 网络地址处理工具函数
// =============================================================================

// resolveAddress 解析监听地址
//
// 处理Run方法的地址参数，支持默认端口和参数验证。
//
// 参数：
//   - addr: 地址参数列表
//
// 返回：
//   - string: 解析后的监听地址
//
// 规则：
//   - 无参数: 返回":8080"
//   - 一个参数: 返回该参数
//   - 多个参数: panic
//
// 示例：
//
//	resolveAddress([]string{})          // 返回: ":8080"
//	resolveAddress([]string{":3000"})   // 返回: ":3000"
//	resolveAddress([]string{"0.0.0.0:8080"}) // 返回: "0.0.0.0:8080"
func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		return ":8080" // 默认端口
	case 1:
		return addr[0] // 使用指定地址
	default:
		panic("too many parameters") // 不允许多个地址
	}
}

// Bind 是一个用于将请求数据绑定到指定结构体的中间件函数。
// 它接收一个非指针类型的结构体实例作为参数，并返回一个 HandlerFunc。
// 在请求处理过程中，Bind 会尝试将请求数据绑定到传入结构体的新实例上。
// 如果绑定成功，绑定的结果会被存储在 Context 中，可以通过 BindKey 获取。
//
// 参数:
//
//	val: 需要绑定的结构体实例（必须是非指针类型）。
//
// 返回值:
//
//	HandlerFunc: 一个处理函数，用于在请求上下文中执行绑定操作。
//
// 注意事项:
//   - 如果传入的参数是指针类型，函数会触发 panic。
//   - 示例用法: 使用 gin.Bind(Struct{}) 而不是 gin.Bind(&Struct{})。
func Bind(val any) HandlerFunc {
	value := reflect.ValueOf(val)
	if value.Kind() == reflect.Ptr {
		panic(`Bind struct can not be a pointer. Example:
	Use: gin.Bind(Struct{}) instead of gin.Bind(&Struct{})
`)
	}
	typ := value.Type()

	return func(c *Context) {
		obj := reflect.New(typ).Interface()
		if c.Bind(obj) == nil {
			c.Set(BindKey, obj)
		}
	}
}

// isASCII 检查字符串是否仅包含 ASCII 字符。
// 参数：
//
//	s: 待检查的字符串。
//
// 返回值：
//
//	bool: 如果字符串中所有字符均为 ASCII 字符，则返回 true；否则返回 false。
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func nameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// chooseData 根据提供的 custom 和 wildcard 参数选择返回的数据。
// 优先返回 custom 的值，如果 custom 为 nil，则返回 wildcard 的值。
// 如果两者均为 nil，则抛出 panic，提示 "negotiation config is invalid"。
// 该函数通常用于配置协商逻辑中，确保至少有一个有效的配置值可用。
func chooseData(custom, wildcard any) any {
	if custom != nil {
		return custom
	}
	if wildcard != nil {
		return wildcard
	}
	panic("negotiation config is invalid")
}

// parseAccept 解析 HTTP 请求头中的 Accept 字段，返回一个去除了权重参数和空值的 MIME 类型列表。
// 输入参数 acceptHeader 是原始的 Accept 头字符串，例如 "text/html,application/xhtml+xml,application/xml;q=0.9"。
// 返回值是一个字符串切片，包含解析后的 MIME 类型，例如 ["text/html", "application/xhtml+xml", "application/xml"]。
// 注意：此函数会忽略分号后的权重参数（如 q=0.9）和空值。
func parseAccept(acceptHeader string) []string {
	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if i := strings.IndexByte(part, ';'); i > 0 {
			part = part[:i]
		}
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// assert1 是一个断言函数，用于在条件不满足时触发 panic。
// 参数:
//   - guard: 布尔值，如果为 false，则触发 panic。
//   - text: panic 时显示的错误信息。
//
// 注意: 此函数通常用于内部检查，不推荐在公共 API 中使用。
func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

// filterFlags 从给定的字符串中过滤掉标志部分。
// 它会扫描字符串，直到遇到空格或分号为止，并返回截取的部分。
// 如果没有找到空格或分号，则返回原始字符串。
func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
