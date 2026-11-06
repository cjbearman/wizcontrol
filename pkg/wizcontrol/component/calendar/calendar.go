package calendar

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

type Calendars map[string]*Calendar

type Calendar struct {
	name          string
	from, to      int
	before, after *CalendarTime
	scenes        []string
}

func Build(cfg []ConfigYAML, scenes scene.Scenes) (Calendars, error) {
	calendars := make(Calendars)

	for _, c := range cfg {
		if _, found := calendars[c.Name]; found {
			return nil, fmt.Errorf("calendar %q is defined more than once", c.Name)
		}

		sceneNames := []string{}
		for _, scene := range c.Scenes {
			if scenes.GetByName(scene) == nil {
				return nil, fmt.Errorf("scene %q for calendar %q not found", scene, c.Name)
			}
			sceneNames = append(sceneNames, scene)
		}

		calendars[c.Name] = &Calendar{
			name:   c.Name,
			from:   DateToDayNumber(c.From),
			to:     DateToDayNumber(c.To),
			before: func() *CalendarTime { t, _ := OptionalCalendarTime(c.Before); return t }(),
			after:  func() *CalendarTime { t, _ := OptionalCalendarTime(c.After); return t }(),
			scenes: sceneNames,
		}
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

func (c *Calendar) Scenes() []string {
	return c.scenes
}

func (c Calendars) GetByName(name string) *Calendar {
	return c[name]
}
