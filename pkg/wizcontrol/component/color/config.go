package color

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

// ConfigYAML is the YAML representation of a color configuration
type ConfigYAML struct {
	Name  string `yaml:"name"`
	Color string `yaml:"color"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ConfigYAML
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	if err := validator.ValidateStr(alias.Name, &c.Name, validator.ID, true); err != nil {
		return fmt.Errorf("color %q, invalid name: %v", alias.Name, err)
	}

	if err := validator.ValidateStr(alias.Name, &c.Name, validator.ID, true); err != nil {
		return fmt.Errorf("color %q, invalid name: %v", alias.Name, err)
	}

	if err := validator.ValidateStr(alias.Color, &c.Color, validator.ColorValue, true); err != nil {
		return fmt.Errorf("color %q, invalid color: %v", alias.Name, err)
	}

	return nil
}
