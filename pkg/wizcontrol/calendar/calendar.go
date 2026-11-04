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

type SceneMode string

const (
	SceneModeAll    SceneMode = "all"
	SceneModeRandom SceneMode = "random"
	SceneModeRotate SceneMode = "rotate"
)

var CalendarDebug = false

var DefaultFrequency = time.Duration(24) * time.Hour

type CalendarTime struct {
	Hour   int
	Minute int
}

type CalendarEntry struct {
	Name       string
	FromDay    int
	ToDay      int
	BeforeTime *CalendarTime
	AfterTime  *CalendarTime
	Scenes     SceneSelector
}

var CalendarEntries = []CalendarEntry{}

var standardDateRegexp = regexp.MustCompile(`^([a-z]+)\s+(\d+)$`)
var holidayDateRegexp = regexp.MustCompile(`^([a-z_]+)\s*(([+-])\s*(\d*))?$`)

func AddCalendarEntry(name, from, to, before, after string, scenesList []string, sceneMode SceneMode, frequency *time.Duration) error {
	if CalendarEntryExists(name) {
		return fmt.Errorf("calendar entry with name %s already exists", name)
	}

	if from == "" {
		return fmt.Errorf("calendar entry %s requires from", name)
	}
	if to == "" {
		return fmt.Errorf("calendar entry %q requires to", name)
	}

	if len(scenesList) == 0 {
		return fmt.Errorf("calendar entry %q requires at least one scene", name)
	}

	for _, scene := range scenesList {
		if exists, missing := scenes.ScenesExist(scene); !exists {
			return fmt.Errorf("calendar entry %q contains undefined scene %q", name, missing)
		}
	}

	frequencyDefined := false
	if frequency != nil {
		frequencyDefined = true
		if *frequency < 1*time.Minute {
			return fmt.Errorf("frequency for calendar entry %q must be at least 1 minute", name)
		}
	}

	if frequency == nil {
		frequency = &DefaultFrequency
	}

	var selector SceneSelector
	switch sceneMode {
	case SceneModeAll, "":
		if frequencyDefined {
			return fmt.Errorf("frequency cannot be defined for scene mode 'all'")
		}
		selector = &AllSceneSelector{scenes: scenesList}
	case SceneModeRandom:
		selector = &RandomSceneSelector{scenes: scenesList, Frequency: *frequency}
	case SceneModeRotate:
		selector = &RotateSceneSelector{scenes: scenesList, Frequency: *frequency}
	default:
		return fmt.Errorf("invalid scene mode %q for calendar entry %q", sceneMode, name)
	}

	entry := CalendarEntry{
		FromDay: DateToDayNumber(from),
		ToDay:   DateToDayNumber(to),
		Scenes:  selector,
		Name:    name,
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

	CalendarEntries = append(CalendarEntries, entry)
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
	logger := common.Logger.GetLogger("calendar")
	result := []string{}

	dayNo, now := GetTodayNumber()
	extCalendarDebugLog("Searching for active calendar scenes for day number %d", dayNo)

	for name := range CalendarEntries {
		extCalendarDebugLog("Found calendar entry: %q", name)
	}

	for _, entry := range CalendarEntries {
		if dayNo >= entry.FromDay && dayNo <= entry.ToDay {
			extCalendarDebugLog("Calendar entry %q matches current day (%d) with from %d, to %d", entry.Name, dayNo, entry.FromDay, entry.ToDay)
			// Scene matches the current day, but how about any time period

			nowHour := now.In(common.Location).Hour()
			nowMinute := now.In(common.Location).Minute()

			compareWith := time.Date(1, 1, 1, nowHour, nowMinute, now.Second(), now.Nanosecond(), common.Location)

			if entry.BeforeTime != nil {
				compareTo := time.Date(1, 1, 1, entry.BeforeTime.Hour, entry.BeforeTime.Minute, 0, 0, common.Location)
				if compareWith.After(compareTo) {
					extCalendarDebugLog("Current time is after the before time for calendar entry %q", entry.Name)
					continue
				}
			}
			if entry.AfterTime != nil {
				compareTo := time.Date(1, 1, 1, entry.AfterTime.Hour, entry.AfterTime.Minute, 0, 0, common.Location)
				if compareWith.Before(compareTo) {
					extCalendarDebugLog("Current time is before the after time for calendar entry %q", entry.Name)
					continue
				}
			}

			sc := entry.Scenes.GetScene(dayNo)
			extCalendarDebugLog("Calendar entry %q is active for the current time, scenes %v will be used", entry.Name, sc)
			result = append(result, sc...)
		} else {
			extCalendarDebugLog("Calendar entry %q does not match current day (%d) with from %d, to %d", entry.Name, dayNo, entry.FromDay, entry.ToDay)
		}
	}

	logger.Infof("Calendar scenes selected: %s", strings.Join(result, ","))

	return result

}

func extCalendarDebugLog(format string, v ...any) {
	if CalendarDebug {
		common.Logger.GetLogger("calendar").Debugf(format, v...)
	}
}

func CalendarEntryExists(name string) bool {
	for _, item := range CalendarEntries {
		if item.Name == name {
			return true
		}
	}
	return false
}
