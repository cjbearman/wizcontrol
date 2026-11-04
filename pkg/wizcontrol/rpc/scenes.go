package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/scenes"
)

type ScenesResponse struct {
	Scenes []string
}

func HandleScenesRequest() []byte {
	resp := &ScenesResponse{}

	for name := range scenes.Scenes {
		resp.Scenes = append(resp.Scenes, name)
	}
	slices.Sort(resp.Scenes)

	jResp, _ := json.Marshal(resp)
	return jResp
}
