package mqtt

import (
	"context"

	pmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
)

type MqttClient struct {
	Broker       string
	User         string
	Password     string
	CleanSession bool
	Clientid     string

	client  pmqtt.Client
	options *pmqtt.ClientOptions
	enabled bool
}

func NewMqttClient(cfg interface{}) (*MqttClient, error) {
	m := &MqttClient{}
	if cfg == nil {
		// Bypass mode - mqtt disabled
		return m, nil
	}

	// Map configuration into structure
	err := mapstructure.Decode(cfg, m)
	if err != nil {
		return nil, err
	}

	m.enabled = true
	m.options = pmqtt.NewClientOptions()
	m.options.AddBroker(m.Broker)
	m.options.SetClientID(m.Clientid)
	m.options.SetUsername(m.User)
	m.options.SetPassword(m.Password)
	m.options.SetCleanSession(m.CleanSession)
	m.options.SetDefaultPublishHandler(m.onMessage)

	return m, err
}

func (m *MqttClient) Run(ctx context.Context) error {
	if !m.enabled {
		// Bypass mode - just wait for context close
		glog.Info("MQTT is not enabled")
		<-ctx.Done()
		return nil
	}

	m.client = pmqtt.NewClient(m.options)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	glog.Infof("Connected to MQTT broker at %s", m.Broker)

	// Wait until termianted
	<-ctx.Done()
	m.client.Disconnect(1)

	return nil
}

func (m *MqttClient) Publish(topic string, payload []byte, qos byte, retained bool) error {
	if token := b.client.Publish(topic, qos, retained, value); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// PublishRetain is simplified version of Publish intended to use by sensors
func (m *MqttClient) PublishRetain(topic string, payload []byte) error {
	return m.Publish(topic, payload, 0, true)
}

func (m *MqttClient) onMessage(client pmqtt.Client, pmsg pmqtt.Message) {
	glog.Infof("MQTT: %v %v", pmsg.Topic(), pmsg.Payload())
}
