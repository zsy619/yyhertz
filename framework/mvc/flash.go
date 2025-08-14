// Package mvc 提供MVC框架的Flash消息功能
//
// Flash消息是一种用于在重定向间传递临时消息的机制，常用于：
// - 用户操作反馈（成功、失败、警告等）
// - 表单验证错误信息
// - 状态提示信息
// - 页面跳转时的消息传递
//
// 工作原理：
// Flash消息通过Cookie机制实现，消息在被读取后会自动删除，
// 确保消息只显示一次，避免用户刷新页面时重复显示。
//
// 特性：
// - 自动过期：消息读取后自动删除
// - 多种类型：支持success、error、warning、notice等类型
// - URL安全：消息内容经过URL编码处理
// - 简单易用：提供便捷的设置和读取方法
//
// 使用示例：
//
//	// 在控制器中设置Flash消息
//	flash := mvc.NewFlash()
//	flash.Success("用户创建成功")
//	flash.Store(c)
//	c.Redirect("/users")
//
//	// 在目标页面读取Flash消息
//	flash := mvc.ReadFromRequest(c)
//	if msg, ok := flash.Data["success"]; ok {
//		// 显示成功消息
//	}
package mvc

import (
	"fmt"
	"net/url"
	"strings"
)

// FlashSeparator Flash消息键值对分隔符
//
// 用于分隔Flash消息中的键和值，使用特殊字符避免与消息内容冲突。
const FlashSeparator = "#"

// FlashName Flash消息在Cookie中的名称
//
// 所有Flash消息都存储在这个名为"yyhertz_flash"的Cookie中。
const FlashName = "yyhertz_flash"

// FlashData Flash消息数据结构
//
// 用于存储和管理Flash消息，支持多种消息类型的键值对存储。
// 消息在被读取后会自动从Cookie中删除。
//
// 字段说明：
//   - Data: map[string]string - 存储消息的键值对，键为消息类型，值为消息内容
//
// 使用示例：
//
//	flash := mvc.NewFlash()
//	flash.Success("操作成功")
//	flash.Error("操作失败")
//	flash.Warning("请注意")
type FlashData struct {
	// Data 存储Flash消息的键值对映射
	//
	// 键通常是消息类型（如"success"、"error"、"warning"、"notice"），
	// 值是具体的消息内容。支持自定义键名来存储不同类型的消息。
	Data map[string]string
}

// NewFlash 创建新的FlashData实例
//
// 初始化一个空的Flash消息容器，可以用来存储各种类型的临时消息。
//
// 返回值：
//   - *FlashData: 新创建的Flash消息实例
//
// 使用示例：
//
//	flash := mvc.NewFlash()
//	flash.Success("用户注册成功")
//	flash.Store(controller)
func NewFlash() *FlashData {
	return &FlashData{
		Data: make(map[string]string),
	}
}
// Set 设置自定义键的Flash消息
//
// 允许使用自定义键名设置Flash消息，支持格式化字符串。
// 这个方法提供了最大的灵活性，可以定义任意类型的消息。
//
// 参数：
//   - key: string - 消息的键名，如"custom"、"validation"等
//   - msg: string - 消息内容，支持格式化占位符
//   - args: ...any - 格式化参数（可选）
//
// 使用示例：
//
//	// 简单消息
//	flash.Set("custom", "自定义消息")
//
//	// 格式化消息
//	flash.Set("validation", "字段 %s 不能为空", "username")
//	flash.Set("info", "共找到 %d 条记录", count)
func (fd *FlashData) Set(key string, msg string, args ...any) {
	if len(args) == 0 {
		fd.Data[key] = msg
	} else {
		fd.Data[key] = fmt.Sprintf(msg, args...)
	}
}

// Success 设置成功类型的Flash消息
//
// 用于设置操作成功的提示消息，通常显示为绿色或带有成功图标。
// 支持格式化字符串以便插入动态内容。
//
// 参数：
//   - msg: string - 成功消息内容，支持格式化占位符
//   - args: ...any - 格式化参数（可选）
//
// 使用示例：
//
//	// 简单成功消息
//	flash.Success("操作成功")
//
//	// 格式化成功消息
//	flash.Success("用户 %s 创建成功", username)
//	flash.Success("已删除 %d 条记录", count)
//
// 常用场景：
//   - 用户注册成功
//   - 数据保存成功
//   - 文件上传成功
//   - 操作完成提示
func (fd *FlashData) Success(msg string, args ...any) {
	if len(args) == 0 {
		fd.Data["success"] = msg
	} else {
		fd.Data["success"] = fmt.Sprintf(msg, args...)
	}
}

// Notice 设置通知类型的Flash消息
//
// 用于设置一般性通知消息，通常显示为蓝色或带有信息图标。
// 适用于向用户传达重要但非紧急的信息。
//
// 参数：
//   - msg: string - 通知消息内容，支持格式化占位符
//   - args: ...any - 格式化参数（可选）
//
// 使用示例：
//
//	// 简单通知消息
//	flash.Notice("系统将在30分钟后维护")
//
//	// 格式化通知消息
//	flash.Notice("您有 %d 条新消息", messageCount)
//	flash.Notice("欢迎回来，%s", username)
//
// 常用场景：
//   - 系统通知
//   - 功能说明
//   - 状态变更提醒
//   - 新功能介绍
func (fd *FlashData) Notice(msg string, args ...any) {
	if len(args) == 0 {
		fd.Data["notice"] = msg
	} else {
		fd.Data["notice"] = fmt.Sprintf(msg, args...)
	}
}

