package rpc

import (
	"encoding/json"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// SceneInfo represents information about a scene, including its name and whether it is currently selected.
type SceneInfo struct {
	Name              string `json:"name"`
	CurrentlySelected bool   `json:"currently_selected"`
}

// ScenesResponse represents the response payload for a request to list available scenes.
func HandleScenesRequest(cnf *config.Config) []byte {
	resp := []SceneInfo{}

	currentScenes, _ := cnf.Calendars.GetCalendarScenes(cnf.Scenes)
	isCurrentlySelected := func(name string) bool {
		for _, item := range currentScenes {
			if item.Name() == name {
				return true
			}
		}
		return false
	}

	for name := range cnf.Scenes {
		resp = append(resp, SceneInfo{
			Name:              name,
			CurrentlySelected: isCurrentlySelected(name),
		})
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}
