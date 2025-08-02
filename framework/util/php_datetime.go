package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DateTime represents a PHP-like DateTime class
type DateTime struct {
	time     time.Time
	timezone *time.Location
}

// DateInterval represents a PHP-like DateInterval class
type DateInterval struct {
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
	Invert  bool // true for negative intervals
}

// DateTimeZone represents a PHP-like DateTimeZone class
type DateTimeZone struct {
	location *time.Location
}

// Constants for PHP DateTime
const (
	ATOM    = "2006-01-02T15:04:05-07:00"
	COOKIE  = "Monday, 02-Jan-2006 15:04:05 MST"
	ISO8601 = "2006-01-02T15:04:05-0700"
	RFC822  = "Mon, 02 Jan 06 15:04:05 -0700"
	RFC850  = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1036 = "Mon, 02 Jan 06 15:04:05 -0700"
	RFC1123 = "Mon, 02 Jan 2006 15:04:05 -0700"
	RFC2822 = "Mon, 02 Jan 2006 15:04:05 -0700"
	RFC3339 = "2006-01-02T15:04:05-07:00"
	RSS     = "Mon, 02 Jan 2006 15:04:05 -0700"
	W3C     = "2006-01-02T15:04:05-07:00"
)

// NewDateTime creates a new DateTime instance)
func NewDateTime(datetime ...string) (*DateTime, error) {
	var t time.Time
	var err error

	if len(datetime) > 0 && datetime[0] != "" {
		// Try to parse the datetime string
		timestamp, parseErr := Strtotime(datetime[0])
		if parseErr != nil {
			return nil, parseErr
		}
		t = time.Unix(timestamp, 0)
	} else {
		t = time.Now()
	}

	return &DateTime{
		time:     t,
		timezone: time.Local,
	}, err
}

// NewDateTimeFromFormat creates a DateTime from a format
func NewDateTimeFromFormat(format, datetime string, timezone ...*DateTimeZone) (*DateTime, error) {
	// Convert PHP format to Go format
	goFormat := convertPHPFormatToGo(format)

	var loc *time.Location = time.Local
	if len(timezone) > 0 {
		loc = timezone[0].location
	}

	t, err := time.ParseInLocation(goFormat, datetime, loc)
	if err != nil {
		return nil, err
	}

	return &DateTime{
		time:     t,
		timezone: loc,
	}, nil
}

// NewDateTimeFromTimestamp creates a DateTime from Unix timestamp
func NewDateTimeFromTimestamp(timestamp int64, timezone ...*DateTimeZone) *DateTime {
	var loc *time.Location = time.Local
	if len(timezone) > 0 {
		loc = timezone[0].location
	}

	return &DateTime{
		time:     time.Unix(timestamp, 0).In(loc),
		timezone: loc,
	}
}

// Format formats the DateTime
func (dt *DateTime) Format(format string) string {
	return formatTime(dt.time, format)
}

// GetTimestamp returns the Unix timestamp
func (dt *DateTime) GetTimestamp() int64 {
	return dt.time.Unix()
}

// SetTimestamp sets the Unix timestamp
func (dt *DateTime) SetTimestamp(timestamp int64) *DateTime {
	dt.time = time.Unix(timestamp, 0).In(dt.timezone)
	return dt
}

// SetDate sets the date
func (dt *DateTime) SetDate(year, month, day int) *DateTime {
	dt.time = time.Date(year, time.Month(month), day,
		dt.time.Hour(), dt.time.Minute(), dt.time.Second(),
		dt.time.Nanosecond(), dt.timezone)
	return dt
}

// SetTime sets the time
func (dt *DateTime) SetTime(hour, minute, second int, microsecond ...int) *DateTime {
	micro := 0
	if len(microsecond) > 0 {
		micro = microsecond[0] * 1000 // convert to nanoseconds
	}

	dt.time = time.Date(dt.time.Year(), dt.time.Month(), dt.time.Day(),
		hour, minute, second, micro, dt.timezone)
	return dt
}

