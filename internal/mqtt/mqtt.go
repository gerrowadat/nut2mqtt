package mqtt

import (
	"fmt"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
	control "github.com/gerrowadat/nut2mqtt/internal/control"
)

type mqttClient struct {
	c          mqtt.Client
	topic_base string
}

func NewMQTTClient(mqtt_url string, user *string, pass *string) (mqttClient, error) {

	ret := mqttClient{}

	log.Print("Connecting to MQTT at " + mqtt_url + " as " + *user)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqtt_url)
	opts.SetClientID("nut2mqtt")
	opts.SetUsername(*user)
	opts.SetPassword(*pass)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return ret, error(token.Error())
	}

	ret.c = client

	return ret, nil
}

func (m *mqttClient) SetTopicBase(topic_base string) {
	m.topic_base = topic_base
}

func (m *mqttClient) GetTopicBase() string {
	return m.topic_base
}

func (m *mqttClient) PublishMessage(msg *channels.MQTTUpdate) error {
	pub_tok := m.c.Publish(m.topic_base+msg.Topic, 0, false, msg.Content)
	pub_tok.Wait()

	if pub_tok.Error() != nil {
		return error(pub_tok.Error())
	}
	return nil
}

func (m *mqttClient) Disconnect(code uint) {
	m.c.Disconnect(code)
}

func TopicFromUPSVariableUpdate(up *channels.UPSVariableUpdate) string {
	ret := fmt.Sprintf("hosts/%v/%v/%v", up.Host, up.UpsName, up.VarName)
	return strings.Replace(ret, ".", "/", -1)
}

func (m *mqttClient) UpdateProducer(c *control.Controller) {
	// Take in UPSVariableUpdate messages and spit out MQTTUpdate messages to be consumed.
	defer c.WaitGroupDone()
	for {
		up := <-c.Channels().MqttConverter
		topic := TopicFromUPSVariableUpdate(up)
		c.Channels().Mqtt <- &channels.MQTTUpdate{Topic: topic, Content: up.Content, OldContent: up.OldContent}
	}
}

func (m *mqttClient) UpdateConsumer(c *control.Controller) {
	defer c.WaitGroupDone()
	for {
		update := <-c.Channels().Mqtt
		old := update.OldContent
		if old == "" {
			old = "[null]"
		}
		log.Printf("MQTT Change: [%v]\t%v -> %v ", m.GetTopicBase()+update.Topic, old, update.Content)
		c.MetricRegistry().Metrics().MQTTUpdatesProcessed.Inc()
		m.PublishMessage(update)
	}
}
