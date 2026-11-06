package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
)

type DelScenesRequest struct {
	Scenes []string `json:"scenes"`
	Groups []string `json:"groups"`
}

func HandleDelScenesRequest(cnf *config.Config, req []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-del-scenes")
	request := &DelScenesRequest{}

	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	for _, scene := range request.Scenes {
		if cnf.Scenes.GetByName(scene) == nil {
			return NewGeneralResponse(false, fmt.Sprintf("scene %q not found", scene)), NewRpcError(404, "scene %q not found", scene)
		}
		_ = cnf.AdditionalScenes.DeleteAdditionalScene(scene)
	}

	for _, group := range request.Groups {

		if scenes, found := cnf.Scenes.GetByGroup(group); !found {
			return NewGeneralResponse(false, fmt.Sprintf("scene group %q not found", group)), NewRpcError(404, "scene group %q not found", group)
		} else {
			for _, scene := range scenes {
				logger.Infof("Additional scene %q removed", scene)
				_ = cnf.AdditionalScenes.DeleteAdditionalScene(scene.GetName())
			}
		}
	}

	daemon.Instance.Kick()
	return NewGeneralResponse(true, ""), nil
}
