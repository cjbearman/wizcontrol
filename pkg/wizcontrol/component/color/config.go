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

	var err error
	if c.Name, err = validator.RequiredComponentName(alias.Name); err != nil {
		return fmt.Errorf("color %q, invalid name: %v", alias.Name, err)
	}

	if c.Color, err = validator.RequiredColorString(alias.Color); err != nil {
		return fmt.Errorf("color %q: %v", alias.Name, err)
	}

	return nil
}
