// Package mqtt provides MQTT integration for the wizcontrol system, handling both publishing and subscribing to relevant topics.
package mqtt

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/bus"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

// List of topics used for MQTT communication
// these are all housed underneath the primary topic and thus are not considered
// fully qualified names
const (
	topicLock        = "lock"
	topicUnlock      = "unlock"
	topicAddScene    = "addscene"
	topicDelScene    = "delscene"
	topicReset       = "reset"
	topicPause       = "pause"
	topicUnpause     = "unpause"
	topicScenes      = "scenes"
	topicColors      = "colors"
	topicTargets     = "targets"
	topicSceneGroups = "scenegroups"
	topicLightGroups = "lightgroups"
	topicCalendars   = "calendars"
	topicHolidays    = "holidays"
)

// Instance holds the singleton instance of the MQTT integration
var Instance *Mqtt

// Mqtt represents the MQTT integration, managing connection, subscriptions, and message handling.
type Mqtt struct {
	*common.BaseRunner
	outboundMessage      <-chan bus.Message
	lastDaemonStatusHash string
	lastLightStatusHash  map[string]string
	MqttURL              string
	MqttUsername         string
	MqttPassword         string
	MqttTopic            string
	ControlEnabled       bool
	cnf                  *config.Config
}

// NewMqtt creates a new instance of the MQTT integration with the provided configuration.
func NewMqtt(cnf *config.Config) *Mqtt {
	mqtt := Mqtt{
		BaseRunner:           common.NewRunner("mqtt"),
		outboundMessage:      cnf.Bus.Channel(),
		lastLightStatusHash:  make(map[string]string),
		lastDaemonStatusHash: "",
		MqttURL:              cnf.MQTT.GetURL(),
		MqttUsername:         cnf.MQTT.GetUsername(),
		MqttPassword:         cnf.MQTT.GetPassword(),
		MqttTopic:            cnf.MQTT.GetTopic(),
		ControlEnabled:       cnf.MQTT.IsControlEnabled(),
		cnf:                  cnf,
	}
	mqtt.Exec = mqtt.run

	return &mqtt

}

// handleLockRPC handles inbound "lock" messages, routing them to the
// relevant RPC handler.
func (m *Mqtt) handleLockRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received day RPC message")
	_, err := rpc.HandleLockRequest(m.cnf, msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle day request: %v", err)
	}
}

// handleUnlockRPC handles inbound "unlock" messages, routing them to the relevant RPC handler.
func (m *Mqtt) handleUnlockRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received unlock RPC message")
	_, err := rpc.HandleUnlockRequest(m.cnf, msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle unlock request: %v", err)
	}
}

// handleAddSceneRPC handles inbound "add scene" messages, routing them to the relevant RPC handler.
func (m *Mqtt) handleAddSceneRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received add scene RPC message")
	_, err := rpc.HandleAddScenesRequest(m.cnf, msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle add scene request: %v", err)
	}
}

// handleDelSceneRPC handles inbound "delete scene" messages, routing them to the relevant RPC handler.
func (m *Mqtt) handleDelSceneRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received delete scene RPC message")
	_, err := rpc.HandleDelScenesRequest(m.cnf, msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle delete scene request: %v", err)
	}
}

func (m *Mqtt) handleResetRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received reset RPC message")
	_ = rpc.HandleResetRequest(m.cnf)
}

// handlePauseRPC handles inbound "pause" messages, routing them to the relevant RPC handler.
func (m *Mqtt) handlePauseRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received pause RPC message")
	_ = rpc.HandlePauseRequest(m.cnf)
}

// handleUnpauseRPC handles inbound "unpause" messages, routing them to the relevant RPC handler.
func (m *Mqtt) handleUnpauseRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received unpause RPC message")
	_ = rpc.HandleUnpauseRequest(m.cnf)
}

