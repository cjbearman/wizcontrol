package newconfig

import (
	"os"

	"github.com/cjbearman/logger"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"gopkg.in/yaml.v3"
)

type NewConfigYAML struct {
	Lights []light.ConfigYAML `yaml:"lights"`
	Colors []color.ConfigYAML `yaml:"colors"`
}

func LoadNew(path string) (*NewConfigYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &NewConfigYAML{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
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

	return nil
}
