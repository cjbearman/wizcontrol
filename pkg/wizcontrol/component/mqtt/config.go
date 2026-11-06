package mqtt

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"
)

type ElementYAML struct {
	Targets []string `yaml:"targets"`
	Color   string   `yaml:"color"`
}

type ConfigYAML struct {
	URL            string `yaml:"url"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Topic          string `yaml:"topic"`
	ControlEnabled *bool  `yaml:"controlEnabled"`
}

func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	var err error

	c.URL = strings.TrimSpace(alias.URL)
	if err = ValidateMQTTURL(c.URL); err != nil {
		return fmt.Errorf("invalid MQTT URL: %v", err)
	}

	c.Username = strings.TrimSpace(alias.Username)
	c.Password = strings.TrimSpace(alias.Password)
	c.Topic = strings.TrimSpace(alias.Topic)

	ce := true
	if alias.ControlEnabled != nil {
		ce = *alias.ControlEnabled
	}
	c.ControlEnabled = &ce

	return nil
}

func ValidateMQTTURL(rawURL string) error {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return err
	}

	// Ensure scheme is present and is a supported MQTT/WebSocket scheme
	validSchemes := map[string]bool{
		"tcp":   true,
		"ssl":   true,
		"tls":   true,
		"mqtt":  true,
		"mqtts": true,
		"ws":    true,
		"wss":   true,
	}

	scheme := strings.ToLower(u.Scheme)
	if !validSchemes[scheme] {
		return fmt.Errorf("unsupported MQTT scheme: %s", u.Scheme)
	}

	// Ensure a host is present
	if u.Host == "" {
		return errors.New("missing host in MQTT URL")
	}

	return nil
}
