package light

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
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

	var err error
	c.Name = strings.TrimSpace(alias.Name)

	if alias.ID == "" {
		return fmt.Errorf("light %q must have id set", alias.Name)
	}
	if !IDReg.MatchString(alias.ID) {
		return fmt.Errorf("light %q id %q invalid, must contain only lowercase letters, numbers, - and _", alias.Name, alias.ID)
	}
	c.ID = alias.ID

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

	if alias.HALockFor != "" {
		if _, err = common.ForParser(alias.HALockFor); err != nil {
			return fmt.Errorf("light %q ha_lock_for %q invalid: %w", alias.Name, alias.HALockFor, err)
		}
	}
	c.HALockFor = alias.HALockFor

	return nil
}

func makeMqttId(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "#", "_")
	name = strings.ReplaceAll(name, "$", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "+", "_")
	return name
}
