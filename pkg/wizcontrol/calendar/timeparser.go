package calendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var timeRegexp = regexp.MustCompile(`^(\d{1,2}):(\d{1,2})\s*(am|pm)?$`)

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
