package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// TargetsResponse represents the response structure for the targets request.
type TargetsResponse struct {
	LightGroups []string `json:"light_groups"`
	Lights      []string `json:"lights"`
}

// HandleTargetsRequest handles the targets request and returns the JSON-encoded response.
func HandleTargetsRequest(cnf *config.Config) []byte {
	resp := &TargetsResponse{}

	for name := range cnf.LightGroups {
		resp.LightGroups = append(resp.LightGroups, name)
	}
	for name := range cnf.Lights {
		resp.Lights = append(resp.Lights, name)
	}
	slices.Sort(resp.LightGroups)
	slices.Sort(resp.Lights)

	jResp, _ := json.Marshal(resp)
	return jResp
}
