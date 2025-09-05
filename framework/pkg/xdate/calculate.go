package xdate

import (
	"fmt"
	"strings"
	"time"
)

// Mktime returns Unix timestamp for a date
func Mktime(hour, minute, second, month, day, year int) int64 {
	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
	return t.Unix()
}

// Gmmktime returns Unix timestamp for a GMT date
func Gmmktime(hour, minute, second, month, day, year int) int64 {
	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return t.Unix()
}

// Strtotime parses about any English textual datetime description into a Unix timestamp
func Strtotime(timestr string, now ...int64) (int64, error) {
	var baseTime time.Time
	if len(now) > 0 {
		baseTime = time.Unix(now[0], 0)
	} else {
		baseTime = time.Now()
	}

	// Handle common formats
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"15:04:05",
		"January 2, 2006",
		"Jan 2, 2006",
		"02/01/2006",
		"01/02/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timestr); err == nil {
			return t.Unix(), nil
		}
	}

	// Handle relative time strings
	timestr = strings.ToLower(strings.TrimSpace(timestr))
	switch timestr {
	case "now":
		return baseTime.Unix(), nil
	case "today":
		today := time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, baseTime.Location())
		return today.Unix(), nil
	case "tomorrow":
		tomorrow := baseTime.AddDate(0, 0, 1)
		tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
		return tomorrow.Unix(), nil
	case "yesterday":
		yesterday := baseTime.AddDate(0, 0, -1)
		yesterday = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
		return yesterday.Unix(), nil
	}

	return 0, fmt.Errorf("unable to parse time string: %s", timestr)
}

// Checkdate validates a Gregorian date
func Checkdate(month, day, year int) bool {
	if month < 1 || month > 12 {
		return false
	}
	if year < 1 || year > 32767 {
		return false
	}
	if day < 1 {
		return false
	}

	// Check days per month
	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// Check for leap year
	if month == 2 && isLeapYear(year) {
		return day <= 29
	}

	return day <= daysInMonth[month-1]
}

// DateDiff calculates the difference between two dates
func DateDiff(datetime1, datetime2 time.Time) map[string]int {
	diff := datetime2.Sub(datetime1)

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	return map[string]int{
		"days":    days,
		"hours":   hours,
		"minutes": minutes,
		"seconds": seconds,
	}
}

// DateAdd adds an amount of days, months, years, hours, minutes and seconds to a date
func DateAdd(t time.Time, years, months, days, hours, minutes, seconds int) time.Time {
	return t.AddDate(years, months, days).
		Add(time.Duration(hours) * time.Hour).
		Add(time.Duration(minutes) * time.Minute).
		Add(time.Duration(seconds) * time.Second)
}

// DateSub subtracts an amount of days, months, years, hours, minutes and seconds from a date
func DateSub(t time.Time, years, months, days, hours, minutes, seconds int) time.Time {
	return DateAdd(t, -years, -months, -days, -hours, -minutes, -seconds)
}

// Sleep delays execution
func Sleep(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

// Usleep delays execution in microseconds
func Usleep(microseconds int) {
	time.Sleep(time.Duration(microseconds) * time.Microsecond)
}

// TimeSleep delays execution with floating point seconds
func TimeSleep(seconds float64) {
	duration := time.Duration(seconds * float64(time.Second))
	time.Sleep(duration)
}

// DateTimeFormat formats a DateTime object
type DateTimeFormat struct {
	time.Time
}

// NewDateTimeFormat creates a new DateTimeFormat instance
func NewDateTimeFormat(timestr ...string) (*DateTimeFormat, error) {
	var t time.Time
	var err error

	if len(timestr) > 0 {
		timestamp, parseErr := Strtotime(timestr[0])
		if parseErr != nil {
			return nil, parseErr
		}
		t = time.Unix(timestamp, 0)
	} else {
		t = time.Now()
	}

	return &DateTimeFormat{t}, err
}

// Format formats the DateTime
func (dt *DateTimeFormat) Format(format string) string {
	return formatTime(dt.Time, format)
}

// SetTimestamp sets the DateTime from Unix timestamp
func (dt *DateTimeFormat) SetTimestamp(timestamp int64) {
	dt.Time = time.Unix(timestamp, 0)
}

// GetTimestamp gets the Unix timestamp
func (dt *DateTimeFormat) GetTimestamp() int64 {
	return dt.Unix()
}

// Add adds an interval
func (dt *DateTimeFormat) Add(years, months, days, hours, minutes, seconds int) {
	dt.Time = DateAdd(dt.Time, years, months, days, hours, minutes, seconds)
}

// Sub subtracts an interval
func (dt *DateTimeFormat) Sub(years, months, days, hours, minutes, seconds int) {
	dt.Time = DateSub(dt.Time, years, months, days, hours, minutes, seconds)
}