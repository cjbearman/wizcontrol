package http

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

/*
http:
  # port on which the HTTP API will listen
  port: 12120

  # To enable HTTPS (TLS) include both cert_chain and key files
  cert_chain: test_certs/wizcontrol.crt
  key: test_certs/wizcontrol.key

  # To establish HTTP basic authentication, uncomment and configure the auth block below
  auth:
    username: "chris"
    password: "changeme"

*/

const DefaultHTTPPort = 12120

type ConfigYAML struct {
	Port      int    `yaml:"port"`
	CertChain string `yaml:"cert_chain"`
	Key       string `yaml:"key"`
	Auth      *struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
}

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
