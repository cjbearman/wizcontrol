package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	l "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/senders"

	c "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
)

type lightSettings struct {
	scene string
	color *c.Color
}

func ApplyScenes(cnf *config.Config, sceneNames []string, unconditional bool, sendMqtt bool) {
	logger := common.Logger.GetLogger("apply").WithField("scenes", sceneNames).WithField("unconditional", unconditional)

	// Prefill a map of all lights with empty scene settings, we will update this as
	// we figure out what is going to be set
	lightSceneSettings := make(map[*l.Light]lightSettings)
	for _, light := range cnf.Lights {
		lightSceneSettings[light] = lightSettings{}
	}

	// Map lights to their target colors
	// mappings from later scenes WILL take precedence by overwriting previous determinations
	operationsToDo := make(map[*l.Light]*c.Color)

	for _, sceneName := range sceneNames {

		elements, found := cnf.Scenes.GetSceneElementsBySceneName(sceneName)
		if !found {
			logger.Warnf("Scene not found: %s", sceneName)
			continue
		}
		logger := logger.WithField("scene", sceneName)

		for _, targets := range elements {
			colorName := targets.GetColor()
			color := cnf.Colors.GetByName(colorName)
			if color == nil {
				logger.Warnf("Color not found: %s", colorName)
				continue
			}

			logger := logger.WithField("colorName", colorName).WithField("color", color)

			for _, target := range targets.GetTargets() {
				lights := l.GetLightOrGroup(target, cnf.Lights, cnf.LightGroups)
				if len(lights) == 0 {
					logger.Warnf("Light or group not found: %s", target)
					continue
				}
				for _, light := range lights {
					operationsToDo[light] = color
					logger.WithField("device", light.Name()).Debugf("Targetted color asserted")
					lightSceneSettings[light] = lightSettings{
						scene: sceneName,
						color: color,
					}
				}
			}
		}
	}

	// Update programming information for all lights
	// those that were not included in a scene will have empty values
	for light, settings := range lightSceneSettings {
		light.SetActiveScene(settings.scene)
		light.SetLastProgrammedColor(settings.color.Name(), settings.color.Value())
	}

	// Lights that are locked get added / overriden here
	for _, light := range cnf.Lights.All() {
		if locked, color, _ := light.Locker().IsLocked(); locked {
			operationsToDo[light] = cnf.Colors.GetByColor(color)
			logger.Debugf("Locked light %q will be set to color %q", light.Name(), operationsToDo[light].Value())
			light.SetActiveScene("")
		}
	}

	// Initiate application operations

	var asyncs wiz.AsyncOperations[*wiz.SetPilotResult]

	for light, color := range operationsToDo {
		logger := logger.WithField("device", light.Name())

		if !unconditional {
			// We are not in unconditional mode, so we skip lights that are locked, unavailable or appear to be correctly set
			if !light.Available() {
				logger.Debugf("Skipping unavailable light")
				continue
			}
			if light.LastActualColor().Value() == color.Value() {
				logger.Debugf("Light has correct color of %q, skipping", color)
				continue
			} else {
				logger.Infof("Will apply color %q since current color is %q (intended %q) and light on state is %v", color, light.LastActualColor().Value(), color, light.On())
			}
		}

		settings, err := wiz.SettingsFromString(color.Value())
		if err != nil {
			logger.Warnf("Failed to parse color settings from %q: %v", color, err)
			continue
		}
		if color.Value() != c.Off {
			settings.On()
		}

		dev := light.Device()
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
			light, found := cnf.Lights.All()[result.Device.Name()]
			if !found {
				logger.Errorf("Light not found")
				continue
			}

			light.SetOn(operationsToDo[light].Value() != c.Off)

			// Retrieve the intended color for this light from the operationsToDo map
			colorSpec, found := operationsToDo[light]
			if !found {
				logger.Errorf("Color for light not found")
				continue
			}

			// Update light now that we definitively know it's setting
			light.SetLastActualColor(colorSpec.Name(), colorSpec.Value())

			logger.Debugf("Successful application of color %q (%q)", colorSpec.Name(), colorSpec.Value())
		}
	}

	// Dispatch MQTT status for all lights
	if sendMqtt {
		for _, light := range cnf.Lights.All() {
			_ = senders.SendLightStatus(light)
		}
	}
}
