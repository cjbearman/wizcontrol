package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
)

type LockRequest struct {
	Targets []string `json:"targets"`
	For     string   `json:"for"`
	Until   string   `json:"until"`
	Color   string   `json:"color"`
}

type LockResponse struct {
}

func HandleLockRequest(cnf *config.Config, req []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-lock")
	request := &LockRequest{}

	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	// Validate color
	if request.Color == "" {
		return NewGeneralResponse(false, "color is required"), NewRpcError(400, "color is required")
	}

	var color string
	col := cnf.Colors[request.Color]
	if col != nil {
		color = col.Value()
	} else {
		// So it wasn't a named color, check if it's a raw color spec
		if _, err := wiz.SettingsFromString(request.Color); err == nil {
			color = request.Color
		} else {
			return NewGeneralResponse(false, fmt.Sprintf("invalid color %q, must be named color or raw color specification", request.Color)), NewRpcError(400, "invalid color: %s", request.Color)
		}
	}

	if request.For == "" && request.Until == "" {
		logger.Warn("Cannot handle request with both for and until set")
		return NewGeneralResponse(false, "Provide 'for' or 'until' but not both'"), NewRpcError(400, "either 'for' or 'until' is required")
	}

	lockedUntil, err2 := validateForUntil(request.For, request.Until)
	if err2 != nil {
		return NewGeneralResponse(false, err2.Error()), err2

	}

	// Validate targets
	if len(request.Targets) == 0 {
		return NewGeneralResponse(false, "at least one target is required"), NewRpcError(400, "at least one target is required")
	}

	var targetLights []*light.Light

	for _, target := range request.Targets {
		lightsInTarget := cnf.GetLightOrGroup(target)
		if len(lightsInTarget) == 0 {
			return NewGeneralResponse(false, fmt.Sprintf("'%q' is not a valid target (not a light or light group)'", target)), NewRpcError(400, "invalid target: %s", target)
		} else {
			targetLights = append(targetLights, lightsInTarget...)
		}

	}
	// Lock the targets
	for _, target := range targetLights {
		colorObj := cnf.Colors.GetByValue(color)

		target.Locker().LockUntil(lockedUntil, colorObj)
		logger.Infof("Locked %q until %v with color %s", target.Name(), lockedUntil, color)
	}

	daemon.Instance.Kick()

	return NewGeneralResponse(true, ""), nil
}
