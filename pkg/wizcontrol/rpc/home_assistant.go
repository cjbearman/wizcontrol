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
	P                       string   `json:"p"`
	StateTopic              string   `json:"state_topic"`
	ValueTemplate           string   `json:"value_template"`
	UniqueID                string   `json:"unique_id"`
	JsonAttributesTopic     string   `json:"json_attributes_topic"`
	JsonAttributesTemplate  string   `json:"json_attributes_template"`
	BrightnessScale         int      `json:"brightness_scale"`
	StateValueTemplate      string   `json:"state_value_template"`
	ColorModeTemplate       string   `json:"color_mode_value_template"`
	BrightnessStateTopic    string   `json:"brightness_state_topic"`
	BrightnessValueTemplate string   `json:"brightness_value_template"`
	ColorTempStateTopic     string   `json:"color_temp_state_topic"`
	ColorTempValueTemplate  string   `json:"color_temp_value_template"`
	RGBWWStateTopic         string   `json:"rgbww_state_topic"`
	RGBWWValueTemplate      string   `json:"rgbww_value_template"`
	Name                    string   `json:"name"`
	CommandTopic            string   `json:"command_topic"`
	ColorTempCommandTopic   string   `json:"color_temp_command_topic"`
	RGBWWCommandTopic       string   `json:"rgbww_command_topic"`
	SupportedColorModes     []string `json:"supported_color_modes"`
	ColorTempKelvin         bool     `json:"color_temp_kelvin"`
	MinKelvin               int      `json:"min_kelvin"`
	MaxKelvin               int      `json:"max_kelvin"`
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
		haTopic := fmt.Sprintf("%s/light/%s", cnf.MQTT.GetTopic(), light.Name())
		uid := Snakeify(light.Name())
		resp.Components[uid] = HAComponent{
			P:          "light",
			StateTopic: haTopic,
			// ValueTemplate:           "{{ value_json.available | iif('ON', 'OFF') }}",
			ValueTemplate:           "{{ value_json.ha.available  }}",
			StateValueTemplate:      "{{ value_json.ha.state }}",
			JsonAttributesTopic:     haTopic,
			JsonAttributesTemplate:  "{{ value_json.attributes | to_json }}",
			BrightnessScale:         100,
			SupportedColorModes:     []string{"color_temp", "rgbww"},
			ColorModeTemplate:       "{{ value_json.ha.color_mode }}",
			BrightnessStateTopic:    haTopic,
			ColorTempStateTopic:     haTopic,
			RGBWWStateTopic:         haTopic,
			BrightnessValueTemplate: "{{ value_json.ha.brightness }}",
			ColorTempValueTemplate:  "{{ value_json.ha.color_temp }}",
			RGBWWValueTemplate:      "{{ value_json.ha.rgbww }}",
			UniqueID:                uid,
			Name:                    light.Name(),
			CommandTopic:            "dross",
			ColorTempCommandTopic:   "dross2",
			RGBWWCommandTopic:       "dross3",
			ColorTempKelvin:         true,
			MinKelvin:               2200,
			MaxKelvin:               6500,
		}
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}

func Snakeify(str string) string {
	return strings.ReplaceAll(strings.ToLower(str), " ", "_")
}
