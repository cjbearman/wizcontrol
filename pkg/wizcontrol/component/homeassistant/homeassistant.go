package homeassistant

type HomeAssistant struct {
	homeAssistantTopic string
	defaultLockFor     string
	alwaysAvailable    bool
}

// Build constructs a HomeAssistant instance from the provided ConfigYAML, applying default values where necessary.
func Build(config *ConfigYAML) *HomeAssistant {

	if config == nil {
		return nil
	}

	return &HomeAssistant{
		homeAssistantTopic: config.HomeAssistantTopic,
		defaultLockFor:     config.DefaultLockFor,
		alwaysAvailable:    config.AlwaysAvailable,
	}
}

func (h *HomeAssistant) HomeAssistantTopic() string {
	return h.homeAssistantTopic
}

func (h *HomeAssistant) DefaultLockFor() string {
	return h.defaultLockFor
}

func (h *HomeAssistant) AlwaysAvailable() bool {
	return h.alwaysAvailable
}
