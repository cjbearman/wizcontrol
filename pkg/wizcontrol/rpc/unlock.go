package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

type UnlockRequest struct {
	Target string `json:"target"`
}

func HandleUnlockRequest(req []byte) ([]byte, *RpcError) {

	logger := common.Logger.GetLogger("rpc-unlock")

	request := &UnlockRequest{}
	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	affectedLights, err := lights.GetLightOrGroup(request.Target)
	if err != nil || len(affectedLights) == 0 {
		logger.Errorf("Light or group %q not found", request.Target)
		return NewGeneralResponse(false, "Light or group not found"), NewRpcError(404, "Light or group not found")
	}

	for _, light := range affectedLights {
		light.Locker.UnlockNow(false)
	}

	return NewGeneralResponse(true, "Unlocked successfully"), nil
}
