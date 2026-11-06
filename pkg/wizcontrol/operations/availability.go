package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/senders"
)

func UpdateAvailability(cnf *config.Config) {
	logger := common.Logger.GetLogger("availability")

	updateStatusOps := wiz.AsyncOperations[*wiz.PilotResult]{}

	for _, result := range cnf.Lights.All().GetAllLightsDeviceGroup().Reachable().AwaitAll() {
		logger := logger.WithField("device", result.Device.Name())
		if light, exists := cnf.Lights.All()[result.Device.Name()]; exists {
			lastState := light.Available()
			if result.Result && !lastState {
				logger.Infof("Device became reachable, updating status")
				updateStatusOps = append(updateStatusOps, result.Device.GetPilotAsync())
				light.SetAvailable(true)

			} else if !result.Result && lastState {
				logger.Warnf("Light %q has become unreachable", result.Device.Name())
				light.SetAvailable(false)
				// Light not available, so update MQTT status since we won't be updating it's status
				_ = senders.SendLightStatus(light)
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
			handleStatusUpdate(cnf, statusUpdate)
			// Above function will update MQTT status for these lights
		}
	}
}
