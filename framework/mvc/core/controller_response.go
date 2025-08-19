package core

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gopkg.in/yaml.v3"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/response"
)

// ============= JSON响应方法 =============

// JSON 返回JSON格式的数据
func (c *BaseController) JSON(data any) {
	c.JSONWithStatus(consts.StatusOK, data)
}

// JSONWithStatus 返回指定状态码的JSON数据
func (c *BaseController) JSONWithStatus(status int, data any) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return JSON")
		return
	}
	c.Ctx.JSON(status, data)
}

// JSONSuccess 返回成功的JSON响应
func (c *BaseController) JSONSuccess(message string, data any) {
	c.JSON(response.Success(message, data))
}

// JSONError 返回错误的JSON响应
func (c *BaseController) JSONError(message string) {
	c.JSON(response.Error(message))
}

// JSONPage 返回分页JSON响应
func (c *BaseController) JSONPage(message string, data any, count int64) {
	c.JSON(response.SuccessPage(message, data, count))
}

// JSONOK 返回成功响应（200）
func (c *BaseController) JSONOK(message string, data any) {
	c.JSONStatus(http.StatusOK, 0, message, data)
}

// JSONStatus 返回指定状态码的JSON响应
func (c *BaseController) JSONStatus(status int, code int, message string, data any) {
	response := map[string]any{
		"code":    code,
		"message": message,
		"data":    data,
	}
	c.JSONWithStatus(status, response)
}

// ============= 字符串响应方法 =============

// String 返回字符串响应
func (c *BaseController) String(s string) {
	c.StringWithStatus(consts.StatusOK, s)
}

// StringWithStatus 返回指定状态码的字符串响应
func (c *BaseController) StringWithStatus(status int, s string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return string")
		return
	}
	c.Ctx.String(status, "%s", s)
}

// ============= 重定向和错误方法 =============

// Redirect 重定向
func (c *BaseController) Redirect(url string, code ...int) {
	statusCode := consts.StatusFound
	if len(code) > 0 {
		statusCode = code[0]
	}
	c.Ctx.Request().Redirect(statusCode, []byte(url))
}

// Error 返回错误响应
func (c *BaseController) Error(code int, msg string) {
	c.Ctx.String(code, "%s", msg)
}

// ============= 响应头和原始数据方法 =============

// SetHeader 设置响应头
func (c *BaseController) SetHeader(key, value string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to set header")
		return
	}
	c.Ctx.SetHeader(key, value)
}

// Write 写入原始字节数据
func (c *BaseController) Write(data []byte) (int, error) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to write data")
		return 0, fmt.Errorf("context is nil")
	}
	return c.Ctx.Write(data)
}

// ============= Beego兼容的ServeJSON方法 =============

