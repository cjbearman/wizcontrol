package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

type TargetsResponse struct {
	LightGroups []string `json:"light_groups"`
	Lights      []string `json:"lights"`
}

func HandleTargetsRequest() []byte {
	resp := &TargetsResponse{}

	for name := range lights.GetLightGroups() {
		resp.LightGroups = append(resp.LightGroups, name)
	}
	for name := range lights.GetLights() {
		resp.Lights = append(resp.Lights, name)
	}
	slices.Sort(resp.LightGroups)
	slices.Sort(resp.Lights)

	jResp, _ := json.Marshal(resp)
	return jResp
}
