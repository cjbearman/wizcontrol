package system

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultHTTPPort = 12120

type ConfigYAML struct {
	Timezone         string `yaml:"timezone"`
	EvaluationPeriod string `yaml:"evaluation_period"`
	RecheckPeriod    string `yaml:"recheck_period"`
}

func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	c.Timezone = strings.TrimSpace(alias.Timezone)
	if c.Timezone != "" {
		_, err := time.LoadLocation(c.Timezone)
		if err != nil {
			return fmt.Errorf("system: invalid timezone %q", c.Timezone)
		}
	}

	c.EvaluationPeriod = strings.TrimSpace(alias.EvaluationPeriod)
	if _, err := time.ParseDuration(c.EvaluationPeriod); err != nil {
		return fmt.Errorf("system: invalid evaluation period %q", c.EvaluationPeriod)
	}
	c.RecheckPeriod = strings.TrimSpace(alias.RecheckPeriod)
	if _, err := time.ParseDuration(c.RecheckPeriod); err != nil {
		return fmt.Errorf("system: invalid recheck period %q", c.RecheckPeriod)
	}

	return nil
}
