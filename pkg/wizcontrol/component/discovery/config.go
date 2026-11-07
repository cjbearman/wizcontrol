package discovery

import (
	"fmt"
	"time"

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

	var err error
	if c.BroadcastAddress, err = validator.IP(alias.BroadcastAddress, false); err != nil {
		return fmt.Errorf("discovery: invalid broadcast address: %v", err)
	}

	_, err = time.ParseDuration(alias.Interval)
	if err != nil {
		return fmt.Errorf("discovery: invalid interval: %v", err)
	}
	c.Interval = alias.Interval

	_, err = time.ParseDuration(alias.Period)
	if err != nil {
		return fmt.Errorf("discovery: invalid period: %v", err)
	}
	c.Period = alias.Period

	return nil
}
