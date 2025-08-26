package util

import (
	"fmt"
	"time"
)

// ToShortTimeFormat 将时间间隔（time.Duration）转换为简短的字符串表示形式。
// 根据时间间隔的长度，返回不同单位的格式化字符串：
// - 小于1秒：返回纳秒（ns）、微秒（us）或毫秒（ms）单位，保留两位小数。
// - 大于等于1秒：返回秒（s）、分钟（m）或小时（h）单位，保留两位小数。
// 例如：
//   - 0 返回 "0"
//   - 100纳秒 返回 "100.00ns"
//   - 1.5秒 返回 "1.50s"
//   - 90分钟 返回 "1.50h"
//
// 参数：
//
//	d: 需要格式化的时间间隔。
//
// 返回值：
//
//	格式化后的字符串。
func ToShortTimeFormat(d time.Duration) string {
	u := uint64(d)
	if u < uint64(time.Second) {
		switch {
		case u == 0:
			return "0"
		case u < uint64(time.Microsecond):
			return fmt.Sprintf("%.2fns", float64(u))
		case u < uint64(time.Millisecond):
			return fmt.Sprintf("%.2fus", float64(u)/1000)
		default:
			return fmt.Sprintf("%.2fms", float64(u)/1000/1000)
		}
	} else {
		switch {
		case u < uint64(time.Minute):
			return fmt.Sprintf("%.2fs", float64(u)/1000/1000/1000)
		case u < uint64(time.Hour):
			return fmt.Sprintf("%.2fm", float64(u)/1000/1000/1000/60)
		default:
			return fmt.Sprintf("%.2fh", float64(u)/1000/1000/1000/60/60)
		}
	}
}
