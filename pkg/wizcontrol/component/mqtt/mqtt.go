package mqtt

const DefaultMQTTTopic = "wizcontrol"

type Mqtt struct {
	url            string
	username       string
	password       string
	topic          string
	controlEnabled bool
}

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
	}
}

func (m *Mqtt) GetTopic() string {
	if m.topic == "" {
		return DefaultMQTTTopic
	}
	return m.topic
}
func (m *Mqtt) GetURL() string {
	return m.url
}

func (m *Mqtt) GetUsername() string {
	return m.username
}

func (m *Mqtt) GetPassword() string {
	return m.password
}

func (m *Mqtt) IsControlEnabled() bool {
	return m.controlEnabled
}

func (m *Mqtt) String() string {
	return "MQTT URL: " + m.GetURL() + ", Topic: " + m.GetTopic() + ", Username: " + m.GetUsername() + ", Password: " + m.GetPassword()
}
