package homeassistant

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

func DiscoveryPayload(cnf *config.Config, light *light.Light) ([]byte, string) {

	if cnf.HomeAssistant == nil {
		common.Logger.GetLogger("homeassistant-disc").Errorf("Home Assistant configuration is not available")
		return nil, ""
	}

	IDS := jsonEscapeStr(snake(light.Name()))
	name := jsonEscapeStr(light.Name())
	topic := jsonEscapeStr(fmt.Sprintf("%s/light/%s", cnf.MQTT.GetTopic(), light.MqqtID()))
	switchUID := jsonEscapeStr(snake(light.Name() + "_lock"))
	cmdBaseTopic := jsonEscapeStrNoQuotes(fmt.Sprintf("%s/light/%s", cnf.MQTT.GetTopic(), light.MqqtID()))

	payloadTemplate := `{
  "dev": {
    "ids": #IDS#,
    "name": #NAME#,
    "mf": "Chris Bearman"
  },
  "o": {
    "name": "Chris Bearman"
  },
  "cmps": {
    "light": {
      "p": "light",
      "unique_id": #IDS#,
      "state_topic": #TOPIC#,
	  "name": "Light",
      "json_attributes_topic": #TOPIC#,
      "brightness_state_topic": #TOPIC#,
      "rgbww_state_topic": #TOPIC#,
      "color_temp_state_topic": #TOPIC#,
	  "availability_topic": #TOPIC#,
	  "availability_template": "{{ value_json.available }}",
      "json_attributes_template": "{{ value_json.attributes | to_json }}",
      "state_value_template": "{{ value_json.ha.state }}",
      "color_mode_value_template": "{{ value_json.ha.color_mode }}",
      "brightness_value_template": "{{ value_json.ha.brightness }}",
      "color_temp_value_template": "{{ value_json.ha.color_temp }}",
      "rgbww_value_template": "{{ value_json.ha.rgbww }}",
      "brightness_scale": 100,
	  "payload_available": true,
	  "payload_not_available": false,
      "supported_color_modes": [
        "color_temp",
        "rgbww"
      ],
      "color_temp_kelvin": true,
      "min_kelvin": 2200,
      "max_kelvin": 6500,
      "command_topic": "#CMD_TOPIC#/command",
      "color_temp_command_topic": "#CMD_TOPIC#/temp",
      "rgbww_command_topic": "#CMD_TOPIC#/rgbww",
      "brightness_command_topic": "#CMD_TOPIC#/brightness"
    },
	"lock": {
      "p": "switch",
      "unique_id": #SWITCH_UID#,
      "command_topic": "#CMD_TOPIC#/auto",
	  "state_topic": #TOPIC#,
	  "name": "Automatic Control",
      "value_template": "{{ not value_json.attributes.locked }}",
	  "state_on": true,
	  "state_off": false
    }
  }
}`

	/*

	 */
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#IDS#", IDS)
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#NAME#", name)
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#TOPIC#", topic)
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#SWITCH_UID#", switchUID)
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#CMD_TOPIC#", cmdBaseTopic)

	discTopic := fmt.Sprintf("%s/device/%s/config", cnf.HomeAssistant.HomeAssistantTopic(), jsonEscapeStrNoQuotes(snake(light.Name())))

	return []byte(payloadTemplate), discTopic
}

func snake(str string) string {
	return strings.ReplaceAll(str, " ", "_")
}

func jsonEscapeStr(str string) string {
	bytes, err := json.Marshal(str)
	if err != nil {
		common.Logger.GetLogger("json-escape").Errorf("Failed to escape JSON string: %v", err)
		return ""
	}

	return string(bytes)
}

func jsonEscapeStrNoQuotes(str string) string {
	bytes, err := json.Marshal(str)
	if err != nil {
		common.Logger.GetLogger("json-escape").Errorf("Failed to escape JSON string: %v", err)
		return ""
	}

	escaped := string(bytes)
	return escaped[1 : len(escaped)-1]
}
