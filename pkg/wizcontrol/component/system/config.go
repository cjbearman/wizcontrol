package system

import (
	"fmt"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

// ConfigYAML represents the system configuration as defined in the YAML file.
type ConfigYAML struct {
	Timezone         string `yaml:"timezone"`
	EvaluationPeriod string `yaml:"evaluation_period"`
	RecheckPeriod    string `yaml:"recheck_period"`
	WizColorPrefix   string `yaml:"wiz_color_prefix"`
	StateFile        string `yaml:"state_file"`
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

	if err := validator.ValidateStr(alias.EvaluationPeriod, &c.EvaluationPeriod, validator.Duration, false); err != nil {
		return fmt.Errorf("system: invalid evaluation period %q: %v", alias.EvaluationPeriod, err)
	}
	if err := validator.ValidateStr(alias.RecheckPeriod, &c.RecheckPeriod, validator.Duration, false); err != nil {
		return fmt.Errorf("system: invalid recheck period %q: %v", alias.RecheckPeriod, err)
	}

	if err := validator.ValidateStr(alias.WizColorPrefix, &c.WizColorPrefix, validator.ID, false); err != nil {
		return fmt.Errorf("system: invalid wiz color prefix %q: %v", alias.WizColorPrefix, err)
	}

	if c.WizColorPrefix == "" {
		c.WizColorPrefix = color.DefaultWizColorPrefix
	}

	_ = validator.ValidateStr(alias.StateFile, &c.StateFile, validator.Any, false)

	return nil
}
