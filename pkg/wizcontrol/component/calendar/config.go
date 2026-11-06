package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

type ConfigYAML struct {
	Name      string   `yaml:"name"`
	From      string   `yaml:"from"`
	To        string   `yaml:"to"`
	Before    string   `yaml:"before"`
	After     string   `yaml:"after"`
	Scenes    []string `yaml:"scenes"`
	SceneMode string   `yaml:"mode"`
	Frequency string   `yaml:"frequency"`
}

func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	var err error
	if c.Name, err = validator.RequiredComponentName(alias.Name); err != nil {
		return fmt.Errorf("calendar %q, invalid name: %v", alias.Name, err)
	}

	if DateToDayNumber(alias.From) == -1 {
		return fmt.Errorf("calendar %q, invalid from date: %v", alias.Name, alias.From)
	}
	c.From = alias.From
	if DateToDayNumber(alias.To) == -1 {
		return fmt.Errorf("calendar %q, invalid to date: %v", alias.Name, alias.To)
	}
	c.To = alias.To

	if _, err := OptionalCalendarTime(alias.Before); err != nil {
		return fmt.Errorf("calendar %q, invalid before time: %v", alias.Name, err)
	} else {
		c.Before = alias.Before
	}

	if _, err := OptionalCalendarTime(alias.After); err != nil {
		return fmt.Errorf("calendar %q, invalid after time: %v", alias.Name, err)
	} else {
		c.After = alias.After
	}

	for _, scene := range alias.Scenes {
		if _, err := validator.RequiredComponentName(scene); err != nil {
			return fmt.Errorf("calendar %q, invalid scene: %v", alias.Name, err)
		}
		c.Scenes = append(c.Scenes, scene)
	}

	c.SceneMode = strings.TrimSpace(alias.SceneMode)
	switch c.SceneMode {
	case SceneModeAll, SceneModeRandom, SceneModeRotate:
		// valid mode
	case "":
		// default to "all" if not specified
		c.SceneMode = SceneModeAll
	default:
		return fmt.Errorf("calendar %q, invalid scene mode: %v", alias.Name, c.SceneMode)
	}

	c.Frequency = strings.TrimSpace(alias.Frequency)
	if c.SceneMode == "all" && c.Frequency != "" {
		return fmt.Errorf("Scene %q is set to mode all and must not have a frequency specified", alias.Name)
	}
	if c.SceneMode == SceneModeRandom || c.SceneMode == SceneModeRotate {
		freq, err := time.ParseDuration(alias.Frequency)
		if err != nil {
			return fmt.Errorf("calendar %q, invalid frequency: %v", alias.Name, err)
		}
		if freq <= 0 {
			return fmt.Errorf("calendar %q, frequency invalid: %s", alias.Name, alias.Frequency)
		}
	}
	alias.Frequency = c.Frequency

	return nil
}
