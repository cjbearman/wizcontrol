package calendar

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

type ConfigYAML struct {
	Name   string   `yaml:"name"`
	From   string   `yaml:"from"`
	To     string   `yaml:"to"`
	Before string   `yaml:"before"`
	After  string   `yaml:"after"`
	Scenes []string `yaml:"scenes"`
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

	return nil
}
