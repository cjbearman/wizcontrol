package rpc

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
)

// LockRequest represents the request payload for a lock operation.
type LockRequest struct {
	Targets common.StringSlice `json:"targets"`
	For     string             `json:"for"`
	Until   string             `json:"until"`
	Color   common.StringSlice `json:"color"`
}

// HandleLockRequest processes a lock request, validating the input and locking the specified targets.
func HandleLockRequest(cnf *config.Config, req []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-lock")
	request := &LockRequest{}

	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	// Validate color
	if len(request.Color) == 0 {
		return NewGeneralResponse(false, "color is required"), NewRpcError(400, "color is required")
	}

	var colors []string

	for _, color := range request.Color {
		col := cnf.Colors[color]
		if col != nil {
			colors = append(colors, col.Value())
		} else {
			// So it wasn't a named color, check if it's a raw color spec
			if _, err := wiz.SettingsFromString(color); err == nil {
				colors = append(colors, color)
			} else {
				return NewGeneralResponse(false, fmt.Sprintf("invalid color %q, must be named color or raw color specification", color)), NewRpcError(400, "invalid color: %s", color)
			}
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

	// If we have more than one target, we'll shuffle the array of targets
	if len(targetLights) > 1 {
		rand.Shuffle(len(targetLights), func(i, j int) {
			targetLights[i], targetLights[j] = targetLights[j], targetLights[i]
		})
	}

	// If we have more than one color, we'll shuffle the array of colors
	if len(colors) > 1 {
		rand.Shuffle(len(colors), func(i, j int) {
			colors[i], colors[j] = colors[j], colors[i]
		})
	}

	// Lock the targets, assigning a color to each based on its index in the shuffled array of targets
	for idx, target := range targetLights {
		color := colors[idx%len(colors)]
		colorObj := cnf.Colors.GetByValue(color)

		target.Locker().LockUntil(lockedUntil, colorObj)
		logger.Infof("Locked %q until %v with color %s", target.ID(), lockedUntil, color)
	}

	daemon.Instance.Kick()

	return NewGeneralResponse(true, ""), nil
}
