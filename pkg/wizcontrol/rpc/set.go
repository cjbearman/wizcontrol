package rpc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

type LightSetRequest struct {
	Target  string `json:"target"`
	Color   string `json:"color"`
	LockFor int    `json:"for"`
	Force   bool   `json:"force"`
}

type LightRespItem struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}

func HandleSetRequest(request []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-day")
	req := &LightSetRequest{}

	err := json.Unmarshal(request, req)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	l := req.Target

	// l is a light or light group, we must find it
	lightsToUpdate, err := lights.GetLightOrGroup(l)
	if err != nil || len(lightsToUpdate) == 0 {
		return NewGeneralResponse(false, fmt.Sprintf("Light or light group %s not found", l)), NewRpcError(400, "Light or light group %s not found", l)
	}

	// Validate request
	if req.Color == "" {
		return NewGeneralResponse(false, "Color not provided"), NewRpcError(400, "Color not provided")
	}

	var colorName, color string
	ok := false

	// Is it one of our registered colors?
	if definedColor, found := colors.Colors[req.Color]; found {
		colorName = req.Color
		color = definedColor
		ok = true
	} else {
		// If not, maybe it's a raw color value

		if rationalized, err := wiz.RationalizeColor(req.Color); err == nil {
			colorName = "unknown"
			color = rationalized
			ok = true
		}
	}

	if !ok {
		return NewGeneralResponse(false, "Invalid color value, not a config defined color or valid raw color"), NewRpcError(400, "Invalid color value, not a config defined color or valid raw color")
	}

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
		return NewGeneralResponse(false, fmt.Sprintf("Failed to create pilot settings: %v", err)), NewRpcError(500, "Failed to create pilot settings: %v", err)
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

			if LightStatusDispatcher != nil {
				LightStatusDispatcher(refLight)
			}
		}
	}

	resp, _ := json.Marshal(jResp)
	return resp, nil
}
