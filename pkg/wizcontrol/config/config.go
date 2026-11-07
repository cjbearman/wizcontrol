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

type Config struct {
	Lights           light.Lights
	LightGroups      light.LightGroups
	Colors           color.Colors
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

func Build(cfg *ConfigYAML) (*Config, error) {

	conf := &Config{}
	var err error

	conf.Lights, conf.LightGroups, err = light.Build(&cfg.Lights)
	if err != nil {
		return nil, fmt.Errorf("error parsing lights: %v", err)
	}

	conf.Colors, err = color.Build(&cfg.Colors)
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
	conf.System = system.Build(cfg.System)

	conf.Bus = bus.NewBus()

	return conf, nil
}

func (c *Config) GetLightOrGroup(lightOrGroup string) []*light.Light {
	if l := c.Lights.GetByName(lightOrGroup); l != nil {
		return []*light.Light{l}
	}
	if g, found := c.LightGroups[lightOrGroup]; found {
		return g.Lights()
	}
	return nil
}
