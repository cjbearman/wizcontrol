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
func ApplyScenes(cnf *config.Config, scenes []*scene.Scene) {

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
			color := targets.GetColor()

			logger := logger.WithField("colorName", color.Name()).WithField("color", color.Value())

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

	// Update the reported status (active scene and programmed color)
	// for all lights
	for light, settings := range lightSceneSettings {
		light.SetActiveScene(settings.scene)
		light.SetLastProgrammedColor(settings.color)
	}

	// Lights that are locked get added / overriden here
	for _, light := range cnf.Lights {
		if locked, color, _ := light.Locker().IsLocked(); locked {
			operationsToDo[light] = color
			logger.Debugf("Locked light %q will be set to color %q", light.Name(), color.Value())
			light.SetActiveScene("")
		}
	}

	// We now prepare async set pilots for each light that needs to be updated
	var asyncs wiz.AsyncOperations[*wiz.SetPilotResult]

	for light, color := range operationsToDo {
		logger := logger.WithField("device", light.Name())

		// Lights must be available for us to work with them
		if !light.Available() {
			logger.Debugf("Skipping unavailable light")
			continue
		}

		// Skip lights that are already off if the intended color is also off
		if !light.On() && color.Value() == c.Off {
			logger.Debugf("Light is off and intended color is off, skipping")
			continue
		}

		// Skip lights that already have the intended color
		if color.SameColorAs(light.LastActualColor()) {
			logger.Debugf("Light has correct color of %q, skipping", color)
			continue
		} else {
			logger.Infof("Will apply color %q to light %q since it's different from current color %q", color, light.Name(), light.LastActualColor())
		}

		// Generate the Wiz-specific settings for the intended color
		settings, err := color.GetWizSettings()
		if err != nil {
			logger.Warnf("Failed to parse color settings from %q: %v", color, err)
			continue
		}
		if color.Value() != c.Off {
			settings.On()
		}

		// Get the wiz device for the light
		dev := light.Device()
		if dev == nil {
			// Ohoh, this device needs discovery and hasn't been discovered yet, we have to skip it
			logger.Warnf("No discovered device, skipping")
			continue
		}

		// At this point, we have a valid discovered device for the light
		// Start the set pilot operation and append it to our list of operations
		// The wiz settings we are applying are attached to the async as correlation data so that
		// we can recover them when the async finishes
		asyncs = append(asyncs, dev.SetPilotAsync(settings).SetData(settings))

	}

	// Wait for all operations to complete
	results := asyncs.AwaitAll()

	// And go through the results
	for _, result := range results {

		logger := logger.WithField("device", result.Device.Name())
		if result.Err != nil {
			logger.Warnf("Set pilot failed")
		} else {
			logger.Debugf("Set pilot succeeded")

			// The result doesn't have our light in it, but it has the device, which has the same name as our light
			// as a back-reference. Use that to find the light instance.
			light, found := cnf.Lights[result.Device.Name()]
			if !found {
				// Should not happen
				logger.Errorf("Light not found")
				continue
			}

			// Record the wiz settings data from the operation correlation data
			pending := result.Data.(*wiz.Settings)

			// Update the light's "on" state based on the intended color
			light.SetOn(operationsToDo[light].Value() != c.Off)

			// Retrieve the intended color for this light from the operationsToDo map since we now
			// know we succesfully set the light to that color
			color, found := operationsToDo[light]
			if !found {
				logger.Errorf("Color for light not found")
				continue
			}

			// Update the light's last actual color based on the intended color
			light.SetLastActualColor(color)

			// Update raw light settings based on the pending settings
			r, g, b := pending.GetRGB()
			light.SetLastRawColor(int(r), int(g), int(b),
				int(pending.GetCW()),
				int(pending.GetWW()),
				pending.GetBrightness(),
				pending.GetTemperature(),
				int(pending.GetScene()))

			logger.Debugf("Successful application of color %q (%q)", color.Name(), color.Value())
		}
	}

	// Dispatch MQTT status for all lights
	for _, light := range cnf.Lights {
		sendLightStatus(cnf.Bus, light)
	}
}
