package mqtt

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigYAML represents the YAML configuration for the MQTT component.
type ConfigYAML struct {
	URL            string `yaml:"url"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Topic          string `yaml:"topic"`
	ControlEnabled *bool  `yaml:"control_enabled"`
	RetainMessages *bool  `yaml:"retain_messages"`
	Insecure       *bool  `yaml:"insecure"`
	ServerCerts    string `yaml:"server_certs"`
}

// UnmarshalYAML customizes the unmarshalling of the MQTT configuration from YAML. It validates the URL and sets default values for optional fields.
func (c *ConfigYAML) UnmarshalYAML(value *yaml.Node) error {

	type Alias ConfigYAML
	var alias Alias

	if err := value.Decode(&alias); err != nil {
		return err
	}

	// TODO: More validation for Home Assistant fields

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

	rm := true
	if alias.RetainMessages != nil {
		rm = *alias.RetainMessages
	}
	c.RetainMessages = &rm

	inc := false
	if alias.Insecure != nil {
		inc = *alias.Insecure
	}
	c.Insecure = &inc
	c.ServerCerts = alias.ServerCerts

	return nil
}

// ValidateMQTTURL checks if the provided MQTT URL is valid, ensuring it has a supported scheme and a host.
// TODO: We probably need more work in the MQTT main component for supporting TLS based transports.
// Also need to check that the MQTT implementation accepts all of these URL patterns
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
