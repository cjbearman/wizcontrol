package operations

import (
	log "github.com/cjbearman/logger"
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	c "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/color"
	l "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
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
				logger.WithField("device", light.ID()).Debugf("Targetted color asserted")
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
		if settings.scene == "" {
			// Light is not part of a scene, so we can go ahead and dispatch a status update for it now
			logger.Debugf("light not in scene, we can update it's status prematurely: %s", light.ID())
			sendLightStatus(cnf.Bus, light)
		}
	}

	// Lights that are locked get added / overriden here
	for _, light := range cnf.Lights {
		if locked, color, _ := light.Locker().IsLocked(); locked {
			operationsToDo[light] = color
			logger.Debugf("Locked light %q will be set to color %q", light.ID(), color.Value())
			light.SetActiveScene("")
		}
	}

	for light, color := range operationsToDo {
		logger := logger.WithField("device", light.ID())

		// Lights must be available for us to work with them
		if !light.Available() {
			logger.Debugf("Skipping unavailable light")
			sendLightStatus(cnf.Bus, light)
			continue
		}

		// Skip lights that are already off if the intended color is also off
		if !light.On() && color.Value() == c.Off {
			logger.Debugf("Light is off and intended color is off, skipping")
			sendLightStatus(cnf.Bus, light)
			continue
		}

		// Skip lights that already have the intended color
		if color.SameColorAs(light.LastActualColor()) {
			logger.Debugf("Light has correct color of %q, skipping", color)
			sendLightStatus(cnf.Bus, light)
			continue
		} else {
			logger.Infof("Will apply color %q to light %q since it's different from current color %q", color, light.ID(), light.LastActualColor())
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

		go func(dev *wiz.Device, settings *wiz.Settings, logger log.Logger, light *l.Light) {
			_, err := dev.SetPilot(settings)
			if err != nil {
				logger.Warnf("Set pilot failed: %v", err)
				return
			} else {
				logger.Debugf("Set pilot succeeded")

			}

			// Okay so our set pilot succeeded, we need to do a get pilot here
			// for total accuracy in setting things downstream
			res, err := dev.GetPilot()
			if err != nil {
				logger.Warnf("Get pilot failed, light is likely out of sync: %v", err)
				return
			} else {
				logger.Debugf("Get pilot succeeded: %v", res)
			}

			light.SetOn(res.State)
			if !res.State {
				light.SetLastActualColor(cnf.Colors.GetByValue(c.Off))
				return
			}

			wizcol, err := wiz.ColorStringFromPilot(res)
			if err != nil {
				logger.Warnf("Failed to parse color from pilot: %v", err)
				return
			}

			light.SetLastActualColor(cnf.Colors.GetByValue(wizcol))
			light.SetLastRawColor(res.R, res.G, res.B, res.C, res.W, res.Dimming, res.Temp, res.SceneId, res.State)
			logger.Debugf("successfully updated light state from pilot")

			// We can go ahead and send the light status
			sendLightStatus(cnf.Bus, light)
		}(dev, settings, logger, light)

	}
}
