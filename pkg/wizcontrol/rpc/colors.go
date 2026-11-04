package rpc

import (
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/colors"
)

type ColorsResponse struct {
	Colors []string
}

func HandleColorsRequest() []byte {
	resp := &ColorsResponse{}

	for name := range colors.Colors {
		resp.Colors = append(resp.Colors, name)
	}
	slices.Sort(resp.Colors)

	jResp, _ := json.Marshal(resp)
	return jResp
}
