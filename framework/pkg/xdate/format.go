package xdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Time returns current Unix timestamp
func Time() int64 {
	return time.Now().Unix()
}

// Date formats a local time/date
func Date(format string, timestamp ...int64) string {
	var t time.Time
	if len(timestamp) > 0 {
		t = time.Unix(timestamp[0], 0)
	} else {
		t = time.Now()
	}

	return formatTime(t, format)
}

// GmDate formats a GMT/UTC date/time
func GmDate(format string, timestamp ...int64) string {
	var t time.Time
	if len(timestamp) > 0 {
		t = time.Unix(timestamp[0], 0).UTC()
	} else {
		t = time.Now().UTC()
	}

	return formatTime(t, format)
}

// Microtime returns current Unix timestamp with microseconds
func Microtime(getAsFloat ...bool) any {
	now := time.Now()

	if len(getAsFloat) > 0 && getAsFloat[0] {
		return float64(now.UnixNano()) / 1e9
	}

	sec := now.Unix()
	usec := (now.UnixNano() % 1e9) / 1e6
	return fmt.Sprintf("0.%06d %d", usec*1000, sec)
}

// Getdate gets date/time information
func Getdate(timestamp ...int64) map[string]any {
	var t time.Time
	if len(timestamp) > 0 {
		t = time.Unix(timestamp[0], 0)
	} else {
		t = time.Now()
	}

	return map[string]any{
		"seconds": t.Second(),
		"minutes": t.Minute(),
		"hours":   t.Hour(),
		"mday":    t.Day(),
		"wday":    int(t.Weekday()),
		"mon":     int(t.Month()),
		"year":    t.Year(),
		"yday":    t.YearDay() - 1, // PHP uses 0-based
		"weekday": t.Weekday().String(),
		"month":   t.Month().String(),
		"0":       t.Unix(),
	}
}

