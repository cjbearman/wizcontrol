package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

type SceneElement struct {
	Targets []string `json:"targets"`
	Color   string   `json:"color"`
}

// SceneInfo represents information about a scene, including its name and whether it is currently selected.
type SceneInfo struct {
	Name              string         `json:"name"`
	CurrentlySelected bool           `json:"currently_selected"`
	SceneElements     []SceneElement `json:"elements"`
}

// ScenesResponse represents the response payload for a request to list available scenes.
func HandleScenesRequest(cnf *config.Config) []byte {
	resp := []SceneInfo{}

	currentScenes, _ := cnf.Calendars.GetCalendarScenes(cnf.Scenes)
	additionalScenes := cnf.AdditionalScenes.GetAdditionalSceneNames()
	isCurrentlySelected := func(name string) bool {
		for _, item := range currentScenes {
			if item.Name() == name {
				return true
			}
		}
		return slices.Contains(additionalScenes, name)
	}

	for name := range cnf.Scenes {

		els := []SceneElement{}
		for _, e := range cnf.Scenes[name].GetElements() {
			targets := []string{}
			for _, t := range e.GetTargets() {
				targets = append(targets, t.Name())
			}
			els = append(els, SceneElement{
				Targets: e.GetRawTargets(),
				Color:   e.GetColor().Name(),
			})
		}

		resp = append(resp, SceneInfo{
			Name:              name,
			CurrentlySelected: isCurrentlySelected(name),
			SceneElements:     els,
		})
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}
