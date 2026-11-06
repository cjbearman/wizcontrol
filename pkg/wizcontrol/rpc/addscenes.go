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
	Groups        []string `json:"groups"`
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

	scenesToAdd := []string{}

	for _, scene := range request.Scenes {
		if _, found := scenes.GetScenes()[scene]; !found {
			return NewGeneralResponse(false, fmt.Sprintf("scene %q not found", scene)), NewRpcError(404, "scene %q not found", scene)
		}
		scenesToAdd = append(scenesToAdd, scene)
	}

	for _, group := range request.Groups {
		if scenes, found := scenes.GetSceneGroup(group); !found {
			return NewGeneralResponse(false, fmt.Sprintf("scene group %q not found", group)), NewRpcError(404, "scene group %q not found", group)
		} else {
			scenesToAdd = append(scenesToAdd, scenes...)
		}
	}

	if len(scenesToAdd) <= 0 {
		return NewGeneralResponse(false, "no valid scenes to add"), NewRpcError(400, "no valid scenes to add")
	}

	for _, scene := range scenesToAdd {
		logger.Infof("Additional scene %q added until %v", scene, until)
		_ = scenes.AddAdditionalScene(scene, until)
	}

	daemon.Instance.Kick()
	return NewGeneralResponse(true, ""), nil
}
