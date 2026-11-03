package calendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

var CalendarDebug = false

type CalendarTime struct {
	Hour   int
	Minute int
}

type CalendarEntry struct {
	FromDay    int
	ToDay      int
	BeforeTime *CalendarTime
	AfterTime  *CalendarTime
	Scene      SceneSelector
}

var CalendarEntries = make(map[string]CalendarEntry)

var standardDateRegexp = regexp.MustCompile(`^([a-z]+)\s+(\d+)$`)
var holidayDateRegexp = regexp.MustCompile(`^([a-z_]+)\s*(([+-])\s*(\d*))?$`)

func AddCalendarEntry(name, from, to, before, after, scene string, random, rotate []string) error {
	if _, exists := CalendarEntries[name]; exists {
		return fmt.Errorf("calendar entry with name %s already exists", name)
	}

	if from == "" {
		return fmt.Errorf("calendar entry %s requires from", name)
	}
	if to == "" {
		return fmt.Errorf("calendar entry %q requires to", name)
	}

	// Must have exactly one of scene, random or rotate specified
	var selector SceneSelector
	count := 0
	if scene != "" {
		count++
		if exists, missing := scenes.ScenesExist(scene); !exists {
			return fmt.Errorf("calendar scene entry %q references unknown scene %q", name, missing)
		}
		selector = &FixedSceneSelector{scene: scene}

	}

	if len(random) > 0 {
		count++
		if exists, missing := scenes.ScenesExist(random...); !exists {
			return fmt.Errorf("calendar random entry %q references unknown scene %q", name, missing)
		}
		selector = &RandomSceneSelector{scenes: random}
	}
	if len(rotate) > 0 {
		count++
		if exists, missing := scenes.ScenesExist(rotate...); !exists {
			return fmt.Errorf("calendar rotate entry %q references unknown scene %q", name, missing)
		}
		selector = &RotateSceneSelector{scenes: rotate}
	}

	if count != 1 {
		return fmt.Errorf("calendar entry %q must have exactly one of scene, random or rotate specified", name)
	}

	entry := CalendarEntry{
		FromDay: DateToDayNumber(from),
		ToDay:   DateToDayNumber(to),
		Scene:   selector,
	}

	if entry.FromDay == -1 {
		return fmt.Errorf("invalid from date: %s", from)
	}

	if entry.ToDay == -1 {
		return fmt.Errorf("invalid to date: %s", to)
	}

	if before != "" {
		beforeTime, err := ParseTime(before)
		if err != nil {
			return fmt.Errorf("invalid before time: %s", before)
		}
		entry.BeforeTime = &beforeTime
	}
	if after != "" {
		afterTime, err := ParseTime(after)
		if err != nil {
			return fmt.Errorf("invalid after time: %s", after)
		}
		entry.AfterTime = &afterTime
	}

	CalendarEntries[name] = entry
	return nil
}

// DateToDayNumber takes a date string and returns the day number (respective to leap year)
// or -1 if it is not valid
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

func GetCalendarScenes() []string {
	result := []string{}

	// The default scene is always returned as the first element of the active scenes
	// because it must always be active at lowest priority
	result = append(result, "default")

	dayNo, now := GetTodayNumber()
	extCalendarDebugLog("Searching for active calendar scenes for day number %d", dayNo)
	for calendarName, entry := range CalendarEntries {
		if dayNo >= entry.FromDay && dayNo <= entry.ToDay {
			extCalendarDebugLog("Calendar entry %q matches current day (%d) with from %d, to %d", calendarName, dayNo, entry.FromDay, entry.ToDay)
			// Scene matches the current day, but how about any time period

			nowHour := now.In(common.Location).Hour()
			nowMinute := now.In(common.Location).Minute()

			compareWith := time.Date(1, 1, 1, nowHour, nowMinute, now.Second(), now.Nanosecond(), common.Location)

			if entry.BeforeTime != nil {
				compareTo := time.Date(1, 1, 1, entry.BeforeTime.Hour, entry.BeforeTime.Minute, 0, 0, common.Location)
				if compareWith.After(compareTo) {
					extCalendarDebugLog("Current time is after the before time for calendar entry %q", calendarName)
					continue
				}
			}
			if entry.AfterTime != nil {
				compareTo := time.Date(1, 1, 1, entry.AfterTime.Hour, entry.AfterTime.Minute, 0, 0, common.Location)
				if compareWith.Before(compareTo) {
					extCalendarDebugLog("Current time is before the after time for calendar entry %q", calendarName)
					continue
				}
			}

			sc := entry.Scene.GetScene(dayNo)
			extCalendarDebugLog("Calendar entry %q is active for the current time, scene %q will be used", calendarName, sc)
			result = append(result, sc)
		} else {
			extCalendarDebugLog("Calendar entry %q does not match current day (%d) with from %d, to %d", calendarName, dayNo, entry.FromDay, entry.ToDay)
		}
	}

	return result

}

func extCalendarDebugLog(format string, v ...any) {
	if CalendarDebug {
		common.Logger.GetLogger("calendar").Debugf(format, v...)
	}
}
