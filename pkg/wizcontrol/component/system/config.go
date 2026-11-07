package system

import (
	"fmt"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"gopkg.in/yaml.v3"
)

// ConfigYAML represents the system configuration as defined in the YAML file.
type ConfigYAML struct {
	Timezone         string `yaml:"timezone"`
	EvaluationPeriod string `yaml:"evaluation_period"`
	RecheckPeriod    string `yaml:"recheck_period"`
	WizColorPrefix   string `yaml:"wiz_color_prefix"`
}

// UnmarshalYAML customizes the unmarshalling of the system configuration from YAML.
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

	c.WizColorPrefix = strings.TrimSpace(strings.ToLower(alias.WizColorPrefix))
	if c.WizColorPrefix == "" {
		c.WizColorPrefix = color.DefaultWizColorPrefix
	}

	return nil
}
