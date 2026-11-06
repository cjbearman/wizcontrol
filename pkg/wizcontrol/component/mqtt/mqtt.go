package mqtt

const DefaultMQTTTopic = "wizcontrol"

type Mqtt struct {
	URL      string
	Username string
	Password string
	Topic    string
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
		URL:      config.URL,
		Username: config.Username,
		Password: config.Password,
		Topic:    topic,
	}
}

func (m *Mqtt) GetTopic() string {
	if m.Topic == "" {
		return DefaultMQTTTopic
	}
	return m.Topic
}
func (m *Mqtt) GetURL() string {
	return m.URL
}

func (m *Mqtt) GetUsername() string {
	return m.Username
}

func (m *Mqtt) GetPassword() string {
	return m.Password
}

func (m *Mqtt) String() string {
	return "MQTT URL: " + m.GetURL() + ", Topic: " + m.GetTopic() + ", Username: " + m.GetUsername() + ", Password: " + m.GetPassword()
}
