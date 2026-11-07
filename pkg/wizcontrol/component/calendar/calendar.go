package calendar

import (
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

const (
	SceneModeAll    = "all"
	SceneModeRandom = "random"
	SceneModeRotate = "rotate"
)

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
