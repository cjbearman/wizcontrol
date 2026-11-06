package scene

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

type ElementYAML struct {
	Targets []string `yaml:"targets"`
	Color   string   `yaml:"color"`
}

type ConfigYAML struct {
	Name     string        `yaml:"name"`
	Elements []ElementYAML `yaml:"elements"`
	Group    string        `yaml:"groups"`
}

func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	var err error
	if c.Name, err = validator.RequiredComponentName(alias.Name); err != nil {
		return fmt.Errorf("scene %q, invalid name: %v", alias.Name, err)
	}

	if c.Group, err = validator.ComponentName(alias.Group); err != nil {
		return fmt.Errorf("scene %q, invalid group: %v", alias.Name, err)
	}

	for _, element := range alias.Elements {
		color, err := validator.RequiredComponentName(element.Color)
		if err != nil {
			return fmt.Errorf("Scene %q, invalid element color: %q: %v", c.Name, element.Color, err)
		}
		for _, target := range element.Targets {
			if _, err := validator.RequiredComponentName(target); err != nil {
				return fmt.Errorf("Scene %q, invalid element target: %q: %v", c.Name, target, err)
			}
		}
		c.Elements = append(c.Elements, ElementYAML{
			Targets: element.Targets,
			Color:   color,
		})
	}

	return nil
}