// Add adds an interval to the DateTime
func (dt *DateTime) Add(interval *DateInterval) *DateTime {
	if interval.Invert {
		dt.time = dt.time.AddDate(-interval.Years, -interval.Months, -interval.Days)
		dt.time = dt.time.Add(-time.Duration(interval.Hours) * time.Hour)
		dt.time = dt.time.Add(-time.Duration(interval.Minutes) * time.Minute)
		dt.time = dt.time.Add(-time.Duration(interval.Seconds) * time.Second)
	} else {
		dt.time = dt.time.AddDate(interval.Years, interval.Months, interval.Days)
		dt.time = dt.time.Add(time.Duration(interval.Hours) * time.Hour)
		dt.time = dt.time.Add(time.Duration(interval.Minutes) * time.Minute)
		dt.time = dt.time.Add(time.Duration(interval.Seconds) * time.Second)
	}
	return dt
}

// Sub subtracts an interval from the DateTime
func (dt *DateTime) Sub(interval *DateInterval) *DateTime {
	if interval.Invert {
		dt.time = dt.time.AddDate(interval.Years, interval.Months, interval.Days)
		dt.time = dt.time.Add(time.Duration(interval.Hours) * time.Hour)
		dt.time = dt.time.Add(time.Duration(interval.Minutes) * time.Minute)
		dt.time = dt.time.Add(time.Duration(interval.Seconds) * time.Second)
	} else {
		dt.time = dt.time.AddDate(-interval.Years, -interval.Months, -interval.Days)
		dt.time = dt.time.Add(-time.Duration(interval.Hours) * time.Hour)
		dt.time = dt.time.Add(-time.Duration(interval.Minutes) * time.Minute)
		dt.time = dt.time.Add(-time.Duration(interval.Seconds) * time.Second)
	}
	return dt
}

