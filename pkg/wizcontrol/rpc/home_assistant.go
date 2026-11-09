package rpc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

//      json_attributes_topic: wizcontrol/light/front_flood_cr
//      json_attributes_template: "{{ value_json.attributes | to_json }}"

type HADevice struct {
	IDs          string `json:"ids"`
	Name         string `json:"name"`
	Manufacturer string `json:"mf"`
}

type HAOrigin struct {
	Name string `json:"name"`
}

type HAComponent struct {
	P                      string `json:"p"`
	StateTopic             string `json:"state_topic"`
	ValueTemplate          string `json:"value_template"`
	UniqueID               string `json:"unique_id"`
	JsonAttributesTopic    string `json:"json_attributes_topic"`
	JsonAttributesTemplate string `json:"json_attributes_template"`
	Name                   string `json:"name"`
}

type HAConfig struct {
	Device     HADevice               `json:"dev"`
	Origin     HAOrigin               `json:"o"`
	Components map[string]HAComponent `json:"cmps"`
}

// ColorsResponse represents the response payload for a request to list available colors.

// HandleColorsRequest processes a request to list available colors.
func HandleHomeAssistantConfiguration(cnf *config.Config) []byte {
	resp := HAConfig{
		Device: HADevice{
			IDs:          "wizcontrol",
			Name:         "WizControl RGB Scene Controller",
			Manufacturer: "Chris Bearman",
		},
		Origin: HAOrigin{
			Name: "Chris Bearman",
		},
		Components: make(map[string]HAComponent),
	}

	for _, light := range cnf.Lights {
		uid := Snakeify(light.Name())
		resp.Components[uid] = HAComponent{
			P:                      "sensor",
			StateTopic:             fmt.Sprintf("%s/light/%s", cnf.MQTT.GetTopic(), light.Name()),
			ValueTemplate:          "{{ value_json.available | iif('on', 'off') }}",
			JsonAttributesTopic:    fmt.Sprintf("%s/light/%s", cnf.MQTT.GetTopic(), light.Name()),
			JsonAttributesTemplate: "{{ value_json.attributes | to_json }}",
			UniqueID:               uid,
			Name:                   light.Name(),
		}
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}

func Snakeify(str string) string {
	return strings.ReplaceAll(strings.ToLower(str), " ", "_")
}
