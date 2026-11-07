package rpc

import (
	"cmp"
	"encoding/json"
	"slices"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

type Color struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ColorsResponse represents the response payload for a request to list available colors.

// HandleColorsRequest processes a request to list available colors.
func HandleColorsRequest(cnf *config.Config) []byte {
	resp := []Color{}

	for name, col := range cnf.Colors {
		resp = append(resp, Color{
			Name:  name,
			Value: col.Value(),
		})
	}
	slices.SortFunc(resp, func(a, b Color) int {
		return cmp.Compare(a.Name, b.Name)
	})

	jResp, _ := json.Marshal(resp)
	return jResp
}