// Diff calculates the difference between two DateTime objects
func (dt *DateTime) Diff(other *DateTime, absolute ...bool) *DateInterval {
	duration := other.time.Sub(dt.time)

	isNegative := duration < 0
	if isNegative {
		duration = -duration
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	interval := &DateInterval{
		Days:    days,
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
		Invert:  isNegative && (len(absolute) == 0 || !absolute[0]),
	}

	return interval
}

// Modify alters the timestamp
func (dt *DateTime) Modify(modify string) (*DateTime, error) {
	// Parse modification string
	modify = strings.ToLower(strings.TrimSpace(modify))

	switch {
	case strings.Contains(modify, "next monday"):
		for dt.time.Weekday() != time.Monday {
			dt.time = dt.time.AddDate(0, 0, 1)
		}
	case strings.Contains(modify, "last monday"):
		for dt.time.Weekday() != time.Monday {
			dt.time = dt.time.AddDate(0, 0, -1)
		}
	case strings.Contains(modify, "+1 day"):
		dt.time = dt.time.AddDate(0, 0, 1)
	case strings.Contains(modify, "-1 day"):
		dt.time = dt.time.AddDate(0, 0, -1)
	case strings.Contains(modify, "+1 week"):
		dt.time = dt.time.AddDate(0, 0, 7)
	case strings.Contains(modify, "-1 week"):
		dt.time = dt.time.AddDate(0, 0, -7)
	case strings.Contains(modify, "+1 month"):
		dt.time = dt.time.AddDate(0, 1, 0)
	case strings.Contains(modify, "-1 month"):
		dt.time = dt.time.AddDate(0, -1, 0)
	case strings.Contains(modify, "+1 year"):
		dt.time = dt.time.AddDate(1, 0, 0)
	case strings.Contains(modify, "-1 year"):
		dt.time = dt.time.AddDate(-1, 0, 0)
	default:
		return dt, fmt.Errorf("unsupported modify string: %s", modify)
	}

	return dt, nil
}

// SetTimezone sets the timezone
func (dt *DateTime) SetTimezone(timezone *DateTimeZone) *DateTime {
	dt.time = dt.time.In(timezone.location)
	dt.timezone = timezone.location
	return dt
}

// GetTimezone returns the timezone
func (dt *DateTime) GetTimezone() *DateTimeZone {
	return &DateTimeZone{location: dt.timezone}
}

// GetOffset returns the timezone offset
func (dt *DateTime) GetOffset() int {
	_, offset := dt.time.Zone()
	return offset
}

// Clone creates a copy of the DateTime
func (dt *DateTime) Clone() *DateTime {
	return &DateTime{
		time:     dt.time,
		timezone: dt.timezone,
	}
}

// String returns string representation
func (dt *DateTime) String() string {
	return dt.Format("Y-m-d H:i:s")
}

// NewDateInterval creates a new DateInterval)
func NewDateInterval(intervalSpec string) (*DateInterval, error) {
	// Parse ISO 8601 duration format or simple format
	// Example: P1Y2M3DT4H5M6S or "1 day", "2 hours", etc.

	interval := &DateInterval{}

	if strings.HasPrefix(intervalSpec, "P") {
		// ISO 8601 format
		return parseISO8601Duration(intervalSpec)
	}

	// Simple format parsing
	parts := strings.Fields(strings.ToLower(intervalSpec))
	if len(parts) >= 2 {
		value, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		unit := parts[1]
		switch {
		case strings.Contains(unit, "year"):
			interval.Years = value
		case strings.Contains(unit, "month"):
			interval.Months = value
		case strings.Contains(unit, "day"):
			interval.Days = value
		case strings.Contains(unit, "hour"):
			interval.Hours = value
		case strings.Contains(unit, "minute"):
			interval.Minutes = value
		case strings.Contains(unit, "second"):
			interval.Seconds = value
		}
	}

	return interval, nil
}

// Format formats the DateInterval
func (di *DateInterval) Format(format string) string {
	result := format

	result = strings.ReplaceAll(result, "%Y", fmt.Sprintf("%d", di.Years))
	result = strings.ReplaceAll(result, "%M", fmt.Sprintf("%d", di.Months))
	result = strings.ReplaceAll(result, "%D", fmt.Sprintf("%d", di.Days))
	result = strings.ReplaceAll(result, "%H", fmt.Sprintf("%d", di.Hours))
	result = strings.ReplaceAll(result, "%I", fmt.Sprintf("%d", di.Minutes))
	result = strings.ReplaceAll(result, "%S", fmt.Sprintf("%d", di.Seconds))
	result = strings.ReplaceAll(result, "%R", map[bool]string{true: "-", false: "+"}[di.Invert])

	return result
}

// NewDateTimeZone creates a new DateTimeZone)
func NewDateTimeZone(timezone string) (*DateTimeZone, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	return &DateTimeZone{location: loc}, nil
}

// GetName returns the timezone name
func (dtz *DateTimeZone) GetName() string {
	return dtz.location.String()
}

// GetOffset returns the timezone offset
func (dtz *DateTimeZone) GetOffset(datetime *DateTime) int {
	_, offset := datetime.time.In(dtz.location).Zone()
	return offset
}

// ListAbbreviations returns timezone abbreviations
func ListAbbreviations() map[string][]map[string]any {
	// Simplified implementation
	return map[string][]map[string]any{
		"utc": {
			{"dst": false, "offset": 0, "timezone_id": "UTC"},
		},
		"est": {
			{"dst": false, "offset": -18000, "timezone_id": "America/New_York"},
		},
		"pst": {
			{"dst": false, "offset": -28800, "timezone_id": "America/Los_Angeles"},
		},
	}
}

// ListIdentifiers returns timezone identifiers
func ListIdentifiers(what ...int) []string {
	// Simplified implementation - return common timezones
	return []string{
		"UTC",
		"America/New_York",
		"America/Los_Angeles",
		"America/Chicago",
		"Europe/London",
		"Europe/Paris",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Australia/Sydney",
	}
}

// Helper functions

func parseISO8601Duration(duration string) (*DateInterval, error) {
	// Parse ISO 8601 duration format: P[n]Y[n]M[n]DT[n]H[n]M[n]S
	if !strings.HasPrefix(duration, "P") {
		return nil, fmt.Errorf("invalid ISO 8601 duration format")
	}

	interval := &DateInterval{}
	duration = duration[1:] // Remove 'P'

	// Split by 'T' for date and time parts
	parts := strings.Split(duration, "T")
	datePart := parts[0]
	timePart := ""
	if len(parts) > 1 {
		timePart = parts[1]
	}

	// Parse date part
	if err := parseDatePart(datePart, interval); err != nil {
		return nil, err
	}

	// Parse time part
	if timePart != "" {
		if err := parseTimePart(timePart, interval); err != nil {
			return nil, err
		}
	}

	return interval, nil
}

func parseDatePart(datePart string, interval *DateInterval) error {
	var current strings.Builder

	for _, char := range datePart {
		if char >= '0' && char <= '9' {
			current.WriteRune(char)
		} else {
			if current.Len() > 0 {
				value, err := strconv.Atoi(current.String())
				if err != nil {
					return err
				}

				switch char {
				case 'Y':
					interval.Years = value
				case 'M':
					interval.Months = value
				case 'D':
					interval.Days = value
				}

				current.Reset()
			}
		}
	}

	return nil
}

func parseTimePart(timePart string, interval *DateInterval) error {
	var current strings.Builder

	for _, char := range timePart {
		if char >= '0' && char <= '9' {
			current.WriteRune(char)
		} else {
			if current.Len() > 0 {
				value, err := strconv.Atoi(current.String())
				if err != nil {
					return err
				}

				switch char {
				case 'H':
					interval.Hours = value
				case 'M':
					interval.Minutes = value
				case 'S':
					interval.Seconds = value
				}

				current.Reset()
			}
		}
	}

	return nil
}

func convertPHPFormatToGo(phpFormat string) string {
	// Convert PHP date format to Go time format
	goFormat := phpFormat

	// Year
	goFormat = strings.ReplaceAll(goFormat, "Y", "2006")
	goFormat = strings.ReplaceAll(goFormat, "y", "06")

	// Month
	goFormat = strings.ReplaceAll(goFormat, "m", "01")
	goFormat = strings.ReplaceAll(goFormat, "n", "1")
	goFormat = strings.ReplaceAll(goFormat, "M", "Jan")
	goFormat = strings.ReplaceAll(goFormat, "F", "January")

	// Day
	goFormat = strings.ReplaceAll(goFormat, "d", "02")
	goFormat = strings.ReplaceAll(goFormat, "j", "2")
	goFormat = strings.ReplaceAll(goFormat, "D", "Mon")
	goFormat = strings.ReplaceAll(goFormat, "l", "Monday")

	// Time
	goFormat = strings.ReplaceAll(goFormat, "H", "15")
	goFormat = strings.ReplaceAll(goFormat, "h", "03")
	goFormat = strings.ReplaceAll(goFormat, "i", "04")
	goFormat = strings.ReplaceAll(goFormat, "s", "05")
	goFormat = strings.ReplaceAll(goFormat, "A", "PM")
	goFormat = strings.ReplaceAll(goFormat, "a", "pm")

	// Timezone
	goFormat = strings.ReplaceAll(goFormat, "T", "MST")
	goFormat = strings.ReplaceAll(goFormat, "O", "-0700")
	goFormat = strings.ReplaceAll(goFormat, "P", "-07:00")

	return goFormat
}

// Utility functions for common operations

// DateTimeGetLastErrors returns warnings and errors
func DateTimeGetLastErrors() map[string]any {
	return map[string]any{
		"warning_count": 0,
		"warnings":      []string{},
		"error_count":   0,
		"errors":        []string{},
	}
}

// DateDefaultTimezoneGet returns the default timezone
func DateDefaultTimezoneGet() string {
	return time.Local.String()
}

// DateDefaultTimezoneSet sets the default timezone
func DateDefaultTimezoneSet(timezoneId string) bool {
	loc, err := time.LoadLocation(timezoneId)
	if err != nil {
		return false
	}
	time.Local = loc
	return true
}

// DateTimezoneGet returns the timezone
func DateTimezoneGet(object *DateTime) *DateTimeZone {
	return object.GetTimezone()
}

// DateTimezoneSet sets the timezone
func DateTimezoneSet(object *DateTime, timezone *DateTimeZone) *DateTime {
	return object.SetTimezone(timezone)
}

// DateIntervalCreateFromDateString creates interval from relative parts
func DateIntervalCreateFromDateString(time string) (*DateInterval, error) {
	return NewDateInterval(time)
}

// DateSunInfo returns sunrise/sunset and twilight begin/end
func DateSunInfo(timestamp int64, latitude, longitude float64) map[string]any {
	// Simplified implementation - return placeholder values
	return map[string]any{
		"sunrise":                     timestamp + 21600, // 6 AM
		"sunset":                      timestamp + 64800, // 6 PM
		"transit":                     timestamp + 43200, // 12 PM
		"civil_twilight_begin":        timestamp + 20700, // 5:45 AM
		"civil_twilight_end":          timestamp + 65700, // 6:15 PM
		"nautical_twilight_begin":     timestamp + 19800, // 5:30 AM
		"nautical_twilight_end":       timestamp + 66600, // 6:30 PM
		"astronomical_twilight_begin": timestamp + 18900, // 5:15 AM
		"astronomical_twilight_end":   timestamp + 67500, // 6:45 PM
	}
}
