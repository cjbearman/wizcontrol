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
	P        string `json:"p"`
	UniqueID string `json:"unique_id"`

	// All our topics are in fact the same, we just extract different bits from it using the value
	StateTopic           string `json:"state_topic"`            // Main state published here
	JsonAttributesTopic  string `json:"json_attributes_topic"`  // Attributes published here
	BrightnessStateTopic string `json:"brightness_state_topic"` // Brightness state published here
	RGBWWStateTopic      string `json:"rgbww_state_topic"`      // RGBWW state published here
	ColorTempStateTopic  string `json:"color_temp_state_topic"` // Color temperature state published here
	Name                 string `json:"name"`

	// ValueTemplate           string   `json:"value_template"`		// Main value for the light, believe incorrect, removed

	// Various templates that extract values from the payload
	JsonAttributesTemplate  string `json:"json_attributes_template"`  // Template for extracting attributes from the state payload
	StateValueTemplate      string `json:"state_value_template"`      // Template for extracting the main state value
	ColorModeTemplate       string `json:"color_mode_value_template"` // Template for extracting the color mode
	BrightnessValueTemplate string `json:"brightness_value_template"` // Template for extracting the brightness value
	ColorTempValueTemplate  string `json:"color_temp_value_template"` // Template for extracting the color temperature value
	RGBWWValueTemplate      string `json:"rgbww_value_template"`      // Template for extracting the RGBWW value

	// These set various parameters for the integration
	BrightnessScale     int      `json:"brightness_scale"` // Scale for brightness values
	SupportedColorModes []string `json:"supported_color_modes"`
	ColorTempKelvin     bool     `json:"color_temp_kelvin"`
	MinKelvin           int      `json:"min_kelvin"`
	MaxKelvin           int      `json:"max_kelvin"`

	// Command topics are where change commands are sent
	CommandTopic          string `json:"command_topic"`
	ColorTempCommandTopic string `json:"color_temp_command_topic"`
	RGBWWCommandTopic     string `json:"rgbww_command_topic"`
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
			P:        "light",
			UniqueID: uid,
			Name:     light.Name(),

			StateTopic:           haTopic,
			JsonAttributesTopic:  haTopic,
			BrightnessStateTopic: haTopic,
			ColorTempStateTopic:  haTopic,
			RGBWWStateTopic:      haTopic,

			StateValueTemplate:      "{{ value_json.ha.state }}",
			JsonAttributesTemplate:  "{{ value_json.attributes | to_json }}",
			ColorModeTemplate:       "{{ value_json.ha.color_mode }}",
			BrightnessValueTemplate: "{{ value_json.ha.brightness }}",
			ColorTempValueTemplate:  "{{ value_json.ha.color_temp }}",
			RGBWWValueTemplate:      "{{ value_json.ha.rgbww }}",

			CommandTopic:          "dross",
			ColorTempCommandTopic: "dross2",
			RGBWWCommandTopic:     "dross3",

			SupportedColorModes: []string{"color_temp", "rgbww"},
			BrightnessScale:     100,
			ColorTempKelvin:     true,
			MinKelvin:           2200,
			MaxKelvin:           6500,
		}
	}

	jResp, _ := json.Marshal(resp)
	return jResp
}

func Snakeify(str string) string {
	return strings.ReplaceAll(strings.ToLower(str), " ", "_")
}
