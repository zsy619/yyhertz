package core

// 实现控制器的格式化输出，完全兼容fmt的输出
import (
	"fmt"
	"io"
)

// ============= Print系列方法 - 直接调用标准fmt包 =============

// Print 使用默认格式格式化其操作数并写入到标准输出。
// 当两个连续的操作数都不是字符串时，会在它们之间添加空格。
// 返回写入的字节数和遇到的任何错误。
//
// 这是标准fmt.Print的直接封装，行为完全一致。
//
// 示例:
//
//	c.Print("Hello", " ", "World") // 输出: Hello World
//	c.Print(123, 456)             // 输出: 123 456
func (c *BaseController) Print(a ...any) (n int, err error) {
	return fmt.Print(a...)
}

// Printf 根据格式说明符格式化并写入到标准输出。
// 返回写入的字节数和遇到的任何错误。
//
// 支持所有标准fmt格式化动词，与fmt.Printf完全一致:
//
//	%v - 默认格式
//	%s - 字符串
//	%d - 十进制整数
//	%f - 浮点数
//	%t - 布尔值
//	等等...
//
// 示例:
//
//	c.Printf("用户ID: %d, 姓名: %s", 123, "张三")
//	c.Printf("温度: %.2f°C", 23.456)
func (c *BaseController) Printf(format string, a ...any) (n int, err error) {
	return fmt.Printf(format, a...)
}

// Println 使用默认格式格式化其操作数并写入到标准输出。
// 操作数之间总是添加空格，并在末尾追加换行符。
// 返回写入的字节数和遇到的任何错误。
//
// 这是标准fmt.Println的直接封装，行为完全一致。
//
// 示例:
//
//	c.Println("调试信息:", debug)
//	c.Println("状态码:", 200, "消息:", "成功")
func (c *BaseController) Println(a ...any) (n int, err error) {
	return fmt.Println(a...)
}

// ============= Sprint系列方法 - 返回格式化字符串 =============

// Sprint 使用默认格式格式化其操作数并返回结果字符串。
// 当两个连续的操作数都不是字符串时，会在它们之间添加空格。
//
// 示例:
//
//	str := c.Sprint("Hello", " ", "World") // 返回: "Hello World"
//	str := c.Sprint(123, 456)             // 返回: "123 456"
func (c *BaseController) Sprint(a ...any) string {
	return fmt.Sprint(a...)
}

// Sprintf 根据格式说明符格式化并返回结果字符串。
//
// 支持所有标准fmt格式化动词，与Printf完全一致。
//
// 示例:
//
//	str := c.Sprintf("用户ID: %d, 姓名: %s", 123, "张三")
//	str := c.Sprintf("进度: %.1f%%", 85.7)
func (c *BaseController) Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

// Sprintln 使用默认格式格式化其操作数并返回结果字符串。
// 操作数之间总是添加空格，并在末尾追加换行符。
//
// 示例:
//
//	str := c.Sprintln("调试信息:", debug)    // 返回: "调试信息: [debug内容]\n"
//	str := c.Sprintln("状态:", 200, "OK")  // 返回: "状态: 200 OK\n"
func (c *BaseController) Sprintln(a ...any) string {
	return fmt.Sprintln(a...)
}

// ============= Fprint系列方法 - 输出到指定Writer =============

// Fprint 使用默认格式格式化其操作数并写入到指定的io.Writer。
// 当两个连续的操作数都不是字符串时，会在它们之间添加空格。
// 返回写入的字节数和遇到的任何错误。
//
// 示例:
//
//	var buf bytes.Buffer
//	c.Fprint(&buf, "Hello", " ", "World")
//	result := buf.String() // "Hello World"
func (c *BaseController) Fprint(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprint(w, a...)
}

// Fprintf 根据格式说明符格式化并写入到指定的io.Writer。
// 返回写入的字节数和遇到的任何错误。
//
// 示例:
//
//	var buf bytes.Buffer
//	c.Fprintf(&buf, "用户ID: %d, 姓名: %s", 123, "张三")
func (c *BaseController) Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	return fmt.Fprintf(w, format, a...)
}

