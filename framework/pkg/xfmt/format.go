package xfmt

import (
	"fmt"
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

// FmtTime 格式化时间
func FmtTime(t time.Time, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return t.Format(format)
}
