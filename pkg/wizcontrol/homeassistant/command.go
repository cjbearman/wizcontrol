package homeassistant

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

func HandleLightCommand(cnf *config.Config, topic string, payload []byte) {
	logger := common.Logger.GetLogger("homeassistant-cmd")
	if cnf.HomeAssistant == nil {
		// Should not happen
		logger.Errorf("Home Assistant configuration is not available")
		return
	}

	logger.Infof("Received light command on topic %q: %s", topic, string(payload))

	// Parse the topic to figure out what we need to do
	parts := strings.Split(topic, "/")

	if len(parts) != 4 {
		logger.Errorf("Invalid light command, expected 4 parts in topic but got %d: %s", len(parts), topic)
		return
	}

	if parts[0] != cnf.MQTT.GetTopic() {
		logger.Errorf("Invalid light command, bad topic prefix: %s", parts[0])
		return
	}
	if parts[1] != "light" {
		logger.Errorf("Invalid light command, expected 'light' as the second topic part but got: %s", parts[1])
		return
	}
	lightID := parts[2]
	var light *light.Light
	for _, l := range cnf.Lights {
		if l.MqqtID() == lightID {
			light = l
			break
		}
	}
	if light == nil {
		logger.Errorf("Light with ID %q not found", lightID)
		return
	}

	if !light.Available() {
		logger.Errorf("Light with ID %q is not available", lightID)
		return
	}

	set := settingsFromRaw(light)

	lockFor := cnf.HomeAssistant.DefaultLockFor()

	lightLockForOverride := light.HALockFor()
	if lightLockForOverride != "" {
		lockFor = lightLockForOverride
	}

	lockTime, err := common.ForParser(lockFor)
	if err != nil {
		// Should not happen as we have already validated the default lock time in the MQTT configuration
		logger.Errorf("Invalid default lock time: %v", err)
		return
	}

	switch parts[3] {
	case "command":
		HandleOnOffCommand(cnf, light, set, payload, lockTime)
	case "rgbww":
		HandleRGBWWCommand(cnf, light, set, payload, lockTime)
	case "temp":
		HandleTempCommand(cnf, light, set, payload, lockTime)
	case "brightness":
		HandleBrightnessCommand(cnf, light, set, payload, lockTime)
	case "auto":
		HandleAutomaticControlCommand(cnf, light, set, payload, lockTime)
	default:
		logger.Errorf("Unknown light command: %s", parts[3])
	}
}

func HandleOnOffCommand(cnf *config.Config, light *light.Light, set *wiz.Settings, payload []byte, lockTime time.Time) {
	logger := common.Logger.GetLogger("homeassistant-cmd")
	switch strings.ToLower(string(payload)) {
	case "on":
		if light.On() {
			return
		}

		light.Locker().LockUntil(lockTime, cnf.Colors.GetByValue("W6000"))
	case "off":
		if !light.On() {
			return
		}

		light.Locker().LockUntil(lockTime, cnf.Colors.GetByValue("off"))
	default:
		logger.Errorf("Unknown on/off command: %s", string(payload))
	}
	light.Locker().Kick()
}

func HandleRGBWWCommand(cnf *config.Config, light *light.Light, set *wiz.Settings, payload []byte, lockTime time.Time) {
	logger := common.Logger.GetLogger("homeassistant-cmd")
	var r, g, b, ww, cw int
	n, err := fmt.Sscanf(strings.ReplaceAll(string(payload), " ", ""), "%d,%d,%d,%d,%d", &r, &g, &b, &ww, &cw)
	if err != nil || n != 5 {
		logger.Errorf("Invalid RGBWW command, expected 5 integers in payload but got %d: %s", n, string(payload))
		return
	}
	bright := set.GetBrightness()
	if bright == 0 {
		bright = 100
	}
	col := cnf.Colors.GetByValue(fmt.Sprintf("#%02x%02x%02x%02x%02x/%d", r, g, b, ww, cw, bright))
	light.Locker().LockUntil(lockTime, col)
	light.Locker().Kick()
}

func HandleTempCommand(cnf *config.Config, light *light.Light, set *wiz.Settings, payload []byte, lockTime time.Time) {
	logger := common.Logger.GetLogger("homeassistant-cmd")

	temp, err := strconv.Atoi(string(payload))
	if err != nil {
		logger.Errorf("Invalid temperature command, expected an integer but got: %s", string(payload))
		return
	}
	if temp < 2200 || temp > 6500 {
		logger.Errorf("Invalid temperature command, expected a value between 2200 and 6500 but got: %d", temp)
		return
	}

	col := cnf.Colors.GetByValue(fmt.Sprintf("W%d", temp))
	light.Locker().LockUntil(lockTime, col)
	light.Locker().Kick()
}

func HandleBrightnessCommand(cnf *config.Config, light *light.Light, set *wiz.Settings, payload []byte, lockTime time.Time) {
	// And so a little bit tricky, because brightness is part of the color setting
	logger := common.Logger.GetLogger("homeassistant-cmd")
	bright, err := strconv.Atoi(string(payload))
	if err != nil {
		logger.Errorf("Invalid brightness command, expected an integer but got: %s", string(payload))
		return
	}
	if bright < 1 || bright > 100 {
		logger.Errorf("Invalid brightness command, expected a value between 1 and 100 but got: %d", bright)
		return
	}

	set.Brightness(bright)
	newcol := set.ColorStr()
	col := cnf.Colors.GetByValue(newcol)
	light.Locker().LockUntil(lockTime, col)
	light.Locker().Kick()
}

func HandleAutomaticControlCommand(cnf *config.Config, light *light.Light, set *wiz.Settings, payload []byte, lockTime time.Time) {
	// Implement the handling of lock command for the light
	logger := common.Logger.GetLogger("homeassistant-cmd")
	switch strings.ToLower(string(payload)) {
	case "off":
		logger.Warn("AC OFF, LOCKING LIGHT")
		// Lock up whatever the current color is
		ccol := light.LastActualColor()
		if ccol == nil {
			logger.Errorf("No actual color available to lock")
			return
		}
		light.Locker().LockUntil(lockTime, ccol)
		light.Locker().Kick()
	case "on":
		logger.Warn("AC ON, UNLOCKING LIGHT")
		light.Locker().UnlockNow(false)
	default:
		logger.Errorf("Invalid automatic control command, expected 'ON' or 'OFF' but got: %s", string(payload))
	}
}

func settingsFromRaw(light *light.Light) *wiz.Settings {
	lastRaw := light.GetLastRawColor()
	if lastRaw == nil {
		return wiz.NewSettings().Temperature(6000).Brightness(100).On()
	}
	if !lastRaw.State {
		return wiz.NewSettings().Temperature(6000).Brightness(100).On()
	}
	return wiz.NewSettings().
		RGB(uint8(lastRaw.R), uint8(lastRaw.G), uint8(lastRaw.B)).
		Temperature(lastRaw.Temp).
		Brightness(lastRaw.BR).
		State(lastRaw.State)

	// TODO Need to figure out scene support here

}
