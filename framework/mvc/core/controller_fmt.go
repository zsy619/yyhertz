package core

// 实现控制器的格式化输出，全部兼容fmt的输出
import (
	"fmt"
	"reflect"
)

// FormatOutput 根据输入数据的类型将其格式化为字符串。
//
// 参数:
//
//	data: 任意类型的数据，支持字符串、整数、浮点数等基本类型。
//
// 返回值:
//
//	返回格式化后的字符串：
//	- 字符串类型：添加双引号（如 "example"）
//	- 整数类型：直接转换为字符串（如 "123"）
//	- 浮点数类型：保留小数位（如 "3.140000"）
//	- 其他类型：使用默认格式（如结构体会转换为 "{Field: value}" 形式）
//
// 注意:
//
//	该方法主要用于内部数据格式化，适用于日志输出或简单调试场景。
func (c *BaseController) FormatOutput(data any) string {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", val.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", val.Float())
	case reflect.Slice:
		return fmt.Sprintf("[%v]", val.Interface())
	case reflect.Map:
		return fmt.Sprintf("%v", val.Interface())
	default:
		return fmt.Sprintf("%v", data)
	}
}
