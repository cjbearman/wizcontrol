package homeassistant

import (
	"fmt"
	"strings"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"gopkg.in/yaml.v3"
)

const DefaultLockTime = "day"

// ConfigYAML represents the YAML configuration for the MQTT component.
type ConfigYAML struct {
	HomeAssistantTopic string `yaml:"home_assistant_topic"`
	DefaultLockFor     string `yaml:"default_lock_for"`
}

// UnmarshalYAML customizes the unmarshalling of the MQTT configuration from YAML. It validates the URL and sets default values for optional fields.
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	// TODO: More validation for Home Assistant fields

	c.HomeAssistantTopic = strings.TrimSpace(alias.HomeAssistantTopic)

	c.DefaultLockFor = alias.DefaultLockFor
	if c.DefaultLockFor == "" {
		c.DefaultLockFor = DefaultLockTime
	}
	if _, err := common.ForParser(c.DefaultLockFor); err != nil {
		return fmt.Errorf("invalid default_lock_time in MQTT configuration: %v", err)
	}

	return nil
}
