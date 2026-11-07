package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// HandleScenesRequest processes a request to list available scenes.
type ScenesResponse struct {
	Scenes []string `json:"scenes"`
}

// ScenesResponse represents the response payload for a request to list available scenes.
func HandleScenesRequest(cnf *config.Config) []byte {
	resp := &ScenesResponse{}

	for name := range cnf.Scenes {
		resp.Scenes = append(resp.Scenes, name)
	}
	slices.Sort(resp.Scenes)

	jResp, _ := json.Marshal(resp)
	return jResp
}
