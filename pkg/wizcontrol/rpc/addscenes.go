package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

type AddScenesRequest struct {
	Scenes        []string `json:"scenes"`
	ClearExisting bool     `json:"clear_existing"`
	For           string   `json:"for"`
	Until         string   `json:"until"`
}

func HandleAddScenesRequest(req []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-add-scenes")
	request := &AddScenesRequest{}

	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	until, err2 := validateForUntil(request.For, request.Until)
	if err2 != nil {
		return NewGeneralResponse(false, err2.Error()), err2
	}
	if request.ClearExisting {
		logger.Infof("Clearing all additional scenes")
		scenes.ClearAdditionalScenes()
	}
	for _, scene := range request.Scenes {
		if _, found := scenes.GetScenes()[scene]; !found {
			return NewGeneralResponse(false, fmt.Sprintf("scene %q not found", scene)), NewRpcError(404, "scene %q not found", scene)
		}
	}

	for _, scene := range request.Scenes {
		logger.Infof("Additional scene %q added until %v", scene, until)
		_ = scenes.AddAdditionalScene(scene, until)
	}
	daemon.Instance.Kick()
	return NewGeneralResponse(true, ""), nil
}
