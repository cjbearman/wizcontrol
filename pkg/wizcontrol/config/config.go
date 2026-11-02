package config

import (
	"fmt"
	"os"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/calendar"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/daemon"

	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/scenes"
	"gopkg.in/yaml.v3"
)

type Light = string
type Color = string

const (
	DefaultEvaluationPeriod = 30 * time.Second
	DefaultRecheckPeriod    = 5 * time.Minute
)

// Config describes the structure in config.yaml.
type Config struct {
	Timezone         string                       `yaml:"timezone"`
	EvaluationPeriod time.Duration                `yaml:"evaluation_period"`
	RecheckPeriod    time.Duration                `yaml:"recheck_period"`
	Colors           map[string]Color             `yaml:"colors"`
	Lights           map[string]Light             `yaml:"lights"`
	LightGroups      map[string][]Light           `yaml:"light_groups"`
	Scenes           map[string][]scenes.CnfScene `yaml:"scenes"`
	DefaultScene     string                       `yaml:"default_scene"`
	Calendar         []CalendarEntry              `yaml:"calendar"`
}

// CalendarEntry describes when a scene should apply.
type CalendarEntry struct {
	Name   string `yaml:"name"`
	From   string `yaml:"from"`
	To     string `yaml:"to"`
	Before string `yaml:"before"`
	After  string `yaml:"after"`
	Scene  string `yaml:"scene"`
}

// Load reads a YAML config from disk.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Parse(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Parse() error {
	logger := common.Logger.GetLogger("config-parser")

	// Settings
	if c.EvaluationPeriod == 0 {
		daemon.EvaluationPeriod = DefaultEvaluationPeriod
	} else {
		daemon.EvaluationPeriod = c.EvaluationPeriod
	}
	if c.RecheckPeriod == 0 {
		daemon.RecheckPeriod = DefaultRecheckPeriod
	} else {
		daemon.RecheckPeriod = c.RecheckPeriod
	}
	if c.Timezone != "" {
		loc, err := time.LoadLocation(c.Timezone)
		if err != nil {
			return fmt.Errorf("failed to load timezone %q: %v", c.Timezone, err)
		}
		common.Location = loc
	}

	// Colors
	for name, color := range c.Colors {
		// Each color must be unique
		// Whilst this is a map, and cannot have duplicates, there may be preloaded colors with the same name
		if _, found := colors.Colors[name]; found {
			logger.Errorf("Color %q is pre-defined and the name is reserved", name)
		}

		// Each color must be a valid color string, we will also rationalize it to include default
		// brightness and speed values as appropriate
		rationalized, err := wiz.RationalizeColor(color)
		if err != nil {
			return fmt.Errorf("Color %q is invalid: %v", name, err)
		}

		// Each named color must have a unique value
		for existingName, existingValue := range colors.Colors {
			if rationalized == existingValue {
				return fmt.Errorf("Color named %q has the same functional value as %q", name, existingName)
			}
		}

		colors.Colors[name] = rationalized
		logger.Debugf("Loaded color %q: %s", name, rationalized)
	}

	// Lights
	for name, ip := range c.Lights {
		if err := lights.AddLight(name, ip); err != nil {
			return fmt.Errorf("failed to add light %q: %v", name, err)
		}
		logger.Debugf("Added light %q with IP %q", name, ip)
	}

	// Light groups
	for name, members := range c.LightGroups {
		if err := lights.AddLightGroup(name, members); err != nil {
			return fmt.Errorf("failed to add light group %q: %v", name, err)
		}
		logger.Debugf("Added light group %q with members %v", name, members)
	}

	// Scenes
	for name, scene := range c.Scenes {
		if err := scenes.AddScene(name, scene); err != nil {
			return fmt.Errorf("failed to add scene %q: %v", name, err)
		}
		logger.Debugf("Added scene %q", name)
	}

	// Calendar entries
	for _, entry := range c.Calendar {
		if err := calendar.AddCalendarEntry(entry.Name, entry.From, entry.To, entry.Before, entry.After, entry.Scene); err != nil {
			return fmt.Errorf("failed to add calendar entry %q: %v", entry.Name, err)
		}
		logger.Debugf("Added calendar entry %q", entry.Name)
	}

	// Enforce default scene
	if c.DefaultScene == "" {
		return fmt.Errorf("default scene is not defined")
	}
	if _, exists := c.Scenes[c.DefaultScene]; !exists {
		return fmt.Errorf("default scene %q does not exist", c.DefaultScene)
	}

	return nil
}
