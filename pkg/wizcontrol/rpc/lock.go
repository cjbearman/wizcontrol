package rpc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

type LockRequest struct {
	Targets []string `json:"target"`
	For     string   `json:"for"`
	Until   string   `json:"until"`
	Color   string   `json:"color"`
}

type LockResponse struct {
}

func HandleLockRequest(req []byte) ([]byte, *RpcError) {
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

	if _, found := colors.Colors[request.Color]; !found {
		return NewGeneralResponse(false, fmt.Sprintf("invalid color: %s", request.Color)), NewRpcError(400, "invalid color: %s", request.Color)
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

	var targetLights []*lights.Light

	for _, target := range request.Targets {
		if lightsInTarget, err := lights.GetLightOrGroup(target); err != nil {
			return NewGeneralResponse(false, fmt.Sprintf("'%q' is not a valid target (not a light or light group)'", target)), NewRpcError(400, "invalid target: %s", target)
		} else {
			targetLights = append(targetLights, lightsInTarget...)
		}

	}
	// Lock the targets
	for _, target := range targetLights {
		target.Locker.LockUntil(lockedUntil, request.Color)
		logger.Infof("Locked %q until %v with color %s", target.Device.Name(), lockedUntil, request.Color)
	}

	daemon.Daemon.Kick()

	return NewGeneralResponse(true, ""), nil
}

func validateForUntil(forspec, until string) (time.Time, *RpcError) {
	if forspec == "" && until == "" {
		return time.Time{}, NewRpcError(400, "either 'for' or 'until' is required")
	}
	if forspec != "" && until != "" {
		return time.Time{}, NewRpcError(400, "'for' and 'until' are mutually exclusive")
	}
	if forspec != "" {
		duration, err := time.ParseDuration(forspec)
		if err != nil {
			return time.Time{}, NewRpcError(400, "invalid 'for' duration: %v", err)
		}
		return time.Now().In(common.Location).Add(duration), nil
	}
	if until != "" {
		// TODO: Support additional time formats
		lockUntil, err := time.ParseInLocation("2006-01-02 15:04", until, common.Location)
		if err != nil {
			return time.Time{}, NewRpcError(400, "invalid 'until' time: %v", err)
		}
		return lockUntil, nil
	}
	return time.Time{}, NewRpcError(400, "either 'for' or 'until' is required")
}
