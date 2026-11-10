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

	IDS := jsonEscapeStr(snake(light.Name()))
	name := jsonEscapeStr(light.Name())
	topic := jsonEscapeStr(fmt.Sprintf("%s/light/%s", cnf.MQTT.GetTopic(), light.MqqtID()))

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
      "json_attributes_topic": #TOPIC#,
      "brightness_state_topic": #TOPIC#,
      "rgbww_state_topic": #TOPIC#,
      "color_temp_state_topic": #TOPIC#,
      "json_attributes_template": "{{ value_json.attributes | to_json }}",
      "state_value_template": "{{ value_json.ha.state }}",
      "color_mode_value_template": "{{ value_json.ha.color_mode }}",
      "brightness_value_template": "{{ value_json.ha.brightness }}",
      "color_temp_value_template": "{{ value_json.ha.color_temp }}",
      "rgbww_value_template": "{{ value_json.ha.rgbww }}",
      "brightness_scale": 100,
      "supported_color_modes": [
        "color_temp",
        "rgbww"
      ],
      "color_temp_kelvin": true,
      "min_kelvin": 2200,
      "max_kelvin": 6500,
      "command_topic": "dross",
      "color_temp_command_topic": "dross2",
      "rgbww_command_topic": "dross3"
    }
  }
}`
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#IDS#", IDS)
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#NAME#", name)
	payloadTemplate = strings.ReplaceAll(payloadTemplate, "#TOPIC#", topic)

	discTopic := fmt.Sprintf("%s/device/%s/config", cnf.MQTT.HomeAssistantTopic(), jsonEscapeStrNoQuotes(snake(light.Name())))

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
