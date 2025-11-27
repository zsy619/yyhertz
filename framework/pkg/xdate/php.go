package xdate

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

// NewDateTime creates a new DateTime instance
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

// Format formats the DateTime using PHP date format
func (dt *DateTime) Format(format string) string {
	return formatTime(dt.time.In(dt.timezone), format)
}

// SetDate sets the date
func (dt *DateTime) SetDate(year, month, day int) {
	dt.time = time.Date(year, time.Month(month), day,
		dt.time.Hour(), dt.time.Minute(), dt.time.Second(), dt.time.Nanosecond(),
		dt.timezone)
}

// SetTime sets the time
func (dt *DateTime) SetTime(hour, minute, second int, microsecond ...int) {
	micro := 0
	if len(microsecond) > 0 {
		micro = microsecond[0]
	}
	
	dt.time = time.Date(dt.time.Year(), dt.time.Month(), dt.time.Day(),
		hour, minute, second, micro*1000,
		dt.timezone)
}

// SetTimestamp sets the DateTime from Unix timestamp
func (dt *DateTime) SetTimestamp(timestamp int64) {
	dt.time = time.Unix(timestamp, 0).In(dt.timezone)
}

// GetTimestamp gets the Unix timestamp
func (dt *DateTime) GetTimestamp() int64 {
	return dt.time.Unix()
}

// Add adds a DateInterval
func (dt *DateTime) Add(interval *DateInterval) {
	years := interval.Years
	months := interval.Months
	days := interval.Days
	duration := time.Duration(interval.Hours)*time.Hour +
		time.Duration(interval.Minutes)*time.Minute +
		time.Duration(interval.Seconds)*time.Second

	if interval.Invert {
		years = -years
		months = -months
		days = -days
		duration = -duration
	}

	dt.time = dt.time.AddDate(years, months, days).Add(duration)
}

// Sub subtracts a DateInterval
func (dt *DateTime) Sub(interval *DateInterval) {
	interval.Invert = !interval.Invert
	dt.Add(interval)
	interval.Invert = !interval.Invert // restore original state
}

// Diff calculates the difference between two DateTime objects
func (dt *DateTime) Diff(other *DateTime) *DateInterval {
	diff := other.time.Sub(dt.time)
	invert := diff < 0
	if invert {
		diff = -diff
	}

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	return &DateInterval{
		Days:    days,
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
		Invert:  invert,
	}
}

// SetTimezone sets the timezone
func (dt *DateTime) SetTimezone(timezone *DateTimeZone) {
	dt.timezone = timezone.location
	dt.time = dt.time.In(dt.timezone)
}

// GetTimezone gets the timezone
func (dt *DateTime) GetTimezone() *DateTimeZone {
	return &DateTimeZone{location: dt.timezone}
}

// NewDateTimeZone creates a new DateTimeZone
func NewDateTimeZone(timezone string) (*DateTimeZone, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	return &DateTimeZone{location: loc}, nil
}

// GetName gets the timezone name
func (dtz *DateTimeZone) GetName() string {
	return dtz.location.String()
}

// GetOffset gets the timezone offset for a specific DateTime
func (dtz *DateTimeZone) GetOffset(datetime *DateTime) int {
	_, offset := datetime.time.In(dtz.location).Zone()
	return offset
}

// NewDateInterval creates a new DateInterval from interval spec
func NewDateInterval(intervalSpec string) (*DateInterval, error) {
	// Simple parser for basic interval specs like "P1Y2M3DT4H5M6S"
	if !strings.HasPrefix(intervalSpec, "P") {
		return nil, fmt.Errorf("invalid interval spec: %s", intervalSpec)
	}

	spec := strings.TrimPrefix(intervalSpec, "P")
	interval := &DateInterval{}

	// Split by T to separate date and time parts
	parts := strings.Split(spec, "T")
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

// Helper function to parse date part of interval spec
func parseDatePart(part string, interval *DateInterval) error {
	for i := 0; i < len(part); {
		j := i
		for j < len(part) && (part[j] >= '0' && part[j] <= '9') {
			j++
		}
		if j == i {
			i++
			continue
		}
		
		num, err := strconv.Atoi(part[i:j])
		if err != nil {
			return err
		}
		
		if j >= len(part) {
			return fmt.Errorf("incomplete interval spec")
		}
		
		switch part[j] {
		case 'Y':
			interval.Years = num
		case 'M':
			interval.Months = num
		case 'D':
			interval.Days = num
		}
		i = j + 1
	}
	return nil
}

// Helper function to parse time part of interval spec
func parseTimePart(part string, interval *DateInterval) error {
	for i := 0; i < len(part); {
		j := i
		for j < len(part) && (part[j] >= '0' && part[j] <= '9') {
			j++
		}
		if j == i {
			i++
			continue
		}
		
		num, err := strconv.Atoi(part[i:j])
		if err != nil {
			return err
		}
		
		if j >= len(part) {
			return fmt.Errorf("incomplete interval spec")
		}
		
		switch part[j] {
		case 'H':
			interval.Hours = num
		case 'M':
			interval.Minutes = num
		case 'S':
			interval.Seconds = num
		}
		i = j + 1
	}
	return nil
}

// convertPHPFormatToGo converts PHP date format to Go time format
func convertPHPFormatToGo(phpFormat string) string {
	goFormat := phpFormat
	
	// Simple mappings - this is a basic implementation
	replacements := map[string]string{
		"Y": "2006",
		"m": "01",
		"d": "02",
		"H": "15",
		"i": "04",
		"s": "05",
		"y": "06",
		"M": "Jan",
		"F": "January",
		"j": "2",
		"n": "1",
		"G": "15", // same as H for simplicity
		"g": "3",
		"A": "PM",
		"a": "pm",
	}
	
	for php, go_ := range replacements {
		goFormat = strings.ReplaceAll(goFormat, php, go_)
	}
	
	return goFormat
}