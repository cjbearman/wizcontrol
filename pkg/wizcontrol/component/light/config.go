package light

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

var UIDReg = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ConfigYAML is the YAML representation of a light configuration
type ConfigYAML struct {
	Name      string   `yaml:"name"`
	IP        string   `yaml:"ip"`
	MAC       string   `yaml:"mac"`
	Groups    []string `yaml:"groups"`
	HALockFor string   `yaml:"ha_lock_for"`
	HAUID     string   `yaml:"ha_uid"`
	MQTTID    string   `yaml:"mqtt_id"`
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
		return fmt.Errorf("light %q, invalid name: %v", alias.Name, err)
	}

	if alias.HAUID != "" {
		if !UIDReg.MatchString(alias.HAUID) {
			return fmt.Errorf("light %q ha_uid %q invalid, must contains alphanumerics, _ and - only", alias.Name, alias.HAUID)
		}
		c.HAUID = alias.HAUID
	} else {
		// We don't have a user provided uid, so we'll use the light name and hash it to create a unique identifier
		hash := sha256.Sum256([]byte(c.Name))
		c.HAUID = fmt.Sprintf("%x", hash)
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

	if alias.HALockFor != "" {
		if _, err = common.ForParser(alias.HALockFor); err != nil {
			return fmt.Errorf("light %q ha_lock_for %q invalid: %w", alias.Name, alias.HALockFor, err)
		}
	}
	c.HALockFor = alias.HALockFor

	if alias.MQTTID != "" {

		if strings.ContainsAny(alias.MQTTID, "/#$+ ") {
			return fmt.Errorf("light %q mqtt_id %q invalid, must not contain /#$+ or spaces", alias.Name, alias.MQTTID)
		}
		c.MQTTID = alias.MQTTID
	} else {
		c.MQTTID = makeMqttId(c.Name)
	}

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
