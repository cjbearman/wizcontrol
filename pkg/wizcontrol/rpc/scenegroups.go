package rpc

import (
	"encoding/json"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

type ScreenGroupsResponse struct {
	SceneGroups map[string][]string `json:"scene_groups"`
}

func HandleSceneGroupsRequest(cnf *config.Config) []byte {
	resp := &ScreenGroupsResponse{SceneGroups: make(map[string][]string)}

	for name, scenes := range cnf.SceneGroups {

		for _, scene := range scenes {
			resp.SceneGroups[name] = append(resp.SceneGroups[name], scene.GetName())
		}
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}
