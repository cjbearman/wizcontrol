package newconfig

import (
	"os"

	"github.com/cjbearman/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
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

type NewConfigYAML struct {
	Lights    []light.ConfigYAML    `yaml:"lights"`
	Colors    []color.ConfigYAML    `yaml:"colors"`
	Scenes    []scene.ConfigYAML    `yaml:"scenes"`
	Calendars []calendar.ConfigYAML `yaml:"calendar"`
	MQTT      *mqtt.ConfigYAML      `yaml:"mqtt"`
	HTTP      *http.ConfigYAML      `yaml:"http"`
	Discovery discovery.ConfigYAML  `yaml:"discovery"`
	System    system.ConfigYAML     `yaml:"system"`
}

func LoadNew(path string) (*NewConfigYAML, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close() // nolint:errcheck

	decoder := yaml.NewDecoder(fp)
	decoder.KnownFields(true) // Reject unknown fields

	cfg := &NewConfigYAML{}
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Build(cfg *NewConfigYAML) error {
	common.Logger = logger.NewLoggerFactory(logger.Debug, logger.NewColorStringFormatter(), logger.NewSyncWriter(os.Stdout)).GetLogger("root")

	logger := common.Logger.GetLogger("build")
	lights, lightgroups, err := light.Build(&cfg.Lights)
	if err != nil {
		logger.Errorf("Error parsing lights: %v", err)
		return err
	}

	for _, light := range lights {
		logger.Infof("Built light: %v", light)
	}
	for _, lightgroup := range lightgroups {
		logger.Infof("Built light group: %v", lightgroup)
	}

	colors, err := color.Build(&cfg.Colors)
	if err != nil {
		logger.Errorf("Error parsing colors: %v", err)
		return err
	}
	for _, color := range colors {
		logger.Infof("Built color: %v", color)
	}

	scenes, err := scene.Build(&cfg.Scenes, colors, lights, lightgroups)
	if err != nil {
		logger.Errorf("Error parsing scenes: %v", err)
		return err
	}
	for _, scene := range scenes {
		logger.Infof("Built scene: %v", scene)
	}

	calendars, err := calendar.Build(cfg.Calendars, scenes)
	if err != nil {
		logger.Errorf("Error parsing calendars: %v", err)
		return err
	}
	for _, calendar := range calendars {
		logger.Infof("Built calendar: %v", calendar)
	}

	mqtt := mqtt.Build(cfg.MQTT)
	logger.Infof("Built MQTT: %v", mqtt)

	http := http.Build(cfg.HTTP)
	logger.Infof("Built HTTP: %v", http)

	discovery := discovery.Build(cfg.Discovery)
	logger.Infof("Build Disc: %v", discovery)

	system := system.Build(cfg.System)
	logger.Infof("Build system: %v", system)

	return nil
}