// Warning 设置警告类型的Flash消息
//
// 用于设置警告性质的消息，通常显示为橙色或带有警告图标。
// 适用于提醒用户注意但不影响正常操作的情况。
//
// 参数：
//   - msg: string - 警告消息内容，支持格式化占位符
//   - args: ...any - 格式化参数（可选）
//
// 使用示例：
//
//	// 简单警告消息
//	flash.Warning("密码即将过期")
//
//	// 格式化警告消息
//	flash.Warning("您的账户余额不足，当前余额：%.2f元", balance)
//	flash.Warning("文件 %s 已存在，将被覆盖", filename)
//
// 常用场景：
//   - 权限不足警告
//   - 资源不足提醒
//   - 安全风险提示
//   - 数据冲突警告
func (fd *FlashData) Warning(msg string, args ...any) {
	if len(args) == 0 {
		fd.Data["warning"] = msg
	} else {
		fd.Data["warning"] = fmt.Sprintf(msg, args...)
	}
}

// Error 设置错误类型的Flash消息
//
// 用于设置错误消息，通常显示为红色或带有错误图标。
// 适用于操作失败、验证错误等需要用户注意的错误情况。
//
// 参数：
//   - msg: string - 错误消息内容，支持格式化占位符
//   - args: ...any - 格式化参数（可选）
//
// 使用示例：
//
//	// 简单错误消息
//	flash.Error("用户名或密码错误")
//
//	// 格式化错误消息
//	flash.Error("文件 %s 上传失败", filename)
//	flash.Error("服务器错误：%v", err)
//
// 常用场景：
//   - 登录失败
//   - 表单验证错误
//   - 文件操作失败
//   - 数据库操作错误
func (fd *FlashData) Error(msg string, args ...any) {
	if len(args) == 0 {
		fd.Data["error"] = msg
	} else {
		fd.Data["error"] = fmt.Sprintf(msg, args...)
	}
}

// Store 将Flash消息存储到Cookie中
//
// 将当前FlashData中的所有消息编码后存储到Cookie中，同时将消息数据
// 添加到控制器的Data中以便在当前请求中也能访问。
//
// 参数：
//   - c: *BaseController - 控制器实例，用于设置Cookie和Data
//
// 存储机制：
//   1. 将消息数据设置到控制器的Data["flash"]中
//   2. 将消息编码为特殊格式的字符串
//   3. 对编码后的字符串进行URL编码
//   4. 存储到名为FlashName的Cookie中
//
// 编码格式：
//   每个键值对使用特殊字符分隔：\x00 + key + \x23# + value + \x00
//
// 使用示例：
//
//	flash := mvc.NewFlash()
//	flash.Success("操作成功")
//	flash.Store(c) // 存储到Cookie
//	c.Redirect("/success") // 重定向到目标页面
//
// 注意事项：
//   - Cookie有大小限制（通常4KB），避免存储过多消息
//   - 消息会在下次请求时被自动读取并删除
//   - 支持在同一请求中访问刚设置的消息（通过c.Data["flash"]）
func (fd *FlashData) Store(c *BaseController) {
	c.Data["flash"] = fd.Data
	var flashValue string
	for key, value := range fd.Data {
		flashValue += "\x00" + key + "\x23" + FlashSeparator + "\x23" + value + "\x00"
	}
	c.Ctx.SetCookie(FlashName, url.QueryEscape(flashValue), 0, "/")
}

// ReadFromRequest 从请求中读取Flash消息
//
// 从Cookie中读取Flash消息并解析为FlashData结构。读取后会立即删除Cookie，
// 确保Flash消息只显示一次。同时将解析的消息数据设置到控制器的Data中。
//
// 参数：
//   - c: *BaseController - 控制器实例，用于读取Cookie和设置Data
//
// 返回值：
//   - *FlashData: 包含所有Flash消息的数据结构
//
// 工作流程：
//   1. 创建新的FlashData实例
//   2. 尝试从Cookie中读取Flash消息
//   3. 如果存在，则进行URL解码和字符串解析
//   4. 解析出键值对并存储到FlashData中
//   5. 立即删除Cookie（设置过期时间为-1）
//   6. 将消息数据设置到控制器的Data["flash"]中
//   7. 返回包含消息的FlashData实例
//
// 使用示例：
//
//	// 在目标页面的控制器中读取Flash消息
//	flash := mvc.ReadFromRequest(c)
//
//	// 检查并使用不同类型的消息
//	if successMsg, ok := flash.Data["success"]; ok {
//		// 处理成功消息
//		c.Data["SuccessMessage"] = successMsg
//	}
//
//	if errorMsg, ok := flash.Data["error"]; ok {
//		// 处理错误消息
//		c.Data["ErrorMessage"] = errorMsg
//	}
//
//	// 或者直接在模板中使用
//	// 模板中可以通过 {{.flash.success}} 访问消息
//
// 特性：
//   - 自动清理：读取后Cookie自动删除
//   - 安全解码：支持URL编码的消息内容
//   - 容错处理：Cookie不存在或格式错误时返回空的FlashData
//   - 模板友好：自动设置到控制器Data中供模板使用
//
// 注意事项：
//   - 每个Flash消息只能读取一次
//   - 如果页面刷新，之前的Flash消息将不再可用
//   - 建议在页面加载时立即读取和处理Flash消息
func ReadFromRequest(c *BaseController) *FlashData {
	flash := NewFlash()
	if cookie, err := c.Ctx.Cookie(FlashName); err == nil {
		v, _ := url.QueryUnescape(cookie)
		vals := strings.Split(v, "\x00")
		for _, v := range vals {
			if len(v) > 0 {
				kv := strings.Split(v, "\x23"+FlashSeparator+"\x23")
				if len(kv) == 2 {
					flash.Data[kv[0]] = kv[1]
				}
			}
		}
		// read one time then delete it
		c.Ctx.SetCookie(FlashName, "", -1, "/")
	}
	c.Data["flash"] = flash.Data
	return flash
}
