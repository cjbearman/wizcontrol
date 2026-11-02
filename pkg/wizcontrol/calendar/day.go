package calendar

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

// Because our calendar may includes dates of Feb 9, we calculate day numbers using a leap year to ensure correct day-of-year values.
const leapYear = 2024

func GetDayNumber(t time.Time) int {
	return time.Date(leapYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), common.Location).YearDay()
}

func GetTodayNumber() (int, time.Time) {
	now := time.Now().In(common.Location)
	return GetDayNumber(now), now
}

func GetDayNumberByDayAndMonth(day int, month time.Month) int {
	return time.Date(leapYear, month, day, 12, 0, 0, 0, common.Location).YearDay()
}