// Fprintln 使用默认格式格式化其操作数并写入到指定的io.Writer。
// 操作数之间总是添加空格，并在末尾追加换行符。
// 返回写入的字节数和遇到的任何错误。
//
// 示例:
//
//	var buf bytes.Buffer
//	c.Fprintln(&buf, "调试信息:", debug)
func (c *BaseController) Fprintln(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprintln(w, a...)
}

// ============= Scan系列方法 - 直接调用标准fmt包 =============

// Scan 从标准输入扫描由空格分隔的文本，并将连续的以空格分隔的值存储到连续的参数中。
// 换行符计为空格。返回成功扫描的项目数量。如果读取的项目数量少于参数数量，err会报告原因。
//
// 这是标准fmt.Scan的直接封装，行为完全一致。
//
// 示例:
//
//	var name string
//	var age int
//	n, err := c.Scan(&name, &age) // 从标准输入读取: "张三 25"
func (c *BaseController) Scan(a ...any) (n int, err error) {
	return fmt.Scan(a...)
}

// Scanf 从标准输入扫描文本，根据格式字符串解析参数。
// 返回成功解析的参数数量。
//
// 这是标准fmt.Scanf的直接封装，行为完全一致。
//
// 示例:
//
//	var name string
//	var age int
//	n, err := c.Scanf("%s %d", &name, &age) // 从标准输入解析: "张三 25"
func (c *BaseController) Scanf(format string, a ...any) (n int, err error) {
	return fmt.Scanf(format, a...)
}

// Scanln 类似于Scan，但在换行符处停止扫描，且在最后一个项目之后必须有换行符或EOF。
//
// 这是标准fmt.Scanln的直接封装，行为完全一致。
//
// 示例:
//
//	var name string
//	var age int
//	n, err := c.Scanln(&name, &age) // 从标准输入读取一行: "张三 25\n"
func (c *BaseController) Scanln(a ...any) (n int, err error) {
	return fmt.Scanln(a...)
}

// ============= Sscan系列方法 - 从字符串解析 =============

// Sscan 从字符串str中扫描由空格分隔的文本，并将连续的以空格分隔的值存储到连续的参数中。
// 换行符计为空格。返回成功扫描的项目数量。
//
// 示例:
//
//	var name string
//	var age int
//	n, err := c.Sscan("张三 25", &name, &age)
func (c *BaseController) Sscan(str string, a ...any) (n int, err error) {
	return fmt.Sscan(str, a...)
}

// Sscanf 从字符串str中扫描文本，根据格式字符串解析参数。
// 返回成功解析的参数数量。
//
// 示例:
//
//	var name string
//	var age int
//	n, err := c.Sscanf("姓名:张三,年龄:25", "姓名:%s,年龄:%d", &name, &age)
func (c *BaseController) Sscanf(str, format string, a ...any) (n int, err error) {
	return fmt.Sscanf(str, format, a...)
}

// Sscanln 类似于Sscan，但在换行符处停止扫描，且在最后一个项目之后必须有换行符或EOF。
//
// 示例:
//
//	var name string
//	var age int
//	n, err := c.Sscanln("张三 25\n其他内容", &name, &age) // 只解析第一行
func (c *BaseController) Sscanln(str string, a ...any) (n int, err error) {
	return fmt.Sscanln(str, a...)
}

// ============= 简化的格式化方法 =============

// FormatOutput 根据输入数据将其格式化为字符串。
//
// 这是标准fmt.Sprintf("%v", data)的简单封装，保持与fmt包的一致行为。
//
// 参数:
//
//	data: 任意类型的数据
//
// 返回值:
//
//	返回格式化后的字符串，使用Go的默认格式化规则
//
// 示例:
//
//	c.FormatOutput("hello")           // 返回: "hello"
//	c.FormatOutput(123)               // 返回: "123"
//	c.FormatOutput([]int{1,2,3})      // 返回: "[1 2 3]"
//	c.FormatOutput(map[string]int{"a":1}) // 返回: "map[a:1]"
func (c *BaseController) FormatOutput(data any) string {
	return fmt.Sprintf("%v", data)
}

// Errorf 返回一个格式化的错误信息。
// 参数：
//   - format: 格式化字符串模板。
//   - a: 可变参数，用于填充格式化字符串。
//
// 返回值：
//   - error: 格式化后的错误信息。
func (c *BaseController) Errorf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}