// ServeJSON 发送JSON响应，完全兼容beego的ServeJSON方法
//
// 该方法从 c.Data["json"] 中读取数据并序列化为JSON响应。
// 根据运行模式自动选择是否格式化JSON输出：
// - 开发模式：格式化输出（有缩进）
// - 生产模式：紧凑输出（无缩进）
//
// 参数:
//
//	encoding: 可选参数，如果为true，启用UTF-8字符转义为\uXXXX格式
//
// 使用方法:
//
//	c.Data["json"] = map[string]any{"status": "success", "data": userData}
//	c.ServeJSON()
//
//	// 或带编码转换
//	c.Data["json"] = responseData
//	c.ServeJSON(true)
//
// 注意:
//   - 如果 c.Data["json"] 不存在，会返回 HTTP 500 错误
//   - JSON序列化失败也会返回 HTTP 500 错误
//   - Content-Type 自动设置为 "application/json; charset=utf-8"
func (c *BaseController) ServeJSON(encoding ...bool) error {
	// 检查上下文是否有效
	if c.Ctx == nil {
		config.Error("Context is nil when trying to serve JSON")
		return fmt.Errorf("context is nil")
	}

	// 检查是否有JSON数据
	jsonData, exists := c.Data["json"]
	if !exists {
		config.Error("No JSON data found in c.Data[\"json\"]")
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString(`{"error": "No JSON data provided"}`)
		return fmt.Errorf("no JSON data provided")
	}

	// 确定是否需要格式化输出（基于运行模式）
	hasIndent := c.shouldIndentJSON()

	// 确定是否需要UTF-8编码转换
	needsEncoding := len(encoding) > 0 && encoding[0]

	// 序列化JSON数据
	var jsonBytes []byte
	var err error

	if hasIndent {
		jsonBytes, err = json.MarshalIndent(jsonData, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(jsonData)
	}

	if err != nil {
		config.Error("Failed to marshal JSON data:", err)
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString(`{"error": "Failed to serialize JSON data"}`)
		return fmt.Errorf("failed to serialize JSON data")
	}

	// 如果需要UTF-8编码转换
	if needsEncoding {
		jsonBytes = c.encodeUTF8ToUnicode(jsonBytes)
	}

	// 设置响应头
	c.Ctx.SetContentType("application/json; charset=utf-8")

	// 写入响应
	_, errx := c.Ctx.Write(jsonBytes)
	return errx
}

// shouldIndentJSON 判断是否应该格式化JSON输出
// 在开发模式下返回true（格式化输出），生产模式下返回false（紧凑输出）
func (c *BaseController) shouldIndentJSON() bool {
	return c.shouldIndentOutput()
}

// shouldIndentOutput 判断是否应该格式化输出
// 在开发模式下返回true（格式化输出），生产模式下返回false（紧凑输出）
// 适用于JSON、XML等格式的缩进判断
func (c *BaseController) shouldIndentOutput() bool {
	// 优先检查测试环境变量（用于测试）
	if testMode := os.Getenv("APP_ENV"); testMode != "" {
		return testMode == "dev" || testMode == "development" || testMode == "debug"
	}

	// 然后检查配置文件中的环境设置
	cnf, _ := config.GetAppConfig()
	runMode := cnf.App.Environment
	if runMode == "" {
		runMode = os.Getenv("GO_ENV")
	}
	if runMode == "" {
		// 默认为开发模式
		runMode = "development"
	}

	// 开发模式下格式化输出，便于调试
	return runMode == "dev" || runMode == "development" || runMode == "debug"
}

// encodeUTF8ToUnicode 将UTF-8字符转换为Unicode转义序列（\uXXXX格式）
// 这个功能与beego保持兼容，只处理JSON字符串值中的非ASCII字符
func (c *BaseController) encodeUTF8ToUnicode(data []byte) []byte {
	var buf strings.Builder
	buf.Grow(len(data) * 2) // 预分配足够的空间

	s := string(data)
	inString := false
	escaped := false

	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)

		if r == utf8.RuneError {
			// 如果是无效的UTF-8序列，保持原样
			if err := buf.WriteByte(s[0]); err != nil {
				panic(err)
			}
			s = s[1:]
			continue
		}

		// 处理转义状态
		if escaped {
			if _, err := buf.WriteRune(r); err != nil {
				config.Error(err)
			}
			escaped = false
			s = s[size:]
			continue
		}

		// 检查是否在字符串内
		if r == '"' && !escaped {
			inString = !inString
			buf.WriteRune(r)
			s = s[size:]
			continue
		}

		if r == '\\' && inString {
			escaped = true
			buf.WriteRune(r)
			s = s[size:]
			continue
		}

		// 只在JSON字符串值内进行Unicode编码
		if inString && r > 127 {
			// 非ASCII字符转换为\uXXXX格式
			if r <= 0xFFFF {
				if _, err := buf.WriteString(fmt.Sprintf("\\u%04X", r)); err != nil {
					config.Error(err)
				}
			} else {
				// 对于超过0xFFFF的字符，使用UTF-16代理对
				r -= 0x10000
				high := 0xD800 + (r >> 10)
				low := 0xDC00 + (r & 0x3FF)
				buf.WriteString(fmt.Sprintf("\\u%04X\\u%04X", high, low))
			}
		} else {
			// ASCII字符或非字符串部分保持原样
			if _, err := buf.WriteRune(r); err != nil {
				config.Error(err)
			}
		}

		s = s[size:]
	}

	return []byte(buf.String())
}

// ============= Beego兼容的Serve方法 =============

