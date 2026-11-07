package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

type UnlockRequest struct {
	Targets []string `json:"targets"`
}

func HandleUnlockRequest(cnf *config.Config, req []byte) ([]byte, *RpcError) {

	logger := common.Logger.GetLogger("rpc-unlock")

	request := &UnlockRequest{}
	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	if len(request.Targets) == 0 {
		return NewGeneralResponse(false, "no targets specified"), NewRpcError(400, "no targets specified")
	}

	affectedLights := []*light.Light{}

	for _, item := range request.Targets {
		lights := cnf.GetLightOrGroup(item)

		if len(lights) == 0 {
			logger.Errorf("Light or group %q not found", item)
			return NewGeneralResponse(false, fmt.Sprintf("Light or group %q not found", item)), NewRpcError(400, "Light or group not found")
		}
		affectedLights = append(affectedLights, lights...)
	}

	for _, light := range affectedLights {
		light.Locker().UnlockNow(false)
	}

	return NewGeneralResponse(true, "Unlocked successfully"), nil
}
