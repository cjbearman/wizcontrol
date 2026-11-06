package mqtt

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/light"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/rpc"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/senders"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	topicLock     = "lock"
	topicUnlock   = "unlock"
	topicAddScene = "addscene"
	topicDelScene = "delscene"
	topicReset    = "reset"
	topicScenes   = "scenes"
	topicColors   = "colors"
	topicTargets  = "targets"
)
const BufferSize = 100

type messageAndTopic struct {
	topic   string
	message []byte
}

var Instance *Mqtt

type Mqtt struct {
	*common.BaseRunner
	outboundMessage      chan messageAndTopic
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
		outboundMessage:      make(chan messageAndTopic, BufferSize),
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

	senders.DaemonStatus = mqtt.SendDaemonStatus
	senders.LightStatus = mqtt.SendLightStatus

	return &mqtt

}

func (m *Mqtt) handleLockRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received day RPC message")
	_, err := rpc.HandleLockRequest(msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle day request: %v", err)
	}
}

func (m *Mqtt) handleUnlockRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received unlock RPC message")
	_, err := rpc.HandleUnlockRequest(msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle unlock request: %v", err)
	}
}

func (m *Mqtt) handleAddSceneRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received add scene RPC message")
	_, err := rpc.HandleAddScenesRequest(msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle add scene request: %v", err)
	}
}
func (m *Mqtt) handleDelSceneRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received delete scene RPC message")
	_, err := rpc.HandleDelScenesRequest(msg.Payload())
	if err != nil {
		logger.Errorf("Failed to handle delete scene request: %v", err)
	}
}

func (m *Mqtt) handleResetRPC(client mqtt.Client, msg mqtt.Message) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("Received reset RPC message")
	_ = rpc.HandleResetRequest()
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

	// Publish initial scenes and colors, retained
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicScenes), 1, true, rpc.HandleScenesRequest()).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicColors), 1, true, rpc.HandleColorsRequest()).Wait()
	client.Publish(fmt.Sprintf("%s/%s", m.MqttTopic, topicTargets), 1, true, rpc.HandleTargetsRequest()).Wait()

}

func (m *Mqtt) disconnectHandler(client mqtt.Client, err error) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Errorf("MQTT disconnected: %v", err)
}

func (m *Mqtt) run() {
	logger := common.Logger.GetLogger("mqtt")
	defer close(m.Closed)

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
			client.Publish(msg.topic, 1, true, msg.message)
			logger.Debugf("Send a message to %s", msg.topic)
		}
	}
}

func (m *Mqtt) SendLightStatus(light *light.Light) error {
	logger := common.Logger.GetLogger("mqtt")
	response := responses.BuildLightStatus(light)

	responseJson, err := json.Marshal(response)
	if err != nil {
		common.Logger.GetLogger("mqtt").Errorf("failed to marshal light status: %v", err)
		return err
	}

	hash := m.hashMessage(responseJson)
	if m.lastLightStatusHash[light.GetName()] == hash {
		logger.Debugf("No need to send light status for %s, no change", light.GetName())
		return nil
	}
	logger.Debugf("Sending light status for %s, has changed", light.GetName())
	m.lastLightStatusHash[light.GetName()] = hash

	mqttMessage := messageAndTopic{
		topic:   m.MqttTopic + "/light/" + light.GetName(),
		message: responseJson,
	}
	select {
	case m.outboundMessage <- mqttMessage:
		// Message sent successfully
	default:
		common.Logger.GetLogger("mqtt").Errorf("failed to send light status: outbound message channel full")
		return errors.New("outbound message channel full")
	}

	return nil
}

func (m *Mqtt) SendDaemonStatus(status responses.DaemonStatus) error {
	logger := common.Logger.GetLogger("mqtt")
	logger.Debugf("Sending daemon status")

	responseJson, err := json.Marshal(status)
	if err != nil {
		common.Logger.GetLogger("mqtt").Errorf("failed to marshal daemon status: %v", err)
		return err
	}
	hash := m.hashMessage(responseJson)
	if m.lastDaemonStatusHash == hash {
		logger.Debugf("No need to send daemon status, no change")
		return nil
	}
	logger.Debugf("Sending daemon status, has changed")
	m.lastDaemonStatusHash = hash

	mqttMessage := messageAndTopic{
		topic:   m.MqttTopic + "/server",
		message: responseJson,
	}
	select {
	case m.outboundMessage <- mqttMessage:
		// Message sent successfully
	default:
		common.Logger.GetLogger("mqtt").Errorf("failed to send daemon status: outbound message channel full")
		return errors.New("outbound message channel full")
	}

	return nil
}

func (m *Mqtt) hashMessage(message []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(message))
}
