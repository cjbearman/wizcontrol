package http

import (
	"fmt"

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

	c.CertChain = alias.CertChain
	c.Key = alias.Key
	if c.CertChain != "" && c.Key == "" {
		return fmt.Errorf("http key must be provided if cert_chain is specified")
	}
	if c.Key != "" && c.CertChain == "" {
		return fmt.Errorf("http cert_chain must be provided if key is specified")
	}

	c.Auth = alias.Auth
	if c.Auth != nil {
		if c.Auth.Username == "" {
			return fmt.Errorf("http auth username must be provided if auth is specified")
		}
		if c.Auth.Password == "" {
			return fmt.Errorf("http auth password must be provided if auth is specified")
		}
	}

	return nil
}