// formatTime formats time according to PHP date format
func formatTime(t time.Time, format string) string {
	var result strings.Builder

	for i := 0; i < len(format); i++ {
		char := format[i]

		switch char {
		case 'd': // Day of the month, 2 digits with leading zeros
			result.WriteString(fmt.Sprintf("%02d", t.Day()))
		case 'D': // A textual representation of a day, three letters
			result.WriteString(t.Weekday().String()[:3])
		case 'j': // Day of the month without leading zeros
			result.WriteString(strconv.Itoa(t.Day()))
		case 'l': // A full textual representation of the day of the week
			result.WriteString(t.Weekday().String())
		case 'N': // ISO-8601 numeric representation of the day of the week
			weekday := int(t.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday = 7 in ISO-8601
			}
			result.WriteString(strconv.Itoa(weekday))
		case 'S': // English ordinal suffix for the day of the month, 2 characters
			day := t.Day()
			suffix := "th"
			if day%10 == 1 && day != 11 {
				suffix = "st"
			} else if day%10 == 2 && day != 12 {
				suffix = "nd"
			} else if day%10 == 3 && day != 13 {
				suffix = "rd"
			}
			result.WriteString(suffix)
		case 'w': // Numeric representation of the day of the week
			result.WriteString(strconv.Itoa(int(t.Weekday())))
		case 'z': // The day of the year (starting from 0)
			result.WriteString(strconv.Itoa(t.YearDay() - 1))
		case 'W': // ISO-8601 week number of year
			_, week := t.ISOWeek()
			result.WriteString(fmt.Sprintf("%02d", week))
		case 'F': // A full textual representation of a month
			result.WriteString(t.Month().String())
		case 'm': // Numeric representation of a month, with leading zeros
			result.WriteString(fmt.Sprintf("%02d", int(t.Month())))
		case 'M': // A short textual representation of a month, three letters
			result.WriteString(t.Month().String()[:3])
		case 'n': // Numeric representation of a month, without leading zeros
			result.WriteString(strconv.Itoa(int(t.Month())))
		case 't': // Number of days in the given month
			result.WriteString(strconv.Itoa(daysInMonth(t.Year(), int(t.Month()))))
		case 'L': // Whether it's a leap year
			if isLeapYear(t.Year()) {
				result.WriteString("1")
			} else {
				result.WriteString("0")
			}
		case 'o': // ISO-8601 week-numbering year
			year, _ := t.ISOWeek()
			result.WriteString(strconv.Itoa(year))
		case 'Y': // A full numeric representation of a year, 4 digits
			result.WriteString(strconv.Itoa(t.Year()))
		case 'y': // A two digit representation of a year
			result.WriteString(fmt.Sprintf("%02d", t.Year()%100))
		case 'a': // Lowercase Ante meridiem and Post meridiem
			if t.Hour() < 12 {
				result.WriteString("am")
			} else {
				result.WriteString("pm")
			}
		case 'A': // Uppercase Ante meridiem and Post meridiem
			if t.Hour() < 12 {
				result.WriteString("AM")
			} else {
				result.WriteString("PM")
			}
		case 'B': // Swatch Internet time
			// Simplified implementation
			result.WriteString("000")
		case 'g': // 12-hour format of an hour without leading zeros
			hour := t.Hour() % 12
			if hour == 0 {
				hour = 12
			}
			result.WriteString(strconv.Itoa(hour))
		case 'G': // 24-hour format of an hour without leading zeros
			result.WriteString(strconv.Itoa(t.Hour()))
		case 'h': // 12-hour format of an hour with leading zeros
			hour := t.Hour() % 12
			if hour == 0 {
				hour = 12
			}
			result.WriteString(fmt.Sprintf("%02d", hour))
		case 'H': // 24-hour format of an hour with leading zeros
			result.WriteString(fmt.Sprintf("%02d", t.Hour()))
		case 'i': // Minutes with leading zeros
			result.WriteString(fmt.Sprintf("%02d", t.Minute()))
		case 's': // Seconds with leading zeros
			result.WriteString(fmt.Sprintf("%02d", t.Second()))
		case 'u': // Microseconds
			result.WriteString(fmt.Sprintf("%06d", t.Nanosecond()/1000))
		case 'v': // Milliseconds
			result.WriteString(fmt.Sprintf("%03d", t.Nanosecond()/1000000))
		case 'e': // Timezone identifier
			zone, _ := t.Zone()
			result.WriteString(zone)
		case 'I': // Whether or not the date is in daylight saving time
			result.WriteString("0") // Simplified
		case 'O': // Difference to Greenwich time (GMT) in hours
			_, offset := t.Zone()
			hours := offset / 3600
			minutes := (offset % 3600) / 60
			result.WriteString(fmt.Sprintf("%+03d%02d", hours, minutes))
		case 'P': // Difference to Greenwich time (GMT) with colon between hours and minutes
			_, offset := t.Zone()
			hours := offset / 3600
			minutes := (offset % 3600) / 60
			result.WriteString(fmt.Sprintf("%+03d:%02d", hours, minutes))
		case 'T': // Timezone abbreviation
			zone, _ := t.Zone()
			result.WriteString(zone)
		case 'Z': // Timezone offset in seconds
			_, offset := t.Zone()
			result.WriteString(strconv.Itoa(offset))
		case 'c': // ISO 8601 date
			result.WriteString(t.Format("2006-01-02T15:04:05-07:00"))
		case 'r': // RFC 2822 formatted date
			result.WriteString(t.Format("Mon, 02 Jan 2006 15:04:05 -0700"))
		case 'U': // Seconds since the Unix Epoch
			result.WriteString(strconv.FormatInt(t.Unix(), 10))
		case '\\': // Escape character
			if i+1 < len(format) {
				i++
				result.WriteByte(format[i])
			}
		default:
			result.WriteByte(char)
		}
	}

	return result.String()
}

// Helper functions

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func daysInMonth(year, month int) int {
	switch month {
	case 2:
		if isLeapYear(year) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

// ToShortTimeFormat 转换时间为短格式字符串 (兼容性函数)
func ToShortTimeFormat(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.2fns", float64(d.Nanoseconds()))
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000.0)
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.2fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.2fh", d.Hours())
	}
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

// ParseTime 解析时间字符串
func ParseTime(timeStr, format string) (time.Time, error) {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return time.Parse(format, timeStr)
}
