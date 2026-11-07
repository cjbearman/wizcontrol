package mqtt

const DefaultMQTTTopic = "wizcontrol"

type Mqtt struct {
	url            string
	username       string
	password       string
	topic          string
	controlEnabled bool
	retainMessages bool
}

// Build constructs an Mqtt instance from the provided ConfigYAML, applying default values where necessary.
func Build(config *ConfigYAML) *Mqtt {

	if config == nil {
		return nil
	}

	topic := config.Topic
	if topic == "" {
		topic = DefaultMQTTTopic
	}

	return &Mqtt{
		url:            config.URL,
		username:       config.Username,
		password:       config.Password,
		topic:          topic,
		controlEnabled: *config.ControlEnabled,
		retainMessages: *config.RetainMessages,
	}
}

// GetTopic returns the MQTT topic, defaulting to DefaultMQTTTopic if not set.
func (m *Mqtt) GetTopic() string {
	if m.topic == "" {
		return DefaultMQTTTopic
	}
	return m.topic
}

// GetURL returns the MQTT URL.
func (m *Mqtt) GetURL() string {
	return m.url
}

// GetUsername returns the MQTT username.
func (m *Mqtt) GetUsername() string {
	return m.username
}

// GetPassword returns the MQTT password.
func (m *Mqtt) GetPassword() string {
	return m.password
}

// IsControlEnabled returns whether control is enabled for the MQTT component.
func (m *Mqtt) IsControlEnabled() bool {
	return m.controlEnabled
}

// ShouldRetainMessages returns whether MQTT messages should be retained.
func (m *Mqtt) ShouldRetainMessages() bool {
	return m.retainMessages
}

// String returns a string representation of the MQTT configuration, including URL, topic, username, and password.
func (m *Mqtt) String() string {
	return "MQTT URL: " + m.GetURL() + ", Topic: " + m.GetTopic() + ", Username: " + m.GetUsername() + ", Password: " + m.GetPassword()
}
