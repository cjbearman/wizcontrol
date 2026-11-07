package calendar

import (
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

type Calendars []*Calendar

func (c Calendars) GetByName(name string) *Calendar {
	for _, item := range c {
		if item.name == name {
			return item
		}
	}
	return nil
}

func (c Calendars) GetCalendarScenes(scenes scene.Scenes) ([]*scene.Scene, []string) {
	logger := common.Logger.GetLogger("calendar")
	result := []*scene.Scene{}

	dayNo, now := GetTodayNumber()
	selectedCalendars := []string{}

	for _, entry := range c {
		if dayNo >= entry.from && dayNo <= entry.to {
			// Scene matches the current day, but how about any time period

			nowHour := now.In(common.Location).Hour()
			nowMinute := now.In(common.Location).Minute()

			compareWith := time.Date(1, 1, 1, nowHour, nowMinute, now.Second(), now.Nanosecond(), common.Location)

			if entry.before != nil {
				compareTo := time.Date(1, 1, 1, entry.before.Hour, entry.before.Minute, 0, 0, common.Location)
				if compareWith.After(compareTo) {
					continue
				}
			}
			if entry.after != nil {
				compareTo := time.Date(1, 1, 1, entry.after.Hour, entry.after.Minute, 0, 0, common.Location)
				if compareWith.Before(compareTo) {
					continue
				}
			}

			sc := entry.selector.GetScene(dayNo)
			result = append(result, sc...)
			selectedCalendars = append(selectedCalendars, entry.name)
		} else {
		}
	}

	selectedSceneName := []string{}
	for _, s := range result {
		selectedSceneName = append(selectedSceneName, s.Name())
	}

	logger.Infof("Calendar scenes selected: %v", selectedSceneName)

	return result, selectedCalendars
}
