package homeassistant

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

const DefaultLockTime = "day"
const DefaultHomeAssistantTopicBase = "homeassistant"

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

	if err := validator.ValidateStr(alias.HomeAssistantTopic, &c.HomeAssistantTopic, validator.BaseTopic, false); err != nil {
		return fmt.Errorf("invalid home_assistant_topic: %v", err)
	}
	if c.HomeAssistantTopic == "" {
		c.HomeAssistantTopic = DefaultHomeAssistantTopicBase
	}

	if err := validator.ValidateStr(alias.DefaultLockFor, &c.DefaultLockFor, validator.For, false); err != nil {
		return fmt.Errorf("invalid default_lock_for: %v", err)
	}
	if c.DefaultLockFor == "" {
		c.DefaultLockFor = DefaultLockTime
	}

	return nil
}
