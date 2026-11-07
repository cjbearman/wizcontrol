package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

func UpdateAvailability(cnf *config.Config) {
	logger := common.Logger.GetLogger("availability")

	updateStatusOps := wiz.AsyncOperations[*wiz.PilotResult]{}

	for _, result := range cnf.Lights.GetAllLightsDeviceGroup().Reachable().AwaitAll() {
		logger := logger.WithField("device", result.Device.Name())
		if light, exists := cnf.Lights[result.Device.Name()]; exists {
			lastState := light.Available()
			if result.Result && !lastState {
				logger.Infof("Device became reachable, updating status")
				updateStatusOps = append(updateStatusOps, result.Device.GetPilotAsync())
				light.SetAvailable(true)

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

	if len(updateStatusOps) > 0 {
		for _, statusUpdate := range updateStatusOps.AwaitAll() {
			light, found := cnf.Lights[statusUpdate.Device.Name()]
			if !found {
				logger.Errorf("Light %q not found for status update", statusUpdate.Device.Name())
				continue
			}

			sendLightStatus(cnf.Bus, light)
		}
	}
}
