package calendar

import (
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

const (
	// SceneModeAll indicates that all scenes in a calendar should be used
	SceneModeAll = "all"

	// SceneModeRandom indicates that a single scene should be selected at random on the specified frequency
	SceneModeRandom = "random"

	// SceneModeRotate indicates that scenes should be rotated based on the specified frequency
	SceneModeRotate = "rotate"
)

// Calendar represents a calendar with a name, date range, optional before and after times, associated scenes, and a scene selector.
type Calendar struct {
	name                     string
	from, to                 int
	originalFrom, originalTo string
	before, after            *CalendarTime
	scenes                   []*scene.Scene
	selector                 SceneSelector
}

// Build constructs a list of Calendar objects from the provided configuration and available scenes.
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
			sc := scenes[scene]
			if sc == nil {
				return nil, fmt.Errorf("scene %q for calendar %q not found", scene, c.Name)
			}
			sceneObjs = append(sceneObjs, sc)
		}

		calendar := &Calendar{
			name:         c.Name,
			from:         DateToDayNumber(c.From),
			to:           DateToDayNumber(c.To),
			originalFrom: c.From,
			originalTo:   c.To,
			before:       func() *CalendarTime { t, _ := OptionalCalendarTime(c.Before); return t }(),
			after:        func() *CalendarTime { t, _ := OptionalCalendarTime(c.After); return t }(),
			scenes:       sceneObjs,
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

// String returns a string representation of the Calendar object.
func (c *Calendar) String() string {
	return fmt.Sprintf("Calendar{name: %q, from: %d, to: %d, before: %v, after: %v, scenes: %v}", c.name, c.from, c.to, c.before, c.after, c.scenes)
}

// Name returns the name of the Calendar.
func (c *Calendar) Name() string {
	return c.name
}

// From returns the starting day number of the Calendar.
func (c *Calendar) From() int {
	return c.from
}

// To returns the ending day number of the Calendar.
func (c *Calendar) To() int {
	return c.to
}

// Before returns the optional before time of the Calendar.
func (c *Calendar) Before() *CalendarTime {
	return c.before
}

// After returns the optional after time of the Calendar.
func (c *Calendar) After() *CalendarTime {
	return c.after
}

// Scenes returns the list of scenes associated with the Calendar.
func (c *Calendar) Scenes() []*scene.Scene {
	return c.scenes
}

// OriginalFrom returns the original from date string of the Calendar.
func (c *Calendar) OriginalFrom() string {
	return c.originalFrom
}

// OriginalTo returns the original to date string of the Calendar.
func (c *Calendar) OriginalTo() string {
	return c.originalTo
}

// Selector returns the scene selector of the Calendar.
func (c *Calendar) Selector() SceneSelector {
	return c.selector
}
