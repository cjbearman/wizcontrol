package common

import (
	"fmt"
	"strings"
	"time"
)

func ForParser(forSpec string) (time.Time, error) {
	forSpec = strings.TrimSpace(strings.ToLower(forSpec))
	// Accept a standard Go duration string, e.g., "1h30m"
	duration, err := time.ParseDuration(forSpec)

	if err == nil {
		return time.Now().In(Location).Add(duration), nil
	}

	// Okay, so that wasn't it, we have a few other suported options
	if forSpec == "day" {
		now := time.Now().In(Location)
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, Location), nil
	}

	if forSpec == "week" {
		now := time.Now().In(Location)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		daysUntilEndOfWeek := 7 - weekday
		endOfWeek := now.AddDate(0, 0, daysUntilEndOfWeek)
		return time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, Location), nil
	}

	if forSpec == "month" {
		now := time.Now().In(Location)
		// Day zero of any next month rationally gives us the last day of the current month,
		// even though we might be setting a month of 13. Works.
		endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, Location)
		return endOfMonth, nil
	}

	return time.Time{}, fmt.Errorf("invalid 'for' duration: %v", err)
}
