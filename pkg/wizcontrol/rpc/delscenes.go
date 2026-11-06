package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

type DelScenesRequest struct {
	Scenes []string `json:"scenes"`
	Groups []string `json:"groups"`
}

func HandleDelScenesRequest(req []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-del-scenes")
	request := &DelScenesRequest{}

	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	for _, scene := range request.Scenes {
		if _, found := scenes.GetScenes()[scene]; !found {
			return NewGeneralResponse(false, fmt.Sprintf("scene %q not found", scene)), NewRpcError(404, "scene %q not found", scene)
		}
		_ = scenes.DeleteAdditionalScene(scene)
	}

	for _, group := range request.Groups {
		if members, found := scenes.GetSceneGroup(group); !found {
			return NewGeneralResponse(false, fmt.Sprintf("scene group %q not found", group)), NewRpcError(404, "scene group %q not found", group)
		} else {
			for _, scene := range members {
				logger.Infof("Additional scene %q removed", scene)
				_ = scenes.DeleteAdditionalScene(scene)
			}
		}
	}

	daemon.Instance.Kick()
	return NewGeneralResponse(true, ""), nil
}
