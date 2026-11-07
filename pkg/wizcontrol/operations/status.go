package operations

import (
	"encoding/json"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/bus"
	l "github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/validator"
)

func UpdateStatusForAllAvailableLights(cnf *config.Config, force bool) {

	targetLights := wiz.NewDeviceGroup()
	if force {
		targetLights = cnf.Lights.GetAllLightsDeviceGroup()
	} else {
		for _, light := range cnf.Lights {
			if light.Available() {
				dev := light.Device()
				if dev == nil {
					continue
				}
				targetLights.AddDevice(dev)
			}
		}
	}

	for _, result := range targetLights.GetPilot().AwaitAll() {
		handleStatusUpdate(cnf, result)
	}
}

func handleStatusUpdate(cnf *config.Config, statusUpdate wiz.DeviceResultAndError[*wiz.PilotResult]) {

	logger := common.Logger.GetLogger("status-update").WithField("device", statusUpdate.Device.Name())
	// Find the light
	light, exists := cnf.Lights[statusUpdate.Device.Name()]
	if !exists {
		logger.Errorf("Light not found")
		return
	}

	if statusUpdate.Err != nil {
		logger.Errorf("Error updating status: %v", statusUpdate.Err)
		return
	}

	macNormalized, err := validator.MAC(statusUpdate.Result.Mac, true)
	if err != nil {
		logger.Warnf("Error normalizing MAC address: %v", err)
		return
	}

	light.SetMAC(macNormalized)

	on := statusUpdate.Result.State
	light.SetOn(on)

	color, err := wiz.ColorStringFromPilot(statusUpdate.Result)
	if err != nil {
		logger.Errorf("Error parsing active light color: %v", err)
		return
	}
	col := cnf.Colors.GetByValue(color)
	light.SetLastActualColor(col)
	logger.Debugf("State: color=%s, on=%v", color, on)

	// MQTT update
	sendLightStatus(cnf.Bus, light)
}

func sendLightStatus(msgBus *bus.Bus, light *l.Light) {
	logger := common.Logger.GetLogger("light-status").WithField("device", light.Name())
	status := responses.BuildLightStatus(light)

	encoded, err := json.Marshal(status)
	if err != nil {
		logger.Warnf("Error encoding light status: %v", err)
		return
	}

	msg := bus.Message{
		Topic:   "light/" + light.Name(),
		Payload: encoded,
	}

	msgBus.Publish(msg)
}
