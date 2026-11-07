package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// UpdateAvailability pings all the lights (except undiscovered ones)
// and updates their availability status
// It dispatches MQTT light updates when status changes
// Also retrieves the current status of any lights that just became available and updates those
func UpdateAvailability(cnf *config.Config) {
	logger := common.Logger.GetLogger("availability")

	updateStatusOps := wiz.AsyncOperations[*wiz.PilotResult]{}

	// Ping all lights that have devices (configured by IP or discovered) and wait for the results
	// then iterate the results
	for _, result := range cnf.Lights.GetAllLightsDeviceGroup().Reachable().AwaitAll() {

		logger := logger.WithField("device", result.Device.Name())

		// Light can be found based on it's device name which matches the light name
		if light, exists := cnf.Lights[result.Device.Name()]; exists {

			// Retrieve the last known availability state of the light, so we know if the state changed
			// on initial run, all lights will have a last state of unavailable
			lastState := light.Available()
			if result.Result && !lastState {
				logger.Infof("Device became reachable, updating status")
				updateStatusOps = append(updateStatusOps, result.Device.GetPilotAsync())
				light.SetAvailable(true)

				// N.B. we do not send MQTT status here because we will do it after the status update completes

			} else if !result.Result && lastState {
				logger.Warnf("Light %q has become unreachable", result.Device.Name())
				light.SetAvailable(false)

				// Light not available, so update MQTT status since we won't be updating it's status
				sendLightStatus(cnf.Bus, light)

			} else {
				if lastState {
					logger.Debugf("Light %q had no change in availability state, currently available", result.Device.Name())
				} else {
					logger.Debugf("Light %q had no change in availability state, currently unavailable", result.Device.Name())
				}
			}
		} else {
			logger.Errorf("Light not found")
		}
	}

	// For any light that just became available, we'll do a getPilot so we can retrieve it's current status
	// which may have changed since we last saw it
	if len(updateStatusOps) > 0 {
		for _, statusUpdate := range updateStatusOps.AwaitAll() {
			light, found := cnf.Lights[statusUpdate.Device.Name()]
			if !found {
				logger.Errorf("Light %q not found for status update", statusUpdate.Device.Name())
				continue
			}

			// Update the light's availability based on the status update result
			sendLightStatus(cnf.Bus, light)
		}
	}
}
