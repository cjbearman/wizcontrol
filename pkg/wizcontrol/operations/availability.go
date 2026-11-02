package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/lights"
)

func UpdateAvailability(withStatusUpdatesWhereNeeded bool) {
	logger := common.Logger.GetLogger("availability")

	updateStatusOps := wiz.AsyncOperations[*wiz.PilotResult]{}

	for _, result := range lights.AllDevicesGroup.Reachable().AwaitAll() {
		logger := logger.WithField("device", result.Device.Name())
		if light, exists := lights.Lights[result.Device.Name()]; exists {
			lastState := light.Available()
			if result.Result && lastState == false {
				if !withStatusUpdatesWhereNeeded {
					logger.Infof("Device became reachable, no status update requested")
				} else {
					logger.Warnf("Device became unreachable")
					updateStatusOps = append(updateStatusOps, result.Device.GetPilotAsync())
				}
				light.SetAvailable(true)
			} else if !result.Result && lastState == true {
				logger.Warnf("Light %q has become unreachable", result.Device.Name())
				light.SetAvailable(false)
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
			handleStatusUpdate(statusUpdate)
		}
	}
}
