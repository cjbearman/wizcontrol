package operations

import (
	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/senders"
)

func UpdateStatusForAllAvailableLights(force bool) {

	targetLights := wiz.NewDeviceGroup()
	if force {
		targetLights = lights.GetAllLightsDeviceGroup()
	} else {
		for _, light := range lights.GetLights() {
			if light.Available() {
				dev := light.GetDevice()
				if dev == nil {
					continue
				}
				targetLights.AddDevice(dev)
			}
		}
	}

	for _, result := range targetLights.GetPilot().AwaitAll() {
		handleStatusUpdate(result)
	}
}

func handleStatusUpdate(statusUpdate wiz.DeviceResultAndError[*wiz.PilotResult]) {
	logger := common.Logger.GetLogger("status-update").WithField("device", statusUpdate.Device.Name())
	// Find the light
	light, exists := lights.GetLights()[statusUpdate.Device.Name()]
	if !exists {
		logger.Errorf("Light not found")
		return
	}

	if statusUpdate.Err != nil {
		logger.Errorf("Error updating status: %v", statusUpdate.Err)
		return
	}

	on := statusUpdate.Result.State
	light.SetOn(on)

	color, err := wiz.ColorStringFromPilot(statusUpdate.Result)
	if err != nil {
		logger.Errorf("Error parsing active light color: %v", err)
		return
	}
	light.SetLastColor(color)
	logger.Debugf("State: color=%s, on=%v", color, on)

	// MQTT update
	_ = senders.SendLightStatus(light)
}
