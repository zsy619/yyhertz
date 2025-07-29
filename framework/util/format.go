package util

import (
	"fmt"
	"strings"
	"time"
)

// FmtByte 字节单位转换
func FmtByte(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%.2fB", float64(size))
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(size)/(1024*1024))
	} else if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fGB", float64(size)/(1024*1024*1024))
	} else if size < 1024*1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fTB", float64(size)/(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fPB", float64(size)/(1024*1024*1024*1024*1024))
	}
}

// FmtFloat 格式化浮点数
func FmtFloat(value float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, value)
}

// FmtFloat2 格式化浮点数(保留2位小数)
func FmtFloat2(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// FmtFloat3 格式化浮点数(保留3位小数)
func FmtFloat3(value float64) string {
	return fmt.Sprintf("%.3f", value)
}

// FmtFloat4 格式化浮点数(保留4位小数)
func FmtFloat4(value float64) string {
	return fmt.Sprintf("%.4f", value)
}

// FmtFloat5 格式化浮点数(保留5位小数)
func FmtFloat5(value float64) string {
	return fmt.Sprintf("%.5f", value)
}

// FmtString 格式化字符串(指定宽度)
func FmtString(value string, width int) string {
	return fmt.Sprintf("%*s", width, value)
}

// GetTime 获取当前时间字符串
func GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetTimestamp 获取当前时间戳
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// GetTimeWithFormat 获取指定格式的时间字符串
func GetTimeWithFormat(format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return time.Now().Format(format)
}

// FormatTime 格式化时间
func FormatTime(t time.Time, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return t.Format(format)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr, format string) (time.Time, error) {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return time.Parse(format, timeStr)
}

// ConcatString 连接字符串
func ConcatString(strs ...string) string {
	var result strings.Builder
	for _, str := range strs {
		result.WriteString(str)
	}
	return result.String()
}

// ContainString 检查字符串是否包含子字符串(逗号分隔)
func ContainString(s, substr string) bool {
	return strings.Contains(","+s+",", ","+substr+",")
}

// MakeSlice 创建任意类型的切片
func MakeSlice(args ...any) []any {
	return args
}