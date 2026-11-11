package scene

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

// ElementYAML represents the configuration for an individual element within a scene, including its targets and color.
type ElementYAML struct {
	Targets []string `yaml:"targets"`
	Color   string   `yaml:"color"`
}

// ConfigYAML represents the configuration for a scene, including its name, elements, and group.
type ConfigYAML struct {
	Name     string        `yaml:"name"`
	Elements []ElementYAML `yaml:"elements"`
	Group    string        `yaml:"group"`
}

// UnmarshalYAML custom unmarshals a ConfigYAML from a YAML node, validating its fields.
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	if err := validator.ValidateStr(alias.Name, &c.Name, validator.ID, true); err != nil {
		return fmt.Errorf("invalid scene name: %v", err)
	}

	if err := validator.ValidateStr(alias.Group, &c.Group, validator.ID, false); err != nil {
		return fmt.Errorf("invalid scene group: %v", err)
	}

	for _, element := range alias.Elements {

		var validatedColorName string
		if err := validator.ValidateStr(element.Color, &validatedColorName, validator.ID, true); err != nil {
			return fmt.Errorf("Scene %q, invalid element color: %q: %v", alias.Name, element.Color, err)
		}

		var validatedTargets []string
		for _, target := range element.Targets {

			var validatedTarget string
			if err := validator.ValidateStr(target, &validatedTarget, validator.ID, true); err != nil {
				return fmt.Errorf("Scene %q, invalid element target: %q: %v", alias.Name, target, err)
			}
			validatedTargets = append(validatedTargets, validatedTarget)

		}
		c.Elements = append(c.Elements, ElementYAML{
			Targets: validatedTargets,
			Color:   validatedColorName,
		})
	}

	return nil
}
