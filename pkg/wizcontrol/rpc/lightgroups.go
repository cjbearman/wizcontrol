package rpc

import (
	"encoding/json"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// LightGroupsResponse represents the response payload for a request to list available light groups.
type LightGroupsResponse struct {
	LightGroups map[string][]string `json:"light_groups"`
}

// HandleLightGroupsRequest processes a request to list available light groups.
func HandleLightGroupsRequest(cnf *config.Config) []byte {
	resp := &LightGroupsResponse{LightGroups: make(map[string][]string)}

	for name, lg := range cnf.LightGroups {

		for _, light := range lg.Lights() {
			resp.LightGroups[name] = append(resp.LightGroups[name], light.Name())
		}
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}
