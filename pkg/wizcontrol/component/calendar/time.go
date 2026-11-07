package calendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CalendarTime represents a specific time of day with hour and minute components.
type CalendarTime struct {
	Hour   int
	Minute int
}

var timeRegexp = regexp.MustCompile(`^(\d{1,2}):(\d{1,2})\s*(am|pm)?$`)

// ParseTime parses a time string in the format "HH:MM am/pm" and returns a CalendarTime.
func ParseTime(timeStr string) (CalendarTime, error) {
	matches := timeRegexp.FindStringSubmatch(strings.TrimSpace(strings.ToLower(timeStr)))

	if matches == nil {
		return CalendarTime{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])

	if matches[3] == "pm" && hour < 12 {
		hour += 12
	}
	if matches[3] == "am" {
		if hour == 0 {
			return CalendarTime{}, fmt.Errorf("time %q is invalid", timeStr)
		}
		if hour == 12 {
			hour = 0
		}
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return CalendarTime{}, fmt.Errorf("time %q is invalid", timeStr)
	}

	return CalendarTime{Hour: hour, Minute: minute}, nil
}

// OptionalCalendarTime parses a time string and returns a pointer to CalendarTime.
// If the time string is empty, it returns nil without an error.
func OptionalCalendarTime(timeStr string) (*CalendarTime, error) {
	timeStr = strings.TrimSpace(strings.ToLower(timeStr))
	if timeStr == "" {
		return nil, nil
	}
	ct, err := ParseTime(timeStr)
	if err != nil {
		return nil, err
	}
	return &ct, nil
}
