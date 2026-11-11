package calendar

import (
	"fmt"
	"strings"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

// ConfigYAML represents the configuration for a calendar as defined in the YAML file.
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

// UnmarshalYAML custom unmarshals the ConfigYAML from the YAML node, performing validation on the fields.
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	if err := validator.ValidateStr(alias.Name, &c.Name, validator.ValidateID, true); err != nil {
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
		var validatedScene string
		if err := validator.ValidateStr(scene, &validatedScene, validator.ValidateID, true); err != nil {
			return fmt.Errorf("calendar %q, invalid scene: %v", alias.Name, err)
		}
		c.Scenes = append(c.Scenes, validatedScene)
	}

	c.SceneMode = strings.TrimSpace(alias.SceneMode)
	switch c.SceneMode {
	case "", SceneModeAll:
		c.SceneMode = SceneModeAll
		if strings.TrimSpace(alias.Frequency) != "" {
			return fmt.Errorf("calendar %q, mode all must not have a frequency specified", alias.Name)
		}
	case SceneModeRandom, SceneModeRotate:
		if err := validator.ValidateStr(alias.Frequency, &c.Frequency, validator.ValidateDuration, true); err != nil {
			return fmt.Errorf("calendar %q, frequency: %v", alias.Name, err)
		}
	default:
		return fmt.Errorf("calendar %q, invalid scene mode: %v", alias.Name, c.SceneMode)
	}

	return nil
}
