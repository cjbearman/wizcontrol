// Package config handles configuration of the wizcontrol application
package config

import (
	"fmt"
	"os"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/bus"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/calendar"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/discovery"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/http"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/mqtt"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/system"
	"gopkg.in/yaml.v3"
)

// ConfigYAML represents the configuration as defined in the YAML file.
type ConfigYAML struct {
	Lights    []light.ConfigYAML    `yaml:"lights"`
	Colors    []color.ConfigYAML    `yaml:"colors"`
	Scenes    []scene.ConfigYAML    `yaml:"scenes"`
	Calendars []calendar.ConfigYAML `yaml:"calendar"`
	MQTT      *mqtt.ConfigYAML      `yaml:"mqtt"`
	HTTP      *http.ConfigYAML      `yaml:"http"`
	Discovery discovery.ConfigYAML  `yaml:"discovery"`
	System    system.ConfigYAML     `yaml:"system"`
}

// Config represents the fully built configuration of the wizcontrol application.
type Config struct {
	Lights           light.Lights
	LightGroups      light.LightGroups
	Colors           color.Colors
	WizColors        color.Colors
	Scenes           scene.Scenes
	SceneGroups      scene.SceneGroups
	Calendars        calendar.Calendars
	MQTT             *mqtt.Mqtt
	HTTP             *http.Http
	Discovery        discovery.Discovery
	System           system.System
	AdditionalScenes scene.AdditionalScenes
	Bus              *bus.Bus
}

// Load reads the configuration from the specified YAML file and returns a ConfigYAML instance.
func Load(path string) (*ConfigYAML, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close() // nolint:errcheck

	decoder := yaml.NewDecoder(fp)
	decoder.KnownFields(true) // Reject unknown fields

	cfg := &ConfigYAML{}
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Build constructs a Config instance from the provided ConfigYAML, applying necessary defaults and building components.
func Build(cfg *ConfigYAML) (*Config, error) {

	conf := &Config{}
	var err error

	conf.System = system.Build(cfg.System)

	conf.Lights, conf.LightGroups, err = light.Build(&cfg.Lights)
	if err != nil {
		return nil, fmt.Errorf("error parsing lights: %v", err)
	}

	conf.Colors, conf.WizColors, err = color.Build(&cfg.Colors, conf.System.WizColorPrefix())
	if err != nil {
		return nil, fmt.Errorf("error parsing colors: %v", err)
	}

	conf.Scenes, conf.SceneGroups, err = scene.Build(&cfg.Scenes, conf.Colors, conf.Lights, conf.LightGroups)
	if err != nil {
		return nil, fmt.Errorf("error parsing scenes: %v", err)
	}

	conf.Calendars, err = calendar.Build(cfg.Calendars, conf.Scenes)
	if err != nil {
		return nil, fmt.Errorf("error parsing calendars: %v", err)
	}

	conf.MQTT = mqtt.Build(cfg.MQTT)
	conf.HTTP = http.Build(cfg.HTTP)
	conf.Discovery = discovery.Build(cfg.Discovery)

	if conf.MQTT != nil {
		conf.Bus = bus.NewBus()
	} else {
		// If MQTT is not running, we use a fake bus that discards all messages
		conf.Bus = bus.NewNilBus()
	}

	return conf, nil
}

// GetLightOrGroup returns either a single light or group of lights by name.
// Lights and light groups cannot share a name, and thus this function is deterministic.
func (c *Config) GetLightOrGroup(lightOrGroup string) []*light.Light {
	if l := c.Lights[lightOrGroup]; l != nil {
		return []*light.Light{l}
	}
	if g, found := c.LightGroups[lightOrGroup]; found {
		return g.Lights()
	}
	return nil
}
