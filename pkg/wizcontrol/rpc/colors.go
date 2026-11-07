package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// ColorsResponse represents the response payload for a request to list available colors.
type ColorsResponse struct {
	Colors []string `json:"colors"`
}

// HandleColorsRequest processes a request to list available colors.
func HandleColorsRequest(cnf *config.Config) []byte {
	resp := &ColorsResponse{}

	for name := range cnf.Colors {
		resp.Colors = append(resp.Colors, name)
	}
	slices.Sort(resp.Colors)

	jResp, _ := json.Marshal(resp)
	return jResp
}
