package homeassistant

const DefaultMQTTTopic = "wizcontrol"
const DefaultHATopic = "homeassistant"
const DefaultHANodeID = "wzdefault"

type HomeAssistant struct {
	homeAssistantTopic string
	defaultLockFor     string
}

// Build constructs a HomeAssistant instance from the provided ConfigYAML, applying default values where necessary.
func Build(config *ConfigYAML) *HomeAssistant {

	if config == nil {
		return nil
	}

	if config.HomeAssistantTopic == "" {
		config.HomeAssistantTopic = DefaultHATopic
	}

	return &HomeAssistant{
		homeAssistantTopic: config.HomeAssistantTopic,
		defaultLockFor:     config.DefaultLockFor,
	}
}

func (h *HomeAssistant) HomeAssistantTopic() string {
	return h.homeAssistantTopic
}

func (h *HomeAssistant) DefaultLockFor() string {
	return h.defaultLockFor
}
