package calendar

import (
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

const (
	SceneModeAll    = "all"
	SceneModeRandom = "random"
	SceneModeRotate = "rotate"
)

type Calendars []*Calendar

type Calendar struct {
	name          string
	from, to      int
	before, after *CalendarTime
	scenes        []*scene.Scene
	selector      SceneSelector
}

func Build(cfg []ConfigYAML, scenes scene.Scenes) (Calendars, error) {
	var calendars Calendars

	for _, c := range cfg {
		for _, existing := range calendars {
			if existing.name == c.Name {
				return nil, fmt.Errorf("calendar %q is defined more than once", c.Name)
			}
		}

		sceneObjs := []*scene.Scene{}
		for _, scene := range c.Scenes {
			sc := scenes.GetByName(scene)
			if sc == nil {
				return nil, fmt.Errorf("scene %q for calendar %q not found", scene, c.Name)
			}
			sceneObjs = append(sceneObjs, sc)
		}

		calendar := &Calendar{
			name:   c.Name,
			from:   DateToDayNumber(c.From),
			to:     DateToDayNumber(c.To),
			before: func() *CalendarTime { t, _ := OptionalCalendarTime(c.Before); return t }(),
			after:  func() *CalendarTime { t, _ := OptionalCalendarTime(c.After); return t }(),
			scenes: sceneObjs,
			selector: func() SceneSelector {
				switch c.SceneMode {
				case SceneModeAll:
					return &AllSceneSelector{scenes: sceneObjs}
				case SceneModeRandom:
					return &RandomSceneSelector{scenes: sceneObjs, Frequency: func() time.Duration { d, _ := time.ParseDuration(c.Frequency); return d }()}
				case SceneModeRotate:
					return &RotateSceneSelector{scenes: sceneObjs, Frequency: func() time.Duration { d, _ := time.ParseDuration(c.Frequency); return d }()}
				default:
					// Should not be possible due to validation in UnmarshalYAML
					panic(fmt.Sprintf("invalid scene mode: %v", c.SceneMode))
				}
			}(),
		}
		calendars = append(calendars, calendar)
	}

	return calendars, nil
}

func (c *Calendar) String() string {
	return fmt.Sprintf("Calendar{name: %q, from: %d, to: %d, before: %v, after: %v, scenes: %v}", c.name, c.from, c.to, c.before, c.after, c.scenes)
}

func (c *Calendar) Name() string {
	return c.name
}

func (c *Calendar) From() int {
	return c.from
}

func (c *Calendar) To() int {
	return c.to
}

func (c *Calendar) Before() *CalendarTime {
	return c.before
}

func (c *Calendar) After() *CalendarTime {
	return c.after
}

func (c *Calendar) Scenes() []*scene.Scene {
	return c.scenes
}

func (c Calendars) GetByName(name string) *Calendar {
	for _, item := range c {
		if item.name == name {
			return item
		}
	}
	return nil
}

func (c Calendars) GetCalendarScenes(scenes scene.Scenes) []*scene.Scene {
	logger := common.Logger.GetLogger("calendar")
	result := []*scene.Scene{}

	dayNo, now := GetTodayNumber()

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
		} else {
		}
	}

	selectedSceneName := []string{}
	for _, s := range result {
		selectedSceneName = append(selectedSceneName, s.GetName())
	}

	logger.Infof("Calendar scenes selected: %v", selectedSceneName)

	return result
}
