package config

import (
	"fmt"
	"os"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/calendar"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/mqtt"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/server"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
	"gopkg.in/yaml.v3"
)

type Light = string
type Color = string

const (
	DefaultEvaluationPeriod = 30 * time.Second
	DefaultRecheckPeriod    = 5 * time.Minute
	DefaultServerPort       = 12120
)

type MqttConfig struct {
	Url            string `yaml:"url"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Topic          string `yaml:"topic"`
	ControlEnabled *bool  `yaml:"allow_control"`
}

type ServerAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Server struct {
	Port      int         `yaml:"port"`
	CertChain string      `yaml:"cert_chain"`
	Key       string      `yaml:"key"`
	Auth      *ServerAuth `yaml:"auth"`
}

// Config describes the structure in config.yaml.
type Config struct {
	Timezone         string                       `yaml:"timezone"`
	EvaluationPeriod time.Duration                `yaml:"evaluation_period"`
	RecheckPeriod    time.Duration                `yaml:"recheck_period"`
	Server           Server                       `yaml:"http"`
	Mqtt             *MqttConfig                  `yaml:"mqtt"`
	Colors           map[string]Color             `yaml:"colors"`
	Lights           map[string]Light             `yaml:"lights"`
	LightGroups      map[string][]Light           `yaml:"light_groups"`
	Scenes           map[string][]scenes.CnfScene `yaml:"scenes"`
	Calendar         []CalendarEntry              `yaml:"calendar"`
}

// CalendarEntry describes when a scene should apply.
type CalendarEntry struct {
	Name      string             `yaml:"name"`
	From      string             `yaml:"from"`
	To        string             `yaml:"to"`
	Before    string             `yaml:"before"`
	After     string             `yaml:"after"`
	Scenes    []string           `yaml:"scenes"`
	SceneMode calendar.SceneMode `yaml:"mode"`
	Frequency *time.Duration     `yaml:"frequency"`
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

	if c.Server.Port == 0 {
		c.Server.Port = DefaultServerPort
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port %d is out of valid range (1-65535)", c.Server.Port)
	}

	server.ServerPort = c.Server.Port

	if (c.Server.CertChain == "" && c.Server.Key != "") || (c.Server.CertChain != "" && c.Server.Key == "") {
		return fmt.Errorf("Both http.cert_chain and http.key must be provided together to enable TLS")
	}

	server.CertChain = c.Server.CertChain
	server.Key = c.Server.Key

	if c.Server.Auth != nil {
		if c.Server.Auth.Username == "" {
			return fmt.Errorf("Must provide http.auth.username when configuring server auth")
		}
		if c.Server.Auth.Password == "" {
			return fmt.Errorf("Must provide http.auth.password when configuring server auth")
		}
		server.AuthUsername = c.Server.Auth.Username
		server.AuthPassword = c.Server.Auth.Password
		if server.CertChain == "" {
			logger.Warnf("*** HTTP auth is configured but TLS is not enabled. Credentials will be sent in plaintext over the network.")
		}
	}

	if c.Mqtt != nil {
		if c.Mqtt.Url == "" {
			return fmt.Errorf("Must provide MQTT URL when configuring MQTT")
		}
		mqtt.MqttURL = c.Mqtt.Url
		mqtt.MqttUsername = c.Mqtt.Username
		if c.Mqtt.Username == "" {
			return fmt.Errorf("Must provide MQTT username when configuring MQTT")
		}
		mqtt.MqttPassword = c.Mqtt.Password
		if c.Mqtt.Password == "" {
			return fmt.Errorf("Must provide MQTT password when configuring MQTT")
		}
		mqtt.MqttTopic = c.Mqtt.Topic
		if c.Mqtt.Topic == "" {
			return fmt.Errorf("Must provide MQTT topic when configuring MQTT")
		}
		if c.Mqtt.ControlEnabled != nil {
			mqtt.ControlEnabled = *c.Mqtt.ControlEnabled
		}
		logger.Infof("Configured MQTT on %s", c.Mqtt.Url)
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
		if err := calendar.AddCalendarEntry(entry.Name, entry.From, entry.To, entry.Before, entry.After, entry.Scenes, entry.SceneMode, entry.Frequency); err != nil {
			return fmt.Errorf("failed to add calendar entry %q: %v", entry.Name, err)
		}
		logger.Debugf("Added calendar entry %q", entry.Name)
	}

	return nil
}
