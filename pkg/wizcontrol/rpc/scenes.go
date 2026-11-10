package rpc

import (
	"encoding/json"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

type SceneElement struct {
	Targets []string `json:"targets"`
	Color   string   `json:"color"`
}

// SceneInfo represents information about a scene, including its name and whether it is currently selected.
type SceneInfo struct {
	Name          string         `json:"name"`
	SceneElements []SceneElement `json:"elements"`
}

// ScenesResponse represents the response payload for a request to list available scenes.
func HandleScenesRequest(cnf *config.Config) []byte {
	resp := []SceneInfo{}

	for name := range cnf.Scenes {

		els := []SceneElement{}
		for _, e := range cnf.Scenes[name].GetElements() {
			targets := []string{}
			for _, t := range e.GetTargets() {
				targets = append(targets, t.ID())
			}
			els = append(els, SceneElement{
				Targets: e.GetRawTargets(),
				Color:   e.GetColor().Name(),
			})
		}

		resp = append(resp, SceneInfo{
			Name:          name,
			SceneElements: els,
		})
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}
