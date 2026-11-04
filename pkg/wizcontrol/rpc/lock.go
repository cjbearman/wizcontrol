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

	// Validate duration
	if request.For == "" {
		return NewGeneralResponse(false, "for is required"), NewRpcError(400, "for is required")
	}

	duration, err := time.ParseDuration(request.For)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid 'for' duration: %v", err)), NewRpcError(400, "invalid for duration: %v", err)
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
		target.Locker.LockFor(duration, request.Color)
		logger.Infof("Locked %q for %v with color %s", target.Device.Name(), request.For, request.Color)
	}

	daemon.Daemon.Kick()

	return NewGeneralResponse(true, ""), nil
}
