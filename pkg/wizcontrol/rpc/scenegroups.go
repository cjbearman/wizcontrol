package rpc

import (
	"encoding/json"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// HandleSceneGroupsRequest processes a request to list available scene groups.
type ScreenGroupsResponse struct {
	SceneGroups map[string][]string `json:"scene_groups"`
}

// ScreenGroupsResponse represents the response payload for a request to list available scene groups.
func HandleSceneGroupsRequest(cnf *config.Config) []byte {
	resp := &ScreenGroupsResponse{SceneGroups: make(map[string][]string)}

	for name, scenes := range cnf.SceneGroups {

		for _, scene := range scenes {
			resp.SceneGroups[name] = append(resp.SceneGroups[name], scene.Name())
		}
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}
