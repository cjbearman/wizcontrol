package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	l "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"

	c "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
)

// lightSettings holds the scene and color information for a light.
type lightSettings struct {
	scene string
	color *c.Color
}

// ApplyScenes will apply the list of scenes to the lights affected by those scenes
// The list of scenes is ordered, and scenes are applied to lights in that order, with later scenes taking precedence.
func ApplyScenes(cnf *config.Config, scenes []*scene.Scene, sendMqtt bool) {

	// Add the list of scene names to the logger
	sceneNames := []string{}
	for _, s := range scenes {
		sceneNames = append(sceneNames, s.Name())
	}
	logger := common.Logger.GetLogger("apply").WithField("scenes", sceneNames)

	// Prefill a map of all lights with empty scene settings, we will update this as
	// we figure out what is going to be set
	lightSceneSettings := make(map[*l.Light]lightSettings)
	for _, light := range cnf.Lights {
		lightSceneSettings[light] = lightSettings{}
	}

	// Map lights to their target colors
	// mappings from later scenes WILL take precedence by overwriting previous determinations
	operationsToDo := make(map[*l.Light]*c.Color)

	for _, scene := range scenes {
		elements := scene.GetElements()
		logger := logger.WithField("scene", scene.Name())

		// Each scene consists of multiple elements, each targeting specific targets (lights or light groups) with a color
		// later elements will take precedence if they target the same lights as earlier elements or scenes
		for _, targets := range elements {
			colorName := targets.GetColor()
			color := cnf.Colors[colorName]
			if color == nil {
				// Should not really happen, but it's not fatal
				logger.Warnf("Color not found: %s", colorName)
				continue
			}

			logger := logger.WithField("colorName", colorName).WithField("color", color)

			for _, light := range targets.GetTargets() {
				operationsToDo[light] = color
				logger.WithField("device", light.Name()).Debugf("Targetted color asserted")
				lightSceneSettings[light] = lightSettings{
					scene: scene.Name(),
					color: color,
				}
			}
		}
	}

	// Update programming information for all lights
	// those that were not included in a scene will have empty values
	for light, settings := range lightSceneSettings {
		light.SetActiveScene(settings.scene)
		light.SetLastProgrammedColor(settings.color)
	}

	// Lights that are locked get added / overriden here
	for _, light := range cnf.Lights {
		if locked, color, _ := light.Locker().IsLocked(); locked {
			setToColor := cnf.Colors.GetByValue(color)
			operationsToDo[light] = setToColor
			logger.Debugf("Locked light %q will be set to color %q", light.Name(), setToColor)
			light.SetActiveScene("")
		}
	}

	// Initiate application operations

	var asyncs wiz.AsyncOperations[*wiz.SetPilotResult]

	for light, color := range operationsToDo {
		logger := logger.WithField("device", light.Name())

		if !light.Available() {
			logger.Debugf("Skipping unavailable light")
			continue
		}

		if !light.On() && color.Value() == c.Off {
			logger.Debugf("Light is off and intended color is off, skipping")
			continue
		}

		if color.SameColorAs(light.LastActualColor()) {
			logger.Debugf("Light has correct color of %q, skipping", color)
			continue
		} else {
			logger.Infof("Will apply color %q to light %q since it's different from current color %q", color, light.Name(), light.LastActualColor())
		}

		settings, err := color.GetWizSettings()
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
			light, found := cnf.Lights[result.Device.Name()]
			if !found {
				logger.Errorf("Light not found")
				continue
			}

			light.SetOn(operationsToDo[light].Value() != c.Off)

			// Retrieve the intended color for this light from the operationsToDo map since we now
			// know we succesfully set the light to that color
			color, found := operationsToDo[light]
			if !found {
				logger.Errorf("Color for light not found")
				continue
			}

			// Update light now that we definitively know it's setting
			light.SetLastActualColor(color)

			logger.Debugf("Successful application of color %q (%q)", color.Name(), color.Value())
		}
	}

	// Dispatch MQTT status for all lights
	if sendMqtt {
		for _, light := range cnf.Lights {
			sendLightStatus(cnf.Bus, light)
		}
	}
}