// connectHandler is called when the MQTT client successfully connects to the broker. It subscribes to control topics and publishes initial state.
func (m *Mqtt) connectHandler(client mqtt.Client) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("MQTT connected")

	// Subscribe to control topics
	if m.ControlEnabled {
		client.Subscribe(fmt.Sprintf("%s/control/%s", m.MqttTopic, topicLock), 1, m.handleLockRPC)
		client.Subscribe(fmt.Sprintf("%s/control/%s", m.MqttTopic, topicUnlock), 1, m.handleUnlockRPC)
		client.Subscribe(fmt.Sprintf("%s/control/%s", m.MqttTopic, topicAddScene), 1, m.handleAddSceneRPC)
		client.Subscribe(fmt.Sprintf("%s/control/%s", m.MqttTopic, topicDelScene), 1, m.handleDelSceneRPC)
		client.Subscribe(fmt.Sprintf("%s/control/%s", m.MqttTopic, topicReset), 1, m.handleResetRPC)
		client.Subscribe(fmt.Sprintf("%s/control/%s", m.MqttTopic, topicPause), 1, m.handlePauseRPC)
		client.Subscribe(fmt.Sprintf("%s/control/%s", m.MqttTopic, topicUnpause), 1, m.handleUnpauseRPC)
	}

	// Publish initial scenes, scenegroups, lightgroups, colors and targets
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicScenes), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleScenesRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicSceneGroups), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleSceneGroupsRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicLightGroups), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleLightGroupsRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicColors), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleColorsRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicTargets), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleTargetsRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicCalendars), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleCalendarsRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicHolidays), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleHolidaysRequest(m.cnf)).Wait()
}

// disconnectHandler is called when the MQTT client loses connection to the broker.
// it logs the disconnection event and the associated error.
func (m *Mqtt) disconnectHandler(client mqtt.Client, err error) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Errorf("MQTT disconnected: %v", err)
}

// run is the main loop for the MQTT integration, handling connection, message dispatch, and stop signals.
func (m *Mqtt) run() {
	logger := common.Logger.GetLogger("mqtt")
	defer close(m.Closed)

	// messageUniqueness is used to store hashes of outbound
	// messages to prevent duplicate dispatches.
	// this is applied to any messages dispatched via the bus, which are daemon and light messages
	// all other sent messages are just sent on connection and therefore exempt from uniqueness checks.
	messageUniqueness := make(map[string]string)

	logger.Info("MQTT started")

	clientId, err := uuid.NewRandom()
	if err != nil {
		// bloody unlikely, but handle it anyway.
		panic(fmt.Sprintf("failed to generate client UUID: %v", err))
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: m.cnf.MQTT.ShouldConnectInsecurely(),
	}

	if m.cnf.MQTT.ServerCerts() != "" {
		certPool := x509.NewCertPool()
		caCert, err := os.ReadFile(m.cnf.MQTT.ServerCerts())
		if err != nil {
			logger.Errorf("Failed to read CA certificate: %v", err)
			return
		}
		if ok := certPool.AppendCertsFromPEM(caCert); !ok {
			logger.Errorf("Failed to parse CA certificate")
			return
		}
		tlsConfig.RootCAs = certPool
	}

	// Set up the MQTT broker
	opts := mqtt.NewClientOptions().
		AddBroker(m.MqttURL).
		SetClientID(clientId.String()).
		SetUsername(m.MqttUsername).
		SetPassword(m.MqttPassword).
		SetAutoReconnect(true).
		SetCleanSession(true).
		SetTLSConfig(tlsConfig)

	opts.OnConnect = m.connectHandler
	opts.OnConnectionLost = m.disconnectHandler
	client := mqtt.NewClient(opts)

	// Initiate connection
	client.Connect()
	logger.Info("MQTT connection initiated")

	// Main loop for handling stop signals and outbound messages
	for {
		select {
		case <-m.StopSignalled:
			logger.Info("stop signalled")
			client.Disconnect(250)
			logger.Info("mqtt stopped")
			return
		case msg := <-m.outboundMessage:

			// Before dispatching this message, make sure it's different from the
			// last message dispatched to the topic
			hash := m.hashMessage(msg.Payload)
			if prevHash, exists := messageUniqueness[msg.Topic]; exists {
				if prevHash == hash {
					logger.Debugf("Duplicate message to %s skipped", msg.Topic)
					continue
				}
			}
			messageUniqueness[msg.Topic] = hash

			// Dispatch the message
			topic := fmt.Sprintf("%s/%s", m.MqttTopic, msg.Topic)
			client.Publish(topic, 1, m.cnf.MQTT.ShouldRetainMessages(), msg.Payload)
			logger.Infof("Dispatched an MQTT update to %s", topic)
		}
	}
}

// hashMessage computes a SHA-256 hash of the given message.
func (m *Mqtt) hashMessage(message []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(message))
}
