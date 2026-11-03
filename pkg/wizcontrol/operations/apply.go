package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
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
					operationsToDo[light] = tc.ColorName
					logger.WithField("device", light.Device.Name()).Debugf("Targetted color asserted")
				}
			}
		}
	}

	// Initiate application operations

	var asyncs wiz.AsyncOperations[*wiz.SetPilotResult]

	for light, colorName := range operationsToDo {
		color := colors.Colors[colorName] // Already validated as present
		logger := logger.WithField("device", light.Device.Name())

		light.SetLastProgrammedColor(colorName, color)

		if !unconditional {
			// We are not in unconditional mode, so we skip lights that are locked, unavailable or appear to be correctly set
			if light.Locker.IsLocked() {
				logger.Debugf("Skipping locked light")
				continue
			}
			if !light.Available() {
				logger.Debugf("Skipping unavailable light")
				continue
			}
			if light.LastColor() == color && light.On() {
				logger.Debugf("Light has correct color of %q and is turned on, skipping", color)
				continue
			} else {
				logger.Infof("Will apply color %q since current color is %q (intended %q) and light on state is %v", color, light.LastColor(), color, light.On())
			}
		}

		settings, err := wiz.SettingsFromString(color)
		if err != nil {
			logger.Warnf("Failed to parse color settings: %v", err)
			continue
		}
		settings.On()

		asyncs = append(asyncs, light.Device.SetPilotAsync(settings))
	}

	results := asyncs.AwaitAll()
	for _, result := range results {
		logger := logger.WithField("device", result.Device.Name())
		if result.Err != nil {
			logger.Warnf("Set pilot failed")
		} else {
			logger.Debugf("Set pilot succeeded")

			// Find the light
			light, found := lights.Lights[result.Device.Name()]
			if !found {
				logger.Errorf("Light not found")
				continue
			}

			// Light is definitively on now
			light.SetOn(true)

			// Retrieve the intended color for this light from the operationsToDo map
			colorName, found := operationsToDo[light]
			if !found {
				logger.Errorf("Color for light not found")
				continue
			}

			color, found := colors.Colors[colorName]
			if !found {
				logger.Errorf("Color not found: %s", colorName)
				continue
			}

			// Update light now that we definitively know it's setting
			light.SetLastColor(color)
			logger.Debugf("Successful application of color %q", color)
		}
	}
}