// ServeXML 发送XML响应，完全兼容beego的ServeXML方法
//
// 该方法从 c.Data["xml"] 中读取数据并序列化为XML响应。
// 根据运行模式自动选择是否格式化XML输出：
// - 开发模式：格式化输出（有缩进）
// - 生产模式：紧凑输出（无缩进）
//
// 使用方法:
//
//	c.Data["xml"] = struct{Name string `xml:"name"`}{Name: "example"}
//	c.ServeXML()
//
// 注意:
//   - 如果 c.Data["xml"] 不存在，会返回 HTTP 500 错误
//   - XML序列化失败也会返回 HTTP 500 错误
//   - Content-Type 自动设置为 "application/xml; charset=utf-8"
func (c *BaseController) ServeXML() {
	// 检查上下文是否有效
	if c.Ctx == nil {
		config.Error("Context is nil when trying to serve XML")
		return
	}

	// 检查是否有XML数据
	xmlData, exists := c.Data["xml"]
	if !exists {
		config.Error("No XML data found in c.Data[\"xml\"]")
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString(`<?xml version="1.0" encoding="UTF-8"?><error>No XML data provided</error>`)
		return
	}

	// 确定是否需要格式化输出（基于运行模式）
	hasIndent := c.shouldIndentOutput()

	// 序列化XML数据
	var xmlBytes []byte
	var err error

	if hasIndent {
		xmlBytes, err = xml.MarshalIndent(xmlData, "", "  ")
	} else {
		xmlBytes, err = xml.Marshal(xmlData)
	}

	if err != nil {
		config.Error("Failed to marshal XML data:", err)
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString(`<?xml version="1.0" encoding="UTF-8"?><error>Failed to serialize XML data</error>`)
		return
	}

	// 设置响应头
	c.Ctx.SetContentType("application/xml; charset=utf-8")

	// 写入响应
	_, err = c.Ctx.Write(xmlBytes)
	if err != nil {
		config.Error("Failed to write XML response:", err)
	}
}

// ServeJSONP 发送JSONP响应，完全兼容beego的ServeJSONP方法
//
// 该方法从 c.Data["jsonp"] 中读取数据并序列化为JSONP响应。
// 根据运行模式自动选择是否格式化JSON输出：
// - 开发模式：格式化输出（有缩进）
// - 生产模式：紧凑输出（无缩进）
//
// JSONP回调函数名从query参数"callback"中获取，默认为"callback"
//
// 使用方法:
//
//	c.Data["jsonp"] = map[string]any{"status": "success", "data": userData}
//	c.ServeJSONP()
//
// 注意:
//   - 如果 c.Data["jsonp"] 不存在，会返回 HTTP 500 错误
//   - JSON序列化失败也会返回 HTTP 500 错误
//   - Content-Type 自动设置为 "application/javascript; charset=utf-8"
//   - 回调函数名会自动从URL参数callback中获取
func (c *BaseController) ServeJSONP() error {
	// 检查上下文是否有效
	if c.Ctx == nil {
		config.Error("Context is nil when trying to serve JSONP")
		return c.Errorf("context is nil")
	}

	// 检查是否有JSONP数据
	jsonpData, exists := c.Data["jsonp"]
	if !exists {
		config.Error("No JSONP data found in c.Data[\"jsonp\"]")
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString(`callback({"error": "No JSONP data provided"});`)
		return c.Errorf("no JSONP data provided")
	}

	// 获取回调函数名
	callback := c.Ctx.Query("callback")
	if callback == "" {
		callback = "callback" // 默认回调函数名
	}

	// 确定是否需要格式化输出（基于运行模式）
	hasIndent := c.shouldIndentOutput()

	// 序列化JSON数据
	var jsonBytes []byte
	var err error

	if hasIndent {
		jsonBytes, err = json.MarshalIndent(jsonpData, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(jsonpData)
	}

	if err != nil {
		config.Error("Failed to marshal JSONP data:", err)
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString(callback + `({"error": "Failed to serialize JSONP data"});`)
		return c.Errorf("failed to serialize JSONP data")
	}

	// 构建JSONP响应
	jsonpResponse := callback + "(" + string(jsonBytes) + ");"

	// 设置响应头
	c.Ctx.SetContentType("application/javascript; charset=utf-8")

	// 写入响应
	c.Ctx.WriteString(jsonpResponse)
	return nil
}

// ServeYAML 发送YAML响应，完全兼容beego的ServeYAML方法
//
// 该方法从 c.Data["yaml"] 中读取数据并序列化为YAML响应。
// YAML格式天然带缩进，无需额外的格式化处理。
//
// 使用方法:
//
//	c.Data["yaml"] = map[string]any{"status": "success", "data": userData}
//	c.ServeYAML()
//
// 注意:
//   - 如果 c.Data["yaml"] 不存在，会返回 HTTP 500 错误
//   - YAML序列化失败也会返回 HTTP 500 错误
//   - Content-Type 自动设置为 "application/yaml; charset=utf-8"
//   - YAML格式天然支持多行和缩进，便于阅读
func (c *BaseController) ServeYAML() error {
	// 检查上下文是否有效
	if c.Ctx == nil {
		config.Error("Context is nil when trying to serve YAML")
		return fmt.Errorf("context is nil")
	}

	// 检查是否有YAML数据
	yamlData, exists := c.Data["yaml"]
	if !exists {
		config.Error("No YAML data found in c.Data[\"yaml\"]")
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString("error: No YAML data provided\n")
		return fmt.Errorf("no YAML data provided")
	}

	// 序列化YAML数据
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		config.Error("Failed to marshal YAML data:", err)
		c.Ctx.Status(consts.StatusInternalServerError)
		c.Ctx.WriteString("error: Failed to serialize YAML data\n")
		return fmt.Errorf("failed to serialize YAML data")
	}

	// 设置响应头
	c.Ctx.SetContentType("application/yaml; charset=utf-8")

	// 写入响应
	_, err = c.Ctx.Write(yamlBytes)
	return err
}

