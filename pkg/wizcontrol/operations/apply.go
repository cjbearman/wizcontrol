package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/senders"
)

type colorAndName struct {
	name, color string
}

func ApplyScenes(sceneNames []string, unconditional bool, sendMqtt, mqttAll bool) {
	logger := common.Logger.GetLogger("apply").WithField("scenes", sceneNames).WithField("unconditional", unconditional)
	sendStatusFor := []*lights.Light{}

	if mqttAll {
		// If we are being asked to send MQTT status for all lights, just prefill a list of all lights
		// into the list of lights to send
		logger.Debug("This is the first daemon run, status for all lights will be sent")
		for _, light := range lights.Lights {
			sendStatusFor = append(sendStatusFor, light)
		}
	}

	// Prefill a map of light to target scenes with all lights and initially empty values
	lightSceneSettings := make(map[*lights.Light]string)
	for _, light := range lights.Lights {
		lightSceneSettings[light] = ""
	}

	// Map lights to their target colors
	// mappings from later scenes WILL take precedence by overwriting previous determinations
	operationsToDo := make(map[*lights.Light]colorAndName)

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
					operationsToDo[light] = colorAndName{name: tc.ColorName, color: color}
					logger.WithField("device", light.GetName()).Debugf("Targetted color asserted")
					lightSceneSettings[light] = sceneName
				}
			}
		}
	}

	// Update target scenes for all lights,
	// this will clear the scene name from any lights that are not in any of our scenes
	// and also set the target scene for lights that are part of our scenes
	for light, sceneName := range lightSceneSettings {
		light.SetActiveScene(sceneName)
	}

	// Lights that are locked get added / overriden here
	for _, light := range lights.Lights {
		if locked, color, _ := light.Locker.IsLocked(); locked {
			operationsToDo[light] = colorAndName{name: colors.NameForColor(color), color: color}
			logger.Debugf("Locked light %q will be set to color %q", light.GetName(), color)
			light.SetActiveScene("")
		}
	}

	// Initiate application operations

	var asyncs wiz.AsyncOperations[*wiz.SetPilotResult]

	for light, colorSpec := range operationsToDo {
		color := colorSpec.color
		logger := logger.WithField("device", light.GetName())

		light.SetLastProgrammedColor(colorSpec.name, color)

		if !unconditional {
			// We are not in unconditional mode, so we skip lights that are locked, unavailable or appear to be correctly set
			if !light.Available() {
				logger.Debugf("Skipping unavailable light")
				continue
			}
			if light.LastColor() == color {
				logger.Debugf("Light has correct color of %q, skipping", color)
				continue
			} else {
				logger.Infof("Will apply color %q since current color is %q (intended %q) and light on state is %v", color, light.LastColor(), color, light.On())
			}
		}

		settings, err := wiz.SettingsFromString(color)
		if err != nil {
			logger.Warnf("Failed to parse color settings from %q: %v", color, err)
			continue
		}
		if color != colors.ColorOff {
			settings.On()
		}

		dev := light.GetDevice()
		if dev == nil {
			logger.Warnf("No discovered device, skipping")
			continue
		}
		asyncs = append(asyncs, dev.SetPilotAsync(settings))
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

			// We will be sending a status for this light, if mqtt dispatch is enabled
			sendStatusFor = append(sendStatusFor, light)

			light.SetOn(operationsToDo[light].color != colors.ColorOff)

			// Retrieve the intended color for this light from the operationsToDo map
			colorSpec, found := operationsToDo[light]
			if !found {
				logger.Errorf("Color for light not found")
				continue
			}

			// Update light now that we definitively know it's setting
			light.SetLastColor(colorSpec.name)
			logger.Debugf("Successful application of color %q (%q)", colorSpec.color, colorSpec.name)
		}
	}

	if sendMqtt {
		for _, light := range sendStatusFor {
			senders.SendLightStatus(light)
		}
	}
}
