package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/scenes"
)

func ApplyScenes(sceneNames []string, unconditional bool) {
	logger := common.Logger.GetLogger("apply").WithField("scenes", sceneNames).WithField("unconditional", unconditional)

	// Map lights to their target colors
	// mappings from later scenes WILL take precedence by overwriting previous determinations
	operationsToDo := make(map[*lights.Light]string)

	for _, sceneName := range sceneNames {

		targets, found := scenes.Scenes[sceneName]
		if !found {
			logger.Warnf("Scene not found: %s", sceneName)
			continue
		}
		logger := logger.WithField("scene", sceneName)

		for _, tc := range targets {

			color, found := colors.Colors[tc.ColorName]
			if !found {
				logger.Warnf("Color not found: %s", tc.ColorName)
				continue
			}
			logger := logger.WithField("colorName", tc.ColorName).WithField("color", color)

			for _, target := range tc.Targets {
				lights, err := lights.GetLightOrGroup(target)
				if err != nil {
					logger.Warnf("Light or group not found: %s", target)
					continue
				}
				for _, light := range lights {
					operationsToDo[light] = color
					logger.WithField("device", light.Light.Name()).Debugf("Targetted color asserted")
				}
			}
		}
	}

	// Initiate application operations

	var asyncs wiz.AsyncOperations[*wiz.SetPilotResult]

	for light, color := range operationsToDo {
		logger := logger.WithField("device", light.Light.Name())
		if !unconditional {
			// We are not in unconditional mode, so we skip lights that are unavailable or appear to be correctly set
			if !light.Available {
				logger.Debugf("Skipping unavailable light")
				continue
			}
			if light.LastColor == color && light.On {
				logger.Debugf("Skipping light with correct color and state")
				continue
			}
		}

		settings, err := wiz.SettingsFromString(color)
		if err != nil {
			logger.Warnf("Failed to parse color settings: %v", err)
			continue
		}
		settings.On()

		asyncs = append(asyncs, light.Light.SetPilotAsync(settings))
	}

	results := asyncs.AwaitAll()
	for _, result := range results {
		logger := logger.WithField("device", result.Device.Name())
		if result.Err != nil {
			logger.Warnf("Set pilot failed")
		} else {
			logger.Debugf("Set pilot succeeded")
		}
	}
}