// ServeFormatted 根据Accept header自动选择格式，完全兼容beego的ServeFormatted方法
//
// 该方法根据HTTP Accept header自动选择最合适的响应格式。
// 支持的格式优先级：JSON -> XML -> YAML
// 如果没有匹配的Accept header，默认返回JSON格式。
//
// 参数:
//
//	encoding: 可选参数，如果为true，启用UTF-8字符转义为\uXXXX格式（仅对JSON有效）
//
// 数据来源:
//   - JSON格式：从 c.Data["json"] 读取
//   - XML格式：从 c.Data["xml"] 读取
//   - YAML格式：从 c.Data["yaml"] 读取
//
// Accept header支持:
//   - application/json, text/json -> JSON
//   - application/xml, text/xml -> XML
//   - application/yaml, text/yaml -> YAML
//   - */* -> 默认JSON
//
// 使用方法:
//
//	c.Data["json"] = jsonData
//	c.Data["xml"] = xmlData
//	c.Data["yaml"] = yamlData
//	c.ServeFormatted()        // 不启用UTF-8编码
//	c.ServeFormatted(true)    // 启用UTF-8编码（仅JSON）
//
// 注意:
//   - 如果没有对应格式的数据，会返回 HTTP 500 错误
//   - 如果Accept header不匹配任何支持的格式，默认使用JSON
//   - UTF-8编码选项仅对JSON格式有效
func (c *BaseController) ServeFormatted(encoding ...bool) {
	// 检查上下文是否有效
	if c.Ctx == nil {
		config.Error("Context is nil when trying to serve formatted response")
		return
	}

	// 获取Accept header
	accept := c.Ctx.GetHeader("Accept")
	if accept == "" {
		accept = "application/json" // 默认为JSON
	}

	// 根据Accept header选择格式
	switch {
	case strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml"):
		// XML格式
		if _, exists := c.Data["xml"]; exists {
			c.ServeXML()
		} else {
			config.Error("No XML data found in c.Data[\"xml\"] for ServeFormatted")
			c.Ctx.Status(consts.StatusInternalServerError)
			c.Ctx.WriteString(`<?xml version="1.0" encoding="UTF-8"?><error>No XML data provided</error>`)
		}

	case strings.Contains(accept, "application/yaml") || strings.Contains(accept, "text/yaml"):
		// YAML格式
		if _, exists := c.Data["yaml"]; exists {
			c.ServeYAML()
		} else {
			config.Error("No YAML data found in c.Data[\"yaml\"] for ServeFormatted")
			c.Ctx.Status(consts.StatusInternalServerError)
			c.Ctx.WriteString("error: No YAML data provided\n")
		}

	default:
		// 默认JSON格式（包括application/json, text/json, */*等）
		if _, exists := c.Data["json"]; exists {
			err := c.ServeJSON(encoding...)
			if err != nil {
				config.Error("Failed to serve JSON response:", err)
			}
		} else {
			config.Error("No JSON data found in c.Data[\"json\"] for ServeFormatted")
			c.Ctx.Status(consts.StatusInternalServerError)
			c.Ctx.WriteString(`{"error": "No JSON data provided"}`)
		}
	}
}
