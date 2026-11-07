package mqtt

import (
	"crypto/sha256"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/bus"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	topicLock        = "lock"
	topicUnlock      = "unlock"
	topicAddScene    = "addscene"
	topicDelScene    = "delscene"
	topicReset       = "reset"
	topicScenes      = "scenes"
	topicColors      = "colors"
	topicTargets     = "targets"
	topicSceneGroups = "scenegroups"
)
const BufferSize = 100

type messageAndTopic struct {
	topic   string
	message []byte
}

var Instance *Mqtt

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

func (m *Mqtt) handleLockRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received day RPC message")
	_, err := rpc.HandleLockRequest(m.cnf, msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle day request: %v", err)
	}
}

func (m *Mqtt) handleUnlockRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received unlock RPC message")
	_, err := rpc.HandleUnlockRequest(m.cnf, msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle unlock request: %v", err)
	}
}

func (m *Mqtt) handleAddSceneRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received add scene RPC message")
	_, err := rpc.HandleAddScenesRequest(m.cnf, msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle add scene request: %v", err)
	}
}
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
	}

	// Publish initial scenes, scenegroups, colors and targets
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicScenes), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleScenesRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicSceneGroups), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleSceneGroupsRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicColors), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleColorsRequest(m.cnf)).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicTargets), 1, m.cnf.MQTT.ShouldRetainMessages(), rpc.HandleTargetsRequest(m.cnf)).Wait()
}

func (m *Mqtt) disconnectHandler(client mqtt.Client, err error) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Errorf("MQTT disconnected: %v", err)
}

func (m *Mqtt) run() {
	logger := common.Logger.GetLogger("mqtt")
	defer close(m.Closed)

	messageUniqueness := make(map[string]string)

	logger.Info("mqtt started")

	clientId, err := uuid.NewRandom()
	if err != nil {
		logger.Errorf("failed to generate client UUID: %v", err)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(m.MqttURL).
		SetClientID(clientId.String()).
		SetUsername(m.MqttUsername).
		SetPassword(m.MqttPassword).
		SetAutoReconnect(true).
		SetCleanSession(true)

	opts.OnConnect = m.connectHandler
	opts.OnConnectionLost = m.disconnectHandler
	client := mqtt.NewClient(opts)

	client.Connect()

	logger.Info("MQTT connection initiated")

	for {
		select {
		case <-m.StopSignalled:
			logger.Info("stop signalled")
			client.Disconnect(250)
			logger.Info("mqtt stopped")
			return
		case msg := <-m.outboundMessage:

			// Message has an ID, check for uniqueness
			hash := m.hashMessage(msg.Payload)
			if prevHash, exists := messageUniqueness[msg.Topic]; exists {
				if prevHash == hash {
					logger.Debugf("Duplicate message to %s skipped", msg.Topic)
					continue
				}
			}
			messageUniqueness[msg.Topic] = hash
			topic := fmt.Sprintf("%s/%s", m.MqttTopic, msg.Topic)
			client.Publish(topic, 1, m.cnf.MQTT.ShouldRetainMessages(), msg.Payload)
			logger.Infof("Dispatched an MQTT update to %s", topic)
		}
	}
}

func (m *Mqtt) hashMessage(message []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(message))
}
