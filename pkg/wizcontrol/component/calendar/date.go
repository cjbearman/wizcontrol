package calendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

// Regular expressions for parsing standard dates, which is 3 letter month followed by numeric day of month
var standardDateRegexp = regexp.MustCompile(`^([a-z]+)\s+(\d+)$`)

// Regular expressions for parsing holiday dates, which is a holiday name optionally followed by an offset (+/- number of days)
var holidayDateRegexp = regexp.MustCompile(`^([a-z_]+)\s*(([+-])\s*(\d*))?$`)

// Because our calendar may includes dates of Feb 9, we calculate day numbers using a leap year to ensure correct day-of-year values.
const leapYear = 2024

// GetDayNumber returns the day number of the given time, using the leap year for calculation.
func GetDayNumber(t time.Time) int {
	return time.Date(leapYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), common.Location).YearDay()
}

// GetTodayNumber returns the day number and the current time for today, using the leap year for calculation.
func GetTodayNumber() (int, time.Time) {
	now := time.Now().In(common.Location)
	return GetDayNumber(now), now
}

// GetDayNumberByDayAndMonth returns the day number for the specified day and month, using the leap year for calculation.
func GetDayNumberByDayAndMonth(day int, month time.Month) int {
	return time.Date(leapYear, month, day, 12, 0, 0, 0, common.Location).YearDay()
}

// DateToDayNumber takes a date string and returns the day number (respective to leap year)
// or -1 if it is not valid.
// The date string can be in the format "MMM DD" (e.g., "Feb 9") or a holiday name with an optional offset (e.g., "christmas+1").
func DateToDayNumber(date string) int {
	lc := strings.ToLower(date)
	matches := standardDateRegexp.FindStringSubmatch(lc)
	if matches != nil {
		monthStr := matches[1]
		day := matches[2]
		dayInt, err := strconv.Atoi(day)
		if err != nil {
			return -1
		}

		month, err := monthDecode(monthStr)
		if err != nil {
			return -1
		}
		if !dayValid(dayInt, month) {
			return -1
		}

		return GetDayNumberByDayAndMonth(dayInt, month)
	}
	matches = holidayDateRegexp.FindStringSubmatch(lc)
	if matches != nil {
		holiday := matches[1]

		resolveFunction, found := HolidayNames[holiday]
		if !found {
			fmt.Printf("NF %s\n", holiday)
			return -1
		}

		op := matches[3]
		offset, err := strconv.Atoi(matches[4])
		if err != nil && matches[4] != "" {
			return -1
		}

		day := resolveFunction()
		switch op {
		case "+":
			day += offset
			if day > 365 {
				day = 365
			}
		case "-":
			day -= offset
			if day < 1 {
				day = 1
			}
		}

		return day
	}
	return -1
}

// monthDecode converts a month string (full or abbreviated) to a time.Month value. Returns an error if the month is invalid.
func monthDecode(month string) (time.Month, error) {
	switch month {
	case "january", "jan":
		return time.January, nil
	case "february", "feb":
		return time.February, nil
	case "march", "mar":
		return time.March, nil
	case "april", "apr":
		return time.April, nil
	case "may":
		return time.May, nil
	case "june", "jun":
		return time.June, nil
	case "july", "jul":
		return time.July, nil
	case "august", "aug":
		return time.August, nil
	case "september", "sep":
		return time.September, nil
	case "october", "oct":
		return time.October, nil
	case "november", "nov":
		return time.November, nil
	case "december", "dec":
		return time.December, nil
	default:
		return -1, fmt.Errorf("invalid month: %s", month)
	}
}

// dayValid checks if the given day is valid for the specified month.
func dayValid(day int, month time.Month) bool {
	if day < 1 {
		return false
	}
	switch month {
	case time.January, time.March, time.May, time.July, time.August, time.October, time.December:
		return day <= 31
	case time.April, time.June, time.September, time.November:
		return day <= 30
	case time.February:
		return day <= 29
	default:
		return false
	}
}

// DayThisYearFromDayNumber returns the month and day string (e.g., "Jan 2") for the given day number in the current year.
func DayThisYearFromDayNumber(day int) (string, error) {
	if day < 1 || day > 366 {
		return "", fmt.Errorf("invalid day number: %d", day)
	}

	// Get a timestamp representing this day number in our target leap year
	t := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, 0, day-1)

	// Return this formatted with just the month and day
	return t.Format("Jan 2"), nil
}
