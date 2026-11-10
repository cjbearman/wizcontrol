package mqtt

const DefaultMQTTTopic = "wizcontrol"
const DefaultHATopic = "homeassistant"
const DefaultHANodeID = "wzdefault"

type Mqtt struct {
	url                 string
	username            string
	password            string
	topic               string
	controlEnabled      bool
	retainMessages      bool
	insecure            bool
	serverCerts         string
	enableHomeAssistant bool
	homeAssistantTopic  string
	homeAssistantNodeID string
	defaultLockFor      string
}

// Build constructs an Mqtt instance from the provided ConfigYAML, applying default values where necessary.
func Build(config *ConfigYAML) *Mqtt {

	if config == nil {
		return nil
	}

	if config.HomeAssistantNodeID == "" {
		config.HomeAssistantNodeID = DefaultHANodeID
	}

	topic := config.Topic
	if topic == "" {
		topic = DefaultMQTTTopic
	}
	if config.HomeAssistantTopic == "" {
		config.HomeAssistantTopic = DefaultHATopic
	}

	return &Mqtt{
		url:                 config.URL,
		username:            config.Username,
		password:            config.Password,
		topic:               topic,
		controlEnabled:      *config.ControlEnabled,
		retainMessages:      *config.RetainMessages,
		insecure:            *config.Insecure,
		serverCerts:         config.ServerCerts,
		enableHomeAssistant: config.EnableHomeAssistant,
		homeAssistantTopic:  config.HomeAssistantTopic,
		homeAssistantNodeID: config.HomeAssistantNodeID,
		defaultLockFor:      config.DefaultLockFor,
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

// ShouldConnectInsecurely returns whether the MQTT connection should be established insecurely (without verifying the server's TLS certificate).
func (m *Mqtt) ShouldConnectInsecurely() bool {
	return m.insecure
}

// String returns a string representation of the MQTT configuration, including URL, topic, username, and password.
func (m *Mqtt) String() string {
	return "MQTT URL: " + m.GetURL() + ", Topic: " + m.GetTopic() + ", Username: " + m.GetUsername() + ", Password: " + m.GetPassword()
}

func (m *Mqtt) ServerCerts() string {
	return m.serverCerts
}

func (m *Mqtt) EnableHomeAssistant() bool {
	return m.enableHomeAssistant
}

func (m *Mqtt) HomeAssistantTopic() string {
	return m.homeAssistantTopic
}

func (m *Mqtt) HomeAssistantNodeID() string {
	return m.homeAssistantNodeID
}

func (m *Mqtt) GetDefaultLockFor() string {
	return m.defaultLockFor
}
