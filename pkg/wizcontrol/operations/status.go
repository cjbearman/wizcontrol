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

// UpdateStatusForAllAvailableLights retrieves the current status of all available lights
// and updates their internal state. If 'force' is true, it will update all lights regardless of availability.
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

// handleStatusUpdate processes the status update for a single light, updating its internal state and sending MQTT updates.
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

	r, g, b, cw, ww := statusUpdate.Result.R, statusUpdate.Result.G, statusUpdate.Result.B, statusUpdate.Result.C, statusUpdate.Result.W
	br, temp, scene := statusUpdate.Result.Dimming, statusUpdate.Result.Temp, statusUpdate.Result.SceneId
	state := statusUpdate.Result.State
	light.SetLastRawColor(r, g, b, cw, ww, br, temp, scene, state)

	light.RecordRSSI(statusUpdate.Result.Rssi)

	col := cnf.Colors.GetByValue(color)
	light.SetLastActualColor(col)
	logger.Debugf("State: color=%s, on=%v", color, on)

	// MQTT update
	sendLightStatus(cnf.Bus, light)
}

// sendLightStatus publishes the current status of the light to the MQTT bus.
func sendLightStatus(msgBus *bus.Bus, light *l.Light) {
	logger := common.Logger.GetLogger("light-status").WithField("device", light.Name())
	status := responses.BuildLightStatus(light)

	encoded, err := json.Marshal(status)
	if err != nil {
		logger.Warnf("Error encoding light status: %v", err)
		return
	}

	msg := bus.Message{
		Topic:   "light/" + light.MqqtID(),
		Payload: encoded,
	}

	msgBus.Publish(msg)
}
