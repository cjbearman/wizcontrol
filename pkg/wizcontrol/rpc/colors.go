package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

type ColorsResponse struct {
	Colors []string `json:"colors"`
}

func HandleColorsRequest(cnf *config.Config) []byte {
	resp := &ColorsResponse{}

	for name := range cnf.Colors {
		resp.Colors = append(resp.Colors, name)
	}
	slices.Sort(resp.Colors)

	jResp, _ := json.Marshal(resp)
	return jResp
}
