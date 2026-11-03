package mqtt

import (
	"encoding/json"
	"errors"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/responses"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const BufferSize = 100

type messageAndTopic struct {
	topic   string
	message []byte
}

type Mqtt struct {
	*common.Runner
	outboundMessage chan messageAndTopic
}

var MqttInstance *Mqtt

var MqttURL string
var MqttUsername string
var MqttPassword string
var MqttTopic string

func NewMqtt() *Mqtt {
	mqtt := Mqtt{
		Runner:          common.NewRunner(),
		outboundMessage: make(chan messageAndTopic, BufferSize),
	}
	mqtt.Exec = mqtt.run
	return &mqtt

}

func (l *Mqtt) connectHandler(client mqtt.Client) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Info("MQTT connected")
}

func (l *Mqtt) disconnectHandler(client mqtt.Client, err error) {
	logger := common.Logger.GetLogger("mqtt")
	logger.Errorf("MQTT disconnected: %v", err)
}

func (m *Mqtt) run() {
	logger := common.Logger.GetLogger("mqtt")
	defer close(m.Runner.Closed)

	logger.Info("mqtt started")

	clientId, err := uuid.NewRandom()
	if err != nil {
		logger.Errorf("failed to generate client UUID: %v", err)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(MqttURL).
		SetClientID(clientId.String()).
		SetUsername(MqttUsername).
		SetPassword(MqttPassword).
		SetAutoReconnect(true).
		SetCleanSession(true)

	opts.OnConnect = m.connectHandler
	opts.OnConnectionLost = m.disconnectHandler
	client := mqtt.NewClient(opts)

	client.Connect()

	logger.Info("MQTT connection initiated")

	for {
		select {
		case <-m.Runner.StopSignalled:
			logger.Info("stop signalled")
			return
		case msg := <-m.outboundMessage:
			client.Publish(msg.topic, 1, true, msg.message)
			logger.Debugf("Send a message to %s", msg.topic)
		}
	}

	// And then disconnect
	client.Disconnect(250)
	logger.Info("mqtt stopped")
}

func (m *Mqtt) SendLightStatus(light *lights.Light) error {
	response := responses.BuildLightStatus(light)

	responseJson, err := json.Marshal(response)
	if err != nil {
		common.Logger.GetLogger("mqtt").Errorf("failed to marshal light status: %v", err)
		return err
	}
	mqttMessage := messageAndTopic{
		topic:   MqttTopic + "/light/" + light.Device.Name(),
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

	responseJson, err := json.Marshal(status)
	if err != nil {
		common.Logger.GetLogger("mqtt").Errorf("failed to marshal daemon status: %v", err)
		return err
	}
	mqttMessage := messageAndTopic{
		topic:   MqttTopic + "/server",
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
