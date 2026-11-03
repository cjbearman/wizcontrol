package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/mqtt"
)

type LightSetRequest struct {
	Color   string `json:"color"`
	LockFor int    `json:"for"`
	Force   bool   `json:"force"`
}

type LightRespItem struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}

func (s *ServerInstance) setHandler(w http.ResponseWriter, r *http.Request) {
	l := r.PathValue("light")
	logger := common.Logger.GetLogger("server").WithField(FieldRemoteAddr, r.RemoteAddr).WithField(FieldLightOrGroup, l).WithField(FieldOperation, "set")

	// l is a light or light group, we must find it
	lightsToUpdate, err := lights.GetLightOrGroup(l)
	if err != nil || len(lightsToUpdate) == 0 {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Light or light group %s not found", l))
		logger.Warnf("Requested light or light group not found")
		return
	}

	// Decodce request
	rawBody, err := io.ReadAll(r.Body)
	logger.Debugf("Raw request body:\n%s", string(rawBody))

	req := &LightSetRequest{}
	if err := json.Unmarshal(rawBody, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		logger.Warnf("Failed to decode request body: %v", err)
		return
	}

	// Validate request
	if req.Color == "" {
		WriteError(w, http.StatusBadRequest, "Color not provided")
		logger.Warnf("Color not provided")
		return
	}

	var colorName, color string
	ok := false

	// Is it one of our registered colors?
	if definedColor, found := colors.Colors[req.Color]; found {
		colorName = req.Color
		color = definedColor
		ok = true
	} else {
		if rationalized, err := wiz.RationalizeColor(req.Color); err == nil {
			colorName = "unknown"
			color = rationalized
			ok = true
		}
	}

	// If not, maybe it's a raw color value
	if !ok {
		WriteError(w, http.StatusBadRequest, "Invalid color value, not a config defined color or valid raw color")
		logger.Warnf("Invalid color: %s", req.Color)
		return
	}

	logger = logger.WithField(FieldColor, color).WithField(FieldColorName, colorName)

	// JSON response to be built
	jResp := make(map[string]LightRespItem)

	// Create a wiz device group with all affected lights that are available
	group := wiz.NewDeviceGroup()
	for _, l := range lightsToUpdate {
		if !l.Available() {
			jResp[l.Device.Name()] = LightRespItem{
				Success: false,
				Reason:  "Light not available",
			}
			continue
		}
		group.AddDevice(l.Device)
	}
	pilot, err := wiz.SettingsFromString(color)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create pilot settings: %v", err))
		logger.Errorf("Failed to create pilot settings: %v", err)
		return
	}

	pilot.On()
	for _, result := range group.SetPilot(pilot).AwaitAll() {
		refLight, found := lights.Lights[result.Device.Name()]
		if !found {
			logger.Warnf("Result for unknown light: %s", result.Device.Name())
			continue
		}
		if result.Err != nil {
			jResp[refLight.Device.Name()] = LightRespItem{
				Success: false,
				Reason:  result.Err.Error(),
			}
		} else {
			jResp[refLight.Device.Name()] = LightRespItem{
				Success: true,
			}
			refLight.SetLastColor(color)
			refLight.SetLastProgrammedColor(colorName, color)
			if req.LockFor > 0 {
				refLight.Locker.LockFor(time.Duration(req.LockFor) * time.Second)
			}
			if mqtt.MqttInstance != nil {
				mqtt.MqttInstance.SendLightStatus(refLight)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jResp)
	logger.Infof("Web set color operation completed succesfully: %s", color)
}
