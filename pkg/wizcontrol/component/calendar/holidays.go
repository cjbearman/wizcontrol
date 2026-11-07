package calendar

import (
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/aa"
	"github.com/rickar/cal/v2/us"
)

var HolidayNames = map[string]func() int{
	"easter_sunday":             func() int { return easterSunday() },
	"good_friday":               func() int { return goodFriday() },
	"maundy_thursday":           func() int { return maundyThursday() },
	"easter_monday":             func() int { return easterMonday() },
	"international_workers_day": func() int { return GetDayNumberByDayAndMonth(1, time.May) },
	"christmas_day":             func() int { return GetDayNumberByDayAndMonth(25, time.December) },
	"new_years_day":             func() int { return GetDayNumberByDayAndMonth(1, time.January) },
	"boxing_day":                func() int { return GetDayNumberByDayAndMonth(26, time.December) },
	"martin_luther_king_day":    func() int { return mlkDay() },
	"presidents_day":            func() int { return presidentsDay() },
	"memorial_day":              func() int { return memorialDay() },
	"labor_day":                 func() int { return laborDay() },
	"columbus_day":              func() int { return columbusDay() },
	"veterans_day":              func() int { return GetDayNumberByDayAndMonth(11, time.November) },
	"thanksgiving_day":          func() int { return thanksgivingDay() },
	"independence_day":          func() int { return GetDayNumberByDayAndMonth(4, time.July) },
	"halloween":                 func() int { return GetDayNumberByDayAndMonth(31, time.October) },
	"st_patricks_day":           func() int { return GetDayNumberByDayAndMonth(17, time.March) },
	"juneteenth":                func() int { return GetDayNumberByDayAndMonth(19, time.June) },
}

func easterSunday() int {
	t := cal.CalcEasterOffset(aa.Easter, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func goodFriday() int {
	t := cal.CalcEasterOffset(aa.GoodFriday, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func maundyThursday() int {
	t := cal.CalcEasterOffset(aa.MaundyThursday, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func easterMonday() int {
	t := cal.CalcEasterOffset(aa.EasterMonday, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func mlkDay() int {
	t := cal.CalcWeekdayFrom(us.MlkDay, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func presidentsDay() int {
	t := cal.CalcWeekdayFrom(us.PresidentsDay, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func memorialDay() int {
	t := cal.CalcWeekdayFrom(us.MemorialDay, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func laborDay() int {
	t := cal.CalcWeekdayFrom(us.LaborDay, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func columbusDay() int {
	t := cal.CalcWeekdayFrom(us.ColumbusDay, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}
func thanksgivingDay() int {
	t := cal.CalcWeekdayFrom(us.ThanksgivingDay, time.Now().In(common.Location).Year())
	return GetDayNumberByDayAndMonth(t.Day(), t.Month())
}

func ListHolidays() string {
	list := []string{}
	for holiday := range HolidayNames {
		list = append(list, holiday)
	}
	return strings.Join(list, ", ")
}
