package http

import (
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
	"gopkg.in/yaml.v3"
)

// DefaultHTTPPort is the default HTTP port for the HTTP component
const DefaultHTTPPort = 12120

// ConfigYAML is the YAML representation of an HTTP configuration
type ConfigYAML struct {
	Port      int    `yaml:"port"`
	CertChain string `yaml:"cert_chain"`
	Key       string `yaml:"key"`
	Auth      *struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ConfigYAML
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	c.Port = alias.Port
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("invalid http port: %d", c.Port)
	}
	if c.Port == 0 {
		c.Port = DefaultHTTPPort
	}

	if err := validator.ValidateStr(alias.CertChain, &c.CertChain, validator.Any, false); err != nil {
		return fmt.Errorf("invalid http cert_chain: %v", err)
	}
	if err := validator.ValidateStr(alias.Key, &c.Key, validator.Any, false); err != nil {
		return fmt.Errorf("invalid http key: %v", err)
	}

	if c.CertChain != "" && c.Key == "" {
		return fmt.Errorf("http key must be provided if cert_chain is specified")
	}
	if c.Key != "" && c.CertChain == "" {
		return fmt.Errorf("http cert_chain must be provided if key is specified")
	}

	if alias.Auth != nil {
		c.Auth = alias.Auth

		if err := validator.ValidateStr(alias.Auth.Username, &c.Auth.Username, validator.Any, true); err != nil {
			return fmt.Errorf("invalid http auth username: %v", err)
		}
		if err := validator.ValidateStr(alias.Auth.Password, &c.Auth.Password, validator.Any, true); err != nil {
			return fmt.Errorf("invalid http auth password: %v", err)
		}
	}

	return nil
}
