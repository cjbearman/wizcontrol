package light

import (
	"fmt"
	"regexp"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

var IDReg = regexp.MustCompile(`^[a-z0-9\-_]+$`)

// ConfigYAML is the YAML representation of a light configuration
type ConfigYAML struct {
	Name      string   `yaml:"name"`
	ID        string   `yaml:"id"`
	IP        string   `yaml:"ip"`
	MAC       string   `yaml:"mac"`
	Groups    []string `yaml:"groups"`
	HALockFor string   `yaml:"ha_lock_for"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ConfigYAML
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	if err := validator.ValidateStr(alias.Name, &c.Name, validator.Any, true); err != nil {
		return fmt.Errorf("light %q, field name: %w", alias.Name, err)
	}

	if err := validator.ValidateStr(alias.ID, &c.ID, validator.ID, true); err != nil {
		return fmt.Errorf("light %q, field id: %w", alias.Name, err)
	}

	if err := validator.ValidateStr(alias.IP, &c.IP, validator.IP, false); err != nil {
		return fmt.Errorf("light %q, field ip: %w", alias.Name, err)
	}

	if err := validator.ValidateStr(alias.MAC, &c.MAC, validator.MAC, false); err != nil {
		return fmt.Errorf("light %q, field mac: %w", alias.Name, err)
	}

	c.Groups = []string{}
	for _, group := range alias.Groups {
		var validatedGroup string
		if err := validator.ValidateStr(group, &validatedGroup, validator.ID, true); err != nil {
			return fmt.Errorf("light %q group %q: %w", c.Name, group, err)
		}
		c.Groups = append(c.Groups, validatedGroup)
	}

	if err := validator.ValidateStr(alias.HALockFor, &c.HALockFor, validator.For, false); err != nil {
		return fmt.Errorf("light %q, field ha_lock_for: %w", alias.Name, err)
	}

	return nil
}
