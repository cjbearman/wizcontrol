package discovery

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

// ConfigYAML is the YAML representation of a discovery configuration
type ConfigYAML struct {
	BroadcastAddress string `yaml:"broadcast_address"`
	Interval         string `yaml:"interval"`
	Period           string `yaml:"period"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ConfigYAML
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	if err := validator.ValidateStr(alias.BroadcastAddress, &c.BroadcastAddress, validator.IP, false); err != nil {
		return fmt.Errorf("discovery: invalid broadcast address: %v", err)
	}

	if err := validator.ValidateStr(alias.Interval, &c.Interval, validator.Duration, false); err != nil {
		return fmt.Errorf("discovery: invalid interval: %v", err)
	}

	if err := validator.ValidateStr(alias.Period, &c.Period, validator.Duration, false); err != nil {
		return fmt.Errorf("discovery: invalid period: %v", err)
	}

	return nil
}
