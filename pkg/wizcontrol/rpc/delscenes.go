package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/daemon"
)

// DelScenesRequest represents the request payload for deleting scenes, including the scenes and groups.
// HandleDelScenesRequest processes a DelScenesRequest, updating the configuration by removing the specified scenes and groups.
// 'cnf' is the current configuration, and 'req' is the raw JSON request payload.
// Returns a JSON-encoded response and an optional RpcError if the request is invalid.
type DelScenesRequest struct {
	Scenes []string `json:"scenes"`
	Groups []string `json:"groups"`
}

// HandleDelScenesRequest handles the deletion of scenes specified in the DelScenesRequest.
func HandleDelScenesRequest(cnf *config.Config, req []byte) ([]byte, *RpcError) {
	logger := common.Logger.GetLogger("rpc-del-scenes")
	request := &DelScenesRequest{}

	err := json.Unmarshal(req, request)
	if err != nil {
		return NewGeneralResponse(false, fmt.Sprintf("invalid request: %v", err)), NewRpcError(400, "invalid request: %v", err)
	}

	for _, scene := range request.Scenes {
		sc := cnf.Scenes[scene]
		if sc == nil {
			return NewGeneralResponse(false, fmt.Sprintf("scene %q not found", scene)), NewRpcError(404, "scene %q not found", scene)
		}
		_ = cnf.AdditionalScenes.DeleteAdditionalScene(sc)
	}

	for _, group := range request.Groups {

		if scenes, found := cnf.SceneGroups[group]; !found {
			return NewGeneralResponse(false, fmt.Sprintf("scene group %q not found", group)), NewRpcError(404, "scene group %q not found", group)
		} else {
			for _, scene := range scenes {
				logger.Infof("Additional scene %q removed", scene.Name())
				_ = cnf.AdditionalScenes.DeleteAdditionalScene(scene)
			}
		}
	}

	daemon.Instance.Kick()
	return NewGeneralResponse(true, ""), nil
}
