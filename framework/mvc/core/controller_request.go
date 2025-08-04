package core

import (
	"net"
	"strconv"
	"strings"
)

// ============= 参数获取方法 =============

// GetString 获取字符串参数
func (c *BaseController) GetString(key string, def ...string) string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
	
	// 直接从Hertz RequestContext获取查询参数
	if c.Ctx.RequestContext != nil {
		if pathBytes := c.Ctx.RequestContext.QueryArgs().Peek(key); pathBytes != nil {
			return string(pathBytes)
		}
	}
	
	// 备用方法
	if val := c.Ctx.Query(key); val != "" {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetInt 获取整数参数
func (c *BaseController) GetInt(key string, def ...int) int {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetParam 获取路径参数
func (c *BaseController) GetParam(key string) string {
	if c.Ctx == nil {
		return ""
	}
	return c.Ctx.Param(key)
}

// GetForm 获取表单参数
func (c *BaseController) GetForm(key string, def ...string) string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
	if val := c.Ctx.PostForm(key); val != "" {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetBool 获取布尔参数（Beego兼容）
func (c *BaseController) GetBool(key string, def ...bool) bool {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return false
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0]
		}
		return false
	}

	// 转换字符串到布尔值
	switch strings.ToLower(val) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		if len(def) > 0 {
			return def[0]
		}
		return false
	}
}

// GetFloat 获取浮点数参数（Beego兼容）
func (c *BaseController) GetFloat(key string, def ...float64) float64 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0.0
	}

	if val := c.Ctx.Query(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}

	if len(def) > 0 {
		return def[0]
	}
	return 0.0
}

// GetQuery 获取查询参数（Beego兼容）
func (c *BaseController) GetQuery(key string, def ...string) string {
	return c.GetString(key, def...)
}

// GetUserAgent 获取User-Agent
func (c *BaseController) GetUserAgent() string {
	if c.Ctx == nil {
		return ""
	}
	return string(c.Ctx.RequestContext.GetHeader("User-Agent"))
}

// GetHeader 获取请求头
func (c *BaseController) GetHeader(key string) string {
	if c.Ctx == nil {
		return ""
	}
	return string(c.Ctx.RequestContext.GetHeader(key))
}

// GetClientIP 获取客户端IP地址
func (c *BaseController) GetClientIP() string {
	if c.Ctx == nil {
		return ""
	}
	// 尝试从X-Forwarded-For获取真实IP
	if xff := c.Ctx.RequestContext.GetHeader("X-Forwarded-For"); len(xff) > 0 {
		xffStr := string(xff)
		ips := strings.Split(xffStr, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 尝试从X-Real-IP获取
	if xri := c.Ctx.RequestContext.GetHeader("X-Real-IP"); len(xri) > 0 {
		return string(xri)
	}

	// 从RemoteAddr获取
	remoteAddr := c.Ctx.RequestContext.RemoteAddr().String()
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// ============= HTTP方法判断 =============

// IsAjax 判断是否为AJAX请求
func (c *BaseController) IsAjax() bool {
	if c.Ctx == nil {
		return false
	}
	return string(c.Ctx.RequestContext.GetHeader("X-Requested-With")) == "XMLHttpRequest"
}

// IsMethod 判断HTTP方法
func (c *BaseController) IsMethod(method string) bool {
	if c.Ctx == nil {
		return false
	}
	return strings.ToUpper(string(c.Ctx.RequestContext.Method())) == strings.ToUpper(method)
}

// IsPost 判断是否为POST请求
func (c *BaseController) IsPost() bool {
	return c.IsMethod("POST")
}

// IsGet 判断是否为GET请求
func (c *BaseController) IsGet() bool {
	return c.IsMethod("GET")
}

// IsPut 判断是否为PUT请求
func (c *BaseController) IsPut() bool {
	return c.IsMethod("PUT")
}

// IsDelete 判断是否为DELETE请求
func (c *BaseController) IsDelete() bool {
	return c.IsMethod("DELETE")
}

// IsPatch 判断是否为PATCH请求
func (c *BaseController) IsPatch() bool {
	return c.IsMethod("PATCH")
}

// IsHead 判断是否为HEAD请求
func (c *BaseController) IsHead() bool {
	return c.IsMethod("HEAD")
}

// IsOptions 判断是否为OPTIONS请求
func (c *BaseController) IsOptions() bool {
	return c.IsMethod("OPTIONS")
}