package light

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

type ConfigYAML struct {
	Name   string   `yaml:"name"`
	IP     string   `yaml:"ip"`
	MAC    string   `yaml:"mac"`
	Groups []string `yaml:"groups"`
}

func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	var err error
	if c.Name, err = validator.RequiredComponentName(alias.Name); err != nil {
		return fmt.Errorf("light %q, invalid name: %v", alias.Name, err)
	}

	if c.IP, err = validator.IP(alias.IP, false); err != nil {
		return fmt.Errorf("light %q ip address %q invalid: %w", alias.Name, c.IP, err)
	}

	if c.MAC, err = validator.MAC(alias.MAC, false); err != nil {
		return fmt.Errorf("light %q mac address %q invalid: %w", alias.Name, c.MAC, err)
	}

	if c.IP == "" && c.MAC == "" {
		return fmt.Errorf("light %q requires ip or mac", alias.Name)
	}
	if c.IP != "" && c.MAC != "" {
		return fmt.Errorf("light %q cannot have both ip and mac", alias.Name)
	}

	c.Groups = []string{}
	for _, group := range alias.Groups {
		if group, err = validator.RequiredComponentName(group); err != nil {
			return fmt.Errorf("light %q group %q: %w", c.Name, group, err)
		}
		c.Groups = append(c.Groups, group)
	}

	return nil
}
